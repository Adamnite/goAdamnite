package bargossip

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"net"
	"sync"
	"time"

	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/crypto/ecies"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/sha3"
)

type peerTransportImpl struct {
	conn             net.Conn
	remotePeerPubKey *ecdsa.PublicKey
	rwmu             sync.RWMutex
	state            *handshakeState

	wbuf bytes.Buffer
}

type handshakeKeys struct {
	AES          []byte
	MAC          []byte
	remotePubKey *ecdsa.PublicKey

	IngressMAC, EgressMAC hash.Hash
}

type handshakeMsg struct {
	Version      uint
	Signature    [65]byte
	SenderPubKey [64]byte
	Nonce        [32]byte
}

type respHandshakeMsg struct {
	Version       uint
	OneTimePubKey [64]byte
	Nonce         [32]byte
}

type handshakeEncKeys struct {
	initiator           bool
	remotePubKey        *ecies.PublicKey
	initNonce           []byte
	respNonce           []byte
	oneTimePrivKey      *ecies.PrivateKey
	remoteOneTimePubKey *ecies.PublicKey
}

type handshakeState struct {
	enc cipher.Stream
	dec cipher.Stream

	macCipher  cipher.Block
	egressMAC  hash.Hash
	ingressMAC hash.Hash
}

type messageDecoder interface {
	Decode([]byte)
}

func (msg *handshakeMsg) Decode(input []byte) {
	n := copy(msg.Nonce[:], input)
	n += 32
	n += copy(msg.SenderPubKey[:], input[n:])
	copy(msg.Signature[:], input[n:])
	msg.Version = AdamniteTCPHandshakeVersion
}

func (msg *respHandshakeMsg) Decode(input []byte) {
	n := copy(msg.OneTimePubKey[:], input)
	copy(msg.Nonce[:], input[n:])
	msg.Version = AdamniteTCPHandshakeVersion
}

func NewPeerTransport(conn net.Conn, remotePeerPubKey *ecdsa.PublicKey) peerTransport {
	return &peerTransportImpl{conn: conn, remotePeerPubKey: remotePeerPubKey}
}

func (t *peerTransportImpl) close(err error) {
	t.rwmu.Lock()
	defer t.rwmu.Unlock()

	if t.conn != nil {

	}
	t.conn.Close()
}

// ********************************************************************************************
// ************************** Adamnite Transport Interface Functions **************************
// ********************************************************************************************

func (t *peerTransportImpl) doHandshake(prvKey *ecdsa.PrivateKey) (*ecdsa.PublicKey, error) {
	t.conn.SetDeadline(time.Now().Add(handshakeTimeout))

	var hKeys handshakeKeys
	var err error
	if t.remotePeerPubKey != nil {
		hKeys, err = startHandshake(t.conn, prvKey, t.remotePeerPubKey)
	} else {
		hKeys, err = receiveHandshake(t.conn, prvKey)
	}

	if err != nil {
		return nil, err
	}

	t.InitWithHandshakeKeys(hKeys)
	return hKeys.remotePubKey, err
}

func (t *peerTransportImpl) doExchangeProtocol() {

}

func (t *peerTransportImpl) ReadMsg() (Msg, error) {
	t.rwmu.RLock()
	defer t.rwmu.RUnlock()

	var msg Msg
	t.conn.SetReadDeadline(time.Now().Add(messageReadTimeout))
	return msg, nil
}

func (t *peerTransportImpl) WriteMsg(msg Msg) error {
	t.rwmu.Lock()
	defer t.rwmu.Unlock()

	t.wbuf.Reset()
	if _, err := io.CopyN(&t.wbuf, msg.Payload, int64(msg.Size)); err != nil {
		return err
	}

	if err := t.conn.SetWriteDeadline(time.Now().Add(messageWriteTimeout)); err != nil {
		return err
	}

	if t.state == nil {
		panic("cannot write message before handshake")
	}
	return nil
}

func (t *peerTransportImpl) InitWithHandshakeKeys(hKeys handshakeKeys) {
	if t.state != nil {
		panic("cannot handshake twice")
	}
	mac, err := aes.NewCipher(hKeys.MAC)
	if err != nil {
		panic("invalid MAC secret: " + err.Error())
	}

	enc, err := aes.NewCipher(hKeys.AES)
	if err != nil {
		panic("invalid AES secret: " + err.Error())
	}

	iv := make([]byte, enc.BlockSize())
	t.state = &handshakeState{
		enc:        cipher.NewCTR(enc, iv),
		dec:        cipher.NewCTR(enc, iv),
		macCipher:  mac,
		egressMAC:  hKeys.EgressMAC,
		ingressMAC: hKeys.IngressMAC,
	}
}

// ********************************************************************************************
// ************************** Adamnite Transport internal Functions ***************************
// ********************************************************************************************

// startHandshake performs the handshake. This must be called on dial side.
func startHandshake(conn io.ReadWriter, prvKey *ecdsa.PrivateKey, remotePeerPubKey *ecdsa.PublicKey) (handshakeKeys, error) {
	// 1. create handshakeMsg
	encKeys := &handshakeEncKeys{remotePubKey: ecies.ImportECDSAPublic(remotePeerPubKey)}
	encKeys.initNonce = make([]byte, 32)
	_, err := rand.Read(encKeys.initNonce)
	if err != nil {
		return handshakeKeys{}, err
	}

	encKeys.oneTimePrivKey, err = ecies.GenerateKey(rand.Reader, crypto.S256(), nil)
	if err != nil {
		return handshakeKeys{}, err
	}

	sk, err := ecies.ImportECDSA(prvKey).GenerateShared(ecies.ImportECDSAPublic(remotePeerPubKey), 16, 16) // generate secret key to use in both side to encrypt
	if err != nil {
		return handshakeKeys{}, err
	}

	signed := utils.XOR(sk, encKeys.initNonce)
	signature, err := crypto.Sign(signed, encKeys.oneTimePrivKey.ExportECDSA())
	if err != nil {
		return handshakeKeys{}, err
	}

	hMsg := new(handshakeMsg)
	copy(hMsg.Nonce[:], encKeys.initNonce)
	copy(hMsg.SenderPubKey[:], crypto.FromECDSAPub(&prvKey.PublicKey)[:])
	copy(hMsg.Signature[:], signature)
	hMsg.Version = AdamniteTCPHandshakeVersion

	// 2. Create a packet from handshake message
	packedMsg, err := msgpack.Marshal(hMsg)
	if err != nil {
		return handshakeKeys{}, err
	}
	enc, err := ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(remotePeerPubKey), packedMsg, nil, nil)
	if err != nil {
		return handshakeKeys{}, err
	}

	// 3. Send packet to remote node
	if _, err = conn.Write(enc); err != nil {
		return handshakeKeys{}, err
	}

	// 4. Receive handshake response message
	handshakeRespMsg := new(respHandshakeMsg)
	handshakeRespPacket, err := readHandshakeMsg(conn, prvKey, handshakeRespMsg)
	if err != nil {
		return handshakeKeys{}, err
	}

	encKeys.respNonce = handshakeRespMsg.Nonce[:]
	encKeys.remoteOneTimePubKey, err = importPublicKey(handshakeRespMsg.OneTimePubKey[:])
	if err != nil {
		return handshakeKeys{}, err
	}

	// 5. Create handshake keys
	return getHandshakeKeys(enc, handshakeRespPacket, encKeys)
}

// receiveHandshake performs the handshake. This must be called on listening side
func receiveHandshake(conn io.ReadWriter, prvKey *ecdsa.PrivateKey) (handshakeKeys, error) {
	hMsg := new(handshakeMsg)
	handshakePacket, err := readHandshakeMsg(conn, prvKey, hMsg)
	if err != nil {
		return handshakeKeys{}, err
	}

	encKeys := new(handshakeEncKeys)
	if err := encKeys.handleHandshakeMsg(hMsg, prvKey); err != nil {
		return handshakeKeys{}, err
	}

	respHMsg, err := encKeys.makeRespHandshakeMsg()
	if err != nil {
		return handshakeKeys{}, err
	}

	// 2. Create packet
	respBytes, err := msgpack.Marshal(respHMsg)
	if err != nil {
		return handshakeKeys{}, err
	}

	enc, err := ecies.Encrypt(rand.Reader, encKeys.remotePubKey, respBytes, nil, nil)
	if err != nil {
		return handshakeKeys{}, err
	}

	// 3. Send packet
	if _, err = conn.Write(enc); err != nil {
		return handshakeKeys{}, err
	}

	return getHandshakeKeys(handshakePacket, enc, encKeys)
}

// readHandshakeMsg decode the message from remote node
func readHandshakeMsg(conn io.Reader, prv *ecdsa.PrivateKey, msg messageDecoder) ([]byte, error) {
	plainSize := 0
	switch msg.(type) {
	case *handshakeMsg:
		plainSize = binary.Size(handshakeMsg{}) + 65 /*pubkey*/ + 16 /*IV*/ + 32 /*MAC*/
	case *respHandshakeMsg:
		plainSize = binary.Size(respHandshakeMsg{}) + 65 /*pubkey*/ + 16 /*IV*/ + 32 /*MAC*/
	}
	buf := make([]byte, plainSize)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	key := ecies.ImportECDSA(prv)
	if dec, err := key.Decrypt(buf, nil, nil); err == nil {
		msg.Decode(dec)
		return buf, nil
	} else {
		return nil, err
	}
}

func getHandshakeKeys(handshakePacket, respHandshakePacket []byte, encKeys *handshakeEncKeys) (handshakeKeys, error) {
	secretKey, err := encKeys.oneTimePrivKey.GenerateShared(encKeys.remoteOneTimePubKey, 16, 16)
	if err != nil {
		return handshakeKeys{}, err
	}

	sharedSecret := crypto.Keccak256(secretKey, crypto.Keccak256(encKeys.respNonce, encKeys.initNonce))
	aesSecret := crypto.Keccak256(secretKey, sharedSecret)

	hKeys := handshakeKeys{
		remotePubKey: encKeys.remotePubKey.ExportECDSA(),
		AES:          aesSecret,
		MAC:          crypto.Keccak256(secretKey, aesSecret),
	}

	mac1 := sha3.NewLegacyKeccak256()
	mac1.Write(utils.XOR(hKeys.MAC, encKeys.respNonce))
	mac1.Write(handshakePacket)
	mac2 := sha3.NewLegacyKeccak256()
	mac2.Write(utils.XOR(hKeys.MAC, encKeys.initNonce))
	mac2.Write(respHandshakePacket)

	if encKeys.initiator {
		hKeys.EgressMAC, hKeys.IngressMAC = mac1, mac2
	} else {
		hKeys.EgressMAC, hKeys.IngressMAC = mac2, mac1
	}
	return hKeys, nil
}

func (encKeys *handshakeEncKeys) handleHandshakeMsg(msg *handshakeMsg, prv *ecdsa.PrivateKey) error {
	remotePubKey, err := importPublicKey(msg.SenderPubKey[:])
	if err != nil {
		return err
	}

	encKeys.initNonce = msg.Nonce[:]
	encKeys.remotePubKey = remotePubKey

	if encKeys.oneTimePrivKey == nil {
		encKeys.oneTimePrivKey, err = ecies.GenerateKey(rand.Reader, crypto.S256(), nil)
		if err != nil {
			return err
		}
	}

	secretKey, err := ecies.ImportECDSA(prv).GenerateShared(encKeys.remotePubKey, 16, 16)
	if err != nil {
		return err
	}

	signedMsg := utils.XOR(secretKey, encKeys.initNonce)
	remoteOntimePubKey, err := crypto.Recover(signedMsg, msg.Signature[:])
	if err != nil {
		return err
	}
	encKeys.remoteOneTimePubKey, _ = importPublicKey(remoteOntimePubKey)
	return nil
}

func (encKeys *handshakeEncKeys) makeRespHandshakeMsg() (msg *respHandshakeMsg, err error) {
	encKeys.respNonce = make([]byte, 32)
	if _, err = rand.Read(encKeys.respNonce); err != nil {
		return nil, err
	}

	msg = new(respHandshakeMsg)
	copy(msg.Nonce[:], encKeys.respNonce)
	copy(msg.OneTimePubKey[:], exportPubkey(&encKeys.oneTimePrivKey.PublicKey))
	msg.Version = AdamniteTCPHandshakeVersion

	return msg, nil
}

// importPublicKey unmarshals 512 bit public keys.
func importPublicKey(pubKey []byte) (*ecies.PublicKey, error) {
	var pubKey65 []byte
	switch len(pubKey) {
	case 64:
		// add 'uncompressed key' flag
		pubKey65 = append([]byte{0x04}, pubKey...)
	case 65:
		pubKey65 = pubKey
	default:
		return nil, fmt.Errorf("invalid public key length %v (expect 64/65)", len(pubKey))
	}
	// TODO: fewer pointless conversions
	pub, err := crypto.UnmarshalPubkey(pubKey65)
	if err != nil {
		return nil, err
	}
	return ecies.ImportECDSAPublic(pub), nil
}

// exportPubkey marshals 512 bit public keys.
func exportPubkey(pub *ecies.PublicKey) []byte {
	if pub == nil {
		panic("nil pubkey")
	}
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y)[1:]
}
