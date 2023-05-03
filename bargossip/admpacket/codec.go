package admpacket

import (
	"bytes"
	"crypto/ecdsa"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/common/mclock"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/vmihailenco/msgpack/v5"
)

// SSLCodec encodes and decodes the Adamnite UDP packets.
type SSLCodec struct {
	localnode    *admnode.LocalNode
	sessionCache *SessionCache
	sha256       hash.Hash
	privKey      *ecdsa.PrivateKey

	// encode buffers
	buf      bytes.Buffer // whole packet
	headbuf  bytes.Buffer // packet header
	msgbuf   bytes.Buffer // message pack plaintext
	ctmsgbuf []byte       // message data ciphertext

	// decode buffers
	reader bytes.Reader
}

func NewSSLCodec(localnode *admnode.LocalNode, clock mclock.Clock, privKey *ecdsa.PrivateKey) *SSLCodec {
	codec := &SSLCodec{
		localnode:    localnode,
		sessionCache: NewSessionCache(1024, clock),
		sha256:       sha256.New(),
		privKey:      privKey,
	}
	return codec
}

func (sc *SSLCodec) Encode(id admnode.NodeID, udpAddr string, packet ADMPacket, askHandshake *AskHandshake) ([]byte, Nonce, error) {
	var session *session
	var head Header
	var msgData []byte
	var err error

	switch {
	case packet.MessageType() == AskHandshakeMsg:
		head, err = sc.encodeAskHandshake(id, packet.(*AskHandshake))
	case askHandshake != nil:
		head, session, err = sc.encodeHandshakeHeader(id, udpAddr, askHandshake)
	default:
		session = sc.sessionCache.session(id, udpAddr)
		if session != nil {
			head, err = sc.encodeMessageHeader(id, session)
		} else {
			// there is no session key; send random data to start the handshake.
			head, msgData, err = sc.encodeRequireHandshakeHeader(id)
		}
	}

	if err != nil {
		return nil, Nonce{}, err
	}

	if err := generateMaskingIV(head.IV[:]); err != nil {
		return nil, Nonce{}, err
	}

	sc.writeHeaders(&head)

	if handshake, ok := packet.(*AskHandshake); ok {
		handshake.HandshakeData = bytesCopy(&sc.buf)
		sc.sessionCache.storeSentHandshake(id, udpAddr, handshake)
		// handshake.
	} else if msgData == nil {
		headerData := sc.buf.Bytes()
		msgData, err = sc.encryptMessage(session, packet, &head, headerData)
		if err != nil {
			return nil, Nonce{}, err
		}
	}

	enc, err := sc.EncodeRaw(id, head, msgData)
	return enc, head.Nonce, err
}

// Decode decodes a adamnite UDP packet
func (sc *SSLCodec) Decode(input []byte, udpAddr string) (src admnode.NodeID, n *admnode.GossipNode, p ADMPacket, err error) {
	if len(input) < sizeofStaticPacketHeader {
		return admnode.NodeID{}, nil, nil, errTooShort
	}
	var head Header
	copy(head.IV[:], input[:sizeOfIV])
	mask := head.mask(*sc.localnode.Node().ID())
	staticHeader := input[sizeOfIV:sizeofStaticPacketHeader]
	mask.XORKeyStream(staticHeader, staticHeader)

	sc.reader.Reset(staticHeader)
	binary.Read(&sc.reader, binary.BigEndian, &head.StaticHeader)

	remainingInput := len(input) - sizeofStaticPacketHeader
	if err := head.checkValid(remainingInput); err != nil {
		return admnode.NodeID{}, nil, nil, err
	}

	authDataEnd := sizeofStaticPacketHeader + int(head.AuthSize)
	authData := input[sizeofStaticPacketHeader:authDataEnd]
	mask.XORKeyStream(authData, authData)
	head.AuthData = authData

	sc.sessionCache.cleanHandshake()

	headerData := input[:authDataEnd]
	msgData := input[authDataEnd:]

	switch head.PacketType {
	case messagePHT:
		p, err = sc.decodeMessage(udpAddr, &head, headerData, msgData)
	case askHandshakePHT:
		p, err = sc.decodeAskHandshake(&head, headerData)
	case handshakeBodyPHT:
		n, p, err = sc.decodeHandshake(udpAddr, &head, headerData, msgData)
	default:
		err = errInvalidPacketType
	}

	return head.srcID, n, p, err
}

func (sc *SSLCodec) makeHeader(id admnode.NodeID, packetType byte, authSize int) Header {
	return Header{
		StaticHeader: StaticHeader{
			ProtocolID: admPacketProtocolID,
			Version:    admPacketVersionV1,
			PacketType: packetType,
			AuthSize:   uint16(authSize),
		},
	}
}

func (sc *SSLCodec) encodeAskHandshake(id admnode.NodeID, packet *AskHandshake) (Header, error) {
	head := sc.makeHeader(id, askHandshakePHT, sizeofAskHandshakeAuthData)
	head.AuthData = bytesCopy(&sc.buf)
	head.Nonce = packet.Nonce

	auth := &askHandshakeAuthData{
		RandomID:  packet.RandomID,
		DposRound: packet.DposRound,
	}

	sc.headbuf.Reset()
	binary.Write(&sc.headbuf, binary.BigEndian, auth)
	head.AuthData = sc.headbuf.Bytes()
	return head, nil
}

func (sc *SSLCodec) encodeMessageHeader(id admnode.NodeID, s *session) (Header, error) {
	head := sc.makeHeader(id, messagePHT, sizeofMessageAuthData)

	nonce, err := sc.sessionCache.nextNonce(s)
	if err != nil {
		return Header{}, fmt.Errorf("cannot generate nonce: %v", err)
	}

	auth := messageAuthData{SrcID: *sc.localnode.Node().ID()}
	sc.headbuf.Reset()
	binary.Write(&sc.headbuf, binary.BigEndian, auth)
	head.AuthData = sc.headbuf.Bytes()
	head.Nonce = nonce
	return head, err
}

func (sc *SSLCodec) encodeRequireHandshakeHeader(id admnode.NodeID) (Header, []byte, error) {
	head := sc.makeHeader(id, messagePHT, sizeofMessageAuthData)
	authData := messageAuthData{SrcID: *sc.localnode.Node().ID()}

	if _, err := crand.Read(head.Nonce[:]); err != nil {
		return head, nil, fmt.Errorf("cannot get random data: %v", err)
	}

	sc.headbuf.Reset()
	binary.Write(&sc.headbuf, binary.BigEndian, authData)
	head.AuthData = sc.headbuf.Bytes()

	sc.ctmsgbuf = append(sc.ctmsgbuf[:0], make([]byte, randomPacketMsgSize)...)
	crand.Read(sc.ctmsgbuf)
	return head, sc.ctmsgbuf, nil
}

func (sc *SSLCodec) encodeHandshakeHeader(id admnode.NodeID, udpAddr string, askHandshake *AskHandshake) (Header, *session, error) {
	if askHandshake.Node == nil {
		panic("missing node information in handshake")
	}

	auth, handshakeSession, err := sc.makeHandshakeAuth(id, udpAddr, askHandshake)
	if err != nil {
		return Header{}, nil, err
	}

	nonce, err := sc.sessionCache.nextNonce(handshakeSession)
	if err != nil {
		return Header{}, nil, fmt.Errorf("cannot generate nonce: %v", err)
	}

	sc.sessionCache.storeNewSession(id, udpAddr, handshakeSession)

	head := sc.makeHeader(id, handshakeBodyPHT, len(auth.SrcID)+binary.Size(auth.PubkeySize)+binary.Size(auth.SignatureSize)+len(auth.pubkey)+len(auth.signature)+len(auth.NodeInfo))
	sc.headbuf.Reset()
	sc.headbuf.Write(auth.SrcID[:])
	binary.Write(&sc.headbuf, binary.BigEndian, auth.PubkeySize)
	binary.Write(&sc.headbuf, binary.BigEndian, auth.SignatureSize)
	sc.headbuf.Write(auth.pubkey)
	sc.headbuf.Write(auth.signature)
	sc.headbuf.Write(auth.NodeInfo)
	head.AuthData = sc.headbuf.Bytes()
	head.Nonce = nonce
	return head, handshakeSession, nil
}

func (sc *SSLCodec) makeHandshakeAuth(id admnode.NodeID, udpAddr string, askHandshake *AskHandshake) (*handshakeAuthData, *session, error) {
	auth := new(handshakeAuthData)
	auth.SrcID = *sc.localnode.Node().ID()

	var remotePubkey = askHandshake.Node.Pubkey()

	tempKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate temporary key to sign")
	}

	tempPubKey := EncodePubkey(&tempKey.PublicKey)
	auth.pubkey = tempPubKey[:]
	auth.PubkeySize = byte(len(auth.pubkey))

	handshakeData := askHandshake.HandshakeData
	sig, err := makeSignature(sc.sha256, sc.privKey, id, handshakeData, tempPubKey[:])
	if err != nil {
		return nil, nil, fmt.Errorf("cannot sign: %v", err)
	}

	auth.signature = sig
	auth.SignatureSize = byte(len(auth.signature))

	auth.NodeInfo, _ = msgpack.Marshal(sc.localnode.NodeInfo())

	handshakeSession := deriveKeys(sha256.New, tempKey, remotePubkey, *sc.localnode.Node().ID(), *askHandshake.Node.ID(), handshakeData)
	if handshakeSession == nil {
		return nil, nil, fmt.Errorf("cannot create session key")
	}
	return auth, handshakeSession, nil
}

func (sc *SSLCodec) writeHeaders(head *Header) {
	sc.buf.Reset()
	sc.buf.Write(head.IV[:])
	binary.Write(&sc.buf, binary.BigEndian, head.StaticHeader)
	sc.buf.Write(head.AuthData)
}

func (sc *SSLCodec) encryptMessage(s *session, p ADMPacket, head *Header, headerData []byte) ([]byte, error) {
	sc.msgbuf.Reset()
	sc.msgbuf.WriteByte(p.MessageType())

	byPacket, err := msgpack.Marshal(p)
	if err != nil {
		return nil, err
	}

	sc.msgbuf.Write(byPacket)
	messagePT := sc.msgbuf.Bytes()

	messageCT, err := encryptGCM(sc.ctmsgbuf[:0], s.writekey, head.Nonce[:], messagePT, headerData)
	if err == nil {
		sc.ctmsgbuf = messageCT
	}
	return messageCT, err
}

func (sc *SSLCodec) decryptMessage(input, nonce, headerData, readKey []byte) (ADMPacket, error) {
	msgdata, err := decryptGCM(readKey, nonce, input, headerData)
	if err != nil {
		return nil, errMessageDecrypt
	}
	if len(msgdata) == 0 {
		return nil, errMessageTooShort
	}
	return DecodeMessage(msgdata[0], msgdata[1:])
}

func (sc *SSLCodec) decodeMessage(fromAddr string, head *Header, headerData, msgData []byte) (ADMPacket, error) {
	if len(head.AuthData) != sizeofMessageAuthData {
		return nil, fmt.Errorf("invalid auth size %d for message packet", len(head.AuthData))
	}

	var auth messageAuthData
	sc.reader.Reset(head.AuthData)
	binary.Read(&sc.reader, binary.BigEndian, &auth)
	head.srcID = auth.SrcID

	key := sc.sessionCache.readKey(auth.SrcID, fromAddr)
	msg, err := sc.decryptMessage(msgData, head.Nonce[:], headerData, key)
	if err == errMessageDecrypt {
		return &SYN{Nonce: head.Nonce}, nil
	}
	return msg, err
}

func (sc *SSLCodec) decodeAskHandshake(head *Header, headerData []byte) (ADMPacket, error) {
	if len(head.AuthData) != sizeofAskHandshakeAuthData {
		return nil, fmt.Errorf("invalid auth size %d for askhandshake", len(head.AuthData))
	}

	var auth askHandshakeAuthData
	sc.reader.Reset(head.AuthData)
	binary.Read(&sc.reader, binary.BigEndian, auth)

	packet := &AskHandshake{
		Nonce:         head.Nonce,
		RandomID:      auth.RandomID,
		DposRound:     auth.DposRound,
		HandshakeData: make([]byte, len(headerData)),
	}
	copy(packet.HandshakeData, headerData)
	return packet, nil
}

func (sc *SSLCodec) decodeHandshake(fromAddr string, head *Header, headerData, msgData []byte) (n *admnode.GossipNode, p ADMPacket, err error) {
	auth, err := sc.decodeHandshakeAuthData(head)
	if err != nil {
		sc.sessionCache.deleteHandshake(auth.SrcID, fromAddr)
		return nil, nil, err
	}

	handshake := sc.sessionCache.getHandshake(auth.SrcID, fromAddr)
	if handshake == nil {
		sc.sessionCache.deleteHandshake(auth.SrcID, fromAddr)
		return nil, nil, err
	}

	n, err = sc.decodeHandshakeNodeInfo(handshake.Node, auth.SrcID, auth.NodeInfo)
	if err != nil {
		sc.sessionCache.deleteHandshake(auth.SrcID, fromAddr)
		return nil, nil, err
	}

	err = verifySignature(sc.sha256, auth.signature, n, *sc.localnode.Node().ID(), handshake.HandshakeData, auth.pubkey)
	if err != nil {
		sc.sessionCache.deleteHandshake(auth.SrcID, fromAddr)
		return n, nil, err
	}

	pubkey, err := DecodePubkey(sc.privKey.Curve, auth.pubkey)
	if err != nil {
		sc.sessionCache.deleteHandshake(auth.SrcID, fromAddr)
		return n, nil, err
	}

	session := deriveKeys(sha256.New, sc.privKey, pubkey, auth.SrcID, *sc.localnode.Node().ID(), handshake.HandshakeData)
	session = session.copy()

	msg, err := sc.decryptMessage(msgData, head.Nonce[:], headerData, session.readkey)
	if err != nil {
		sc.sessionCache.deleteHandshake(auth.SrcID, fromAddr)
		return n, msg, err
	}

	sc.sessionCache.storeNewSession(auth.SrcID, fromAddr, session)
	sc.sessionCache.deleteHandshake(auth.SrcID, fromAddr)
	return n, msg, nil
}

func (sc *SSLCodec) decodeHandshakeAuthData(head *Header) (auth handshakeAuthData, err error) {
	if len(head.AuthData) < sizeofHandshakeAuthData {
		return auth, fmt.Errorf("header authsize %d too low", head.AuthSize)
	}

	sc.reader.Reset(head.AuthData)
	sc.reader.Read(auth.SrcID[:])
	auth.PubkeySize, err = sc.reader.ReadByte()
	if err != nil {
		return auth, fmt.Errorf("read error")
	}
	auth.SignatureSize, err = sc.reader.ReadByte()
	if err != nil {
		return auth, fmt.Errorf("read error")
	}
	head.srcID = auth.SrcID

	var remainData = head.AuthData[binary.Size(auth.SrcID)+binary.Size(auth.PubkeySize)+binary.Size(auth.SignatureSize):]
	auth.pubkey = remainData[:auth.PubkeySize]
	auth.signature = remainData[auth.PubkeySize : auth.PubkeySize+auth.SignatureSize]
	auth.NodeInfo = remainData[auth.PubkeySize+auth.SignatureSize:]
	return auth, nil
}

func (sc *SSLCodec) decodeHandshakeNodeInfo(local *admnode.GossipNode, wantID admnode.NodeID, remote []byte) (*admnode.GossipNode, error) {
	node := local
	if len(remote) > 0 {
		var nodeInfo admnode.NodeInfo
		if err := msgpack.Unmarshal(remote, &nodeInfo); err != nil {
			return nil, err
		}
		if local == nil || local.Version() < nodeInfo.GetVersion() || local.IP().Equal(nodeInfo.GetIP()) || local.UDP() != nodeInfo.GetUDP() || local.TCP() != nodeInfo.GetTCP() {
			n, err := admnode.New(&nodeInfo)
			if err != nil {
				return nil, fmt.Errorf("invalid node info: %v", err)
			}
			if *n.ID() != wantID {
				return nil, fmt.Errorf("wrong ID: %v", n.ID())
			}
			node = n
		}
	}
	if node == nil {
		return nil, errNoRecord
	}
	return node, nil
}

func (sc *SSLCodec) EncodeRaw(id admnode.NodeID, head Header, msgdata []byte) ([]byte, error) {
	sc.writeHeaders(&head)

	masked := sc.buf.Bytes()[sizeOfIV:]
	mask := head.mask(id)
	mask.XORKeyStream(masked[:], masked[:])

	sc.buf.Write(msgdata)
	return sc.buf.Bytes(), nil
}

func DecodeMessage(packetType byte, body []byte) (ADMPacket, error) {
	var dec ADMPacket
	switch packetType {
	case FindnodeMsg:
		dec = new(Findnode)
	case RspFindnodeMsg:
		dec = new(RspNodes)
	case PingMsg:
		dec = new(Ping)
	case PongMsg:
		dec = new(Pong)
	default:
		return nil, fmt.Errorf("unknown packet type %d", packetType)
	}

	if err := msgpack.Unmarshal(body, dec); err != nil {
		return nil, err
	}
	if dec.RequestID() != nil && len(dec.RequestID()) > 8 {
		return nil, ErrInvalidReqID
	}
	return dec, nil
}
