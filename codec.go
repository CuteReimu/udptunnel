package main

import (
	_ "github.com/CuteReimu/udptunnel/pb"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"reflect"
)

type protoCodec struct {
}

// Name 编码器的名称
func (c *protoCodec) Name() string {
	return "protobuf"
}

func (c *protoCodec) MimeType() string {
	return "application/x-protobuf"
}

func (c *protoCodec) Encode(msgObj interface{}, _ cellnet.ContextSet) (data interface{}, err error) {
	return proto.Marshal(msgObj.(proto.Message))
}

func (c *protoCodec) Decode(data interface{}, msgObj interface{}) error {
	return proto.Unmarshal(data.([]byte), msgObj.(proto.Message))
}

// 将消息注册到系统
func init() {
	c := new(protoCodec)
	// 注册所有协议
	protoregistry.GlobalTypes.RangeMessages(func(messageType protoreflect.MessageType) bool {
		cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
			Codec: c,
			Type:  reflect.TypeOf(messageType.Zero().Interface()).Elem(),
			ID:    int(util.StringHash(string(messageType.Descriptor().Name()))),
		})
		return true
	})
}
