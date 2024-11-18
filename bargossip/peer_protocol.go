package bargossip

import (
	"errors"
	"io"
)

type protoRW struct {
	SubProtocol
	in     chan Msg        // receives read messages
	closed <-chan struct{} // receives when peer is shutting down
	wstart <-chan struct{} // receives when write may start
	werr   chan<- error    // for write results
	w      MsgWriter
}

func (rw *protoRW) WriteMsg(msg Msg) (err error) {
	if msg.Code >= uint64(rw.SubProtocol.ProtocolCodeOffset) {
		return errors.New("invalid message code")
	}

	msg.Code += uint64(rw.ProtocolCodeOffset)

	select {
	case <-rw.wstart:
		err = rw.w.WriteMsg(msg)
		// Report write status back to Peer.run. It will initiate
		// shutdown if the error is non-nil and unblock the next write
		// otherwise. The calling protocol code should exit for errors
		// as well but we don't want to rely on that.
		rw.werr <- err
	case <-rw.closed:
		err = errShuttingDown
	}
	return err
}

func (rw *protoRW) ReadMsg() (Msg, error) {
	select {
	case msg := <-rw.in:
		msg.Code -= uint64(rw.ProtocolCodeOffset)
		return msg, nil
	case <-rw.closed:
		return Msg{}, io.EOF
	}
}
