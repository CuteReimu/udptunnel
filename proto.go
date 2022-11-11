package main

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	_ "github.com/davyxu/cellnet/codec/binary"
	"reflect"
)

type HeartTos struct {
}

type HeartToc struct {
}

type PbServer struct {
	Id      int64
	Address string
	Port    uint32
}

type CreateServerTos struct {
	Port uint32
}

type CreateServerToc struct {
	Success bool
}

type GetAllServersTos struct {
}

type GetAllServersToc struct {
	List []*PbServer
}

type UdpTos struct {
	ToId int64
	Data []byte
}

type UdpToc struct {
	FromId int64
	Data   []byte
}

// 将消息注册到系统
func init() {
	c := codec.MustGetCodec("binary")
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*HeartTos)(nil)).Elem(),
		ID:    1,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*HeartToc)(nil)).Elem(),
		ID:    2,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*CreateServerTos)(nil)).Elem(),
		ID:    3,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*CreateServerToc)(nil)).Elem(),
		ID:    4,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*GetAllServersTos)(nil)).Elem(),
		ID:    5,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*GetAllServersToc)(nil)).Elem(),
		ID:    6,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*UdpTos)(nil)).Elem(),
		ID:    7,
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*UdpToc)(nil)).Elem(),
		ID:    8,
	})
}
