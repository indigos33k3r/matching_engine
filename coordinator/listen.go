package coordinator

import (
	"fmt"
	"github.com/fmstephe/matching_engine/msg"
	"github.com/fmstephe/matching_engine/msg/msgutil"
	"io"
)

type stdListener struct {
	reader io.ReadCloser
	ticker *msgutil.Ticker
	msgHelper
}

func newListener(reader io.ReadCloser) *stdListener {
	return &stdListener{reader: reader, ticker: msgutil.NewTicker()}
}

func (l *stdListener) Run() {
	defer l.shutdown()
	for {
		m := l.deserialise()
		shutdown := l.forward(m)
		if shutdown {
			return
		}
	}
}

func (l *stdListener) deserialise() *msg.Message {
	b := make([]byte, msg.SizeofMessage)
	m := &msg.Message{}
	n, err := l.reader.Read(b)
	m.WriteFrom(b[:n])
	if err != nil {
		m.Status = msg.READ_ERROR
		l.logErr("Listener - UDP Read: " + err.Error())
	} else if n != msg.SizeofMessage {
		m.Status = msg.SMALL_READ_ERROR
		l.logErr(fmt.Sprintf("Listener: Error incorrect number of bytes. Expecting %d, found %d in %v", msg.SizeofMessage, n, b))
	}
	return m
}

func (l *stdListener) logErr(errStr string) {
	if l.log {
		println(errStr)
	}
}

func (l *stdListener) forward(o *msg.Message) (shutdown bool) {
	if o.Route != msg.ACK && o.Route != msg.SHUTDOWN {
		a := &msg.Message{}
		a.WriteAckFor(o)
		l.msgs <- a
	}
	if o.Route == msg.SHUTDOWN || o.Route == msg.ACK || l.ticker.Tick(o) {
		l.msgs <- o
	}
	return o.Route == msg.SHUTDOWN
}

func (l *stdListener) shutdown() {
	l.reader.Close()
}
