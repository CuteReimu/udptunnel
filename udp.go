package main

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/peer/udp"
	"github.com/davyxu/cellnet/proc"
	"reflect"
	"strings"
)

type UDPMessage struct {
	Msg []byte
}

func (m *UDPMessage) String() string {
	ret := make([]string, len(m.Msg))
	for i, b := range m.Msg {
		ret[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(ret, " ")
}

type UDPMessageTransmitter struct {
}

func (UDPMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {
	data := ses.Raw().(udp.DataReader).ReadData()
	m := make([]byte, len(data))
	copy(m, data)
	msg = &UDPMessage{Msg: m}
	msglog.WriteRecvLogger(log, "udp", ses, msg)
	return
}

func (UDPMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {
	writer := ses.(udp.DataWriter)
	message, ok := msg.(*UDPMessage)
	if !ok {
		log.Warnf("unsupported message type: %T", message)
		return nil
	}
	msglog.WriteSendLogger(log, "udp", ses, msg)
	writer.WriteData(message.Msg)
	return nil
}

func init() {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*UDPMessage)(nil)).Elem(),
		ID:    9999,
	})
	proc.RegisterProcessor("udp.pure", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(UDPMessageTransmitter))
		bundle.SetCallback(userCallback)
	})
}
