package main

import (
	"fmt"
	"reflect"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/peer/udp"
	"github.com/davyxu/cellnet/proc"
)

type GetPortTos struct {
}

func (i *GetPortTos) String() string {
	return fmt.Sprintf("%+v", *i)
}

type GetPortToc struct {
	Port int64
}

func (i *GetPortToc) String() string {
	return fmt.Sprintf("%+v", *i)
}

type UDPMessage struct {
	Msg []byte
}

type UDPMessageTransmitter struct {
}

func (UDPMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {
	data := ses.Raw().(udp.DataReader).ReadData()
	msg = &UDPMessage{Msg: data}
	msglog.WriteRecvLogger(log, "udp", ses, msg)
	return
}

func (UDPMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {
	writer := ses.(udp.DataWriter)
	msglog.WriteSendLogger(log, "udp", ses, msg)
	message, ok := msg.(*UDPMessage)
	if !ok {
		log.Warnf("unsupported message type: %T", message)
		return nil
	}
	writer.WriteData(message.Msg)
	return nil
}

// 将消息注册到系统
func init() {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*GetPortTos)(nil)).Elem(),
		ID:    1,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*GetPortToc)(nil)).Elem(),
		ID:    2,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("binary"),
		Type:  reflect.TypeOf((*UDPMessage)(nil)).Elem(),
		ID:    3,
	})
	proc.RegisterProcessor("udp.pure", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(UDPMessageTransmitter))
		bundle.SetCallback(userCallback)
	})
}
