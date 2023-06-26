package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
)

const FunctionIdentifierLength = 16

type RuntimeChanges struct {
	Caller            common.Address //who called this
	CallTime          time.Time      //when was it called
	ContractCalled    common.Address //what was called
	ParametersPassed  []byte         //what was passed on the call
	GasLimit          uint64         //did they set a gas limit
	GasUsed           uint64         //how much gas was used to process this
	ChangeStartPoints []uint64       //data from the results
	Changed           [][]byte       //^
	ErrorsEncountered error          //if anything went wrong at runtime
}

func (rtc RuntimeChanges) OutputChanges() {
	for i, x := range rtc.ChangeStartPoints {
		dataString := "\n"
		for foo := 0; foo < len(rtc.Changed[i]); foo += 8 {
			dataString = fmt.Sprintf(
				"%v\t (%3v) \t0x % X  \n",
				dataString, binary.LittleEndian.Uint64(rtc.Changed[i][foo:foo+8]),
				rtc.Changed[i][foo:foo+8],
			)
		}
		log.Printf("change at real index %v, byte index: %v, changed to: %v", x/8, x, dataString)

	}
	log.Println()
}

// returns a copy without any changes. Ideal for running again to check consistency
func (rtc RuntimeChanges) CleanCopy() *RuntimeChanges {
	ans := RuntimeChanges{
		Caller:            rtc.Caller,
		ContractCalled:    rtc.ContractCalled,
		GasLimit:          rtc.GasLimit,
		ParametersPassed:  rtc.ParametersPassed,
		ChangeStartPoints: []uint64{},
		Changed:           [][]byte{},
		ErrorsEncountered: nil,
	}
	return &ans
}

// returns the hash of the runtime changes.
func (rtc RuntimeChanges) Hash() common.Hash {
	timeBytes := binary.LittleEndian.AppendUint64([]byte{}, uint64(rtc.CallTime.UnixMilli()))
	data := append(rtc.Caller.Bytes(), timeBytes...)
	data = append(data, rtc.ContractCalled[:]...)
	data = append(data, rtc.ParametersPassed...)
	data = append(data, binary.LittleEndian.AppendUint64([]byte{}, rtc.GasLimit)[:]...)
	for i, start := range rtc.ChangeStartPoints {
		data = append(data, binary.LittleEndian.AppendUint64([]byte{}, start)[:]...)
		data = append(data, rtc.Changed[i]...)
	}
	if rtc.ErrorsEncountered != nil {
		data = append(data, []byte(rtc.ErrorsEncountered.Error())...)
	}

	return common.BytesToHash(crypto.Sha512(data))
}
func (a RuntimeChanges) Equal(b *RuntimeChanges) bool {
	if bytes.Equal(a.Caller.Bytes(), b.Caller.Bytes()) &&
		bytes.Equal(a.ContractCalled[:], b.ContractCalled[:]) &&
		a.GasLimit == b.GasLimit &&
		a.ErrorsEncountered == b.ErrorsEncountered &&
		len(a.ChangeStartPoints) == len(b.ChangeStartPoints) &&
		len(a.Changed) == len(b.Changed) &&
		a.GasUsed == b.GasUsed {

		//check the change start points
		for i, aStart := range a.ChangeStartPoints {
			if b.ChangeStartPoints[i] != aStart {
				return false
			}
		}
		//check the changed values
		for i, aChanged := range a.Changed {
			if !bytes.Equal(aChanged, b.Changed[i]) {
				return false
			}
		}

		return true
	}
	return false
}
