package main

import (
	"fmt"
	_ "github.com/CuteReimu/cellnet-plus/codec/protobuf"
	_ "github.com/CuteReimu/cellnet-plus/peer/kcp"
	"github.com/CuteReimu/cellnet-plus/util"
	"github.com/CuteReimu/udptunnel/pb"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/cellnet/rpc"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"time"
)

//go:generate protoc --proto_path=. --go_out=. vpn.proto
const lastHeartTime = "last_heart_time"

type vpnServer struct {
	timeout time.Duration
	peer    cellnet.TCPAcceptor
}

func (s *vpnServer) start() {
	s.peer = peer.NewGenericPeer("kcp.Acceptor", "server", fmt.Sprint("0.0.0.0:", *tunnelPort), nil).(cellnet.TCPAcceptor)
	proc.BindProcessorHandler(s.peer, "tcp.ltv", func(ev cellnet.Event) {
		defer func() {
			if e := recover(); e != nil {
				log.Errorf("%v \n%s\n", e, string(debug.Stack()))
			}
		}()
		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			log.Debugln("server accepted: ", ev.Session().ID())
			ctx := ev.Session().(cellnet.ContextSet)
			ctx.SetContext(lastHeartTime, time.Now())
		case *cellnet.SessionClosed:
			log.Debugln("session closed: ", ev.Session().ID())
		case *pb.HeartTos:
			ev.Session().(cellnet.ContextSet).SetContext(lastHeartTime, time.Now())
			if rpcEvent, ok := ev.(*rpc.RecvMsgEvent); ok {
				rpcEvent.Reply(&pb.HeartToc{})
			} else {
				ev.Session().Send(&pb.HeartToc{})
			}
		case *pb.CreateServerTos:
			resp := &pb.CreateServerToc{}
			if msg.Port < 65536 {
				ev.Session().(cellnet.ContextSet).SetContext("port", msg.Port)
				resp.Success = true
			}
			if rpcEvent, ok := ev.(*rpc.RecvMsgEvent); ok {
				rpcEvent.Reply(resp)
			} else {
				ev.Session().Send(resp)
			}
		case *pb.GetAllServersTos:
			resp := &pb.GetAllServersToc{}
			s.peer.VisitSession(func(session cellnet.Session) bool {
				conn := session.(interface{ Conn() net.Conn }).Conn()
				svr := &pb.PbServer{Id: session.ID()}
				if session.(cellnet.ContextSet).FetchContext("port", &svr.Port) {
					svr.Address = conn.RemoteAddr().String()
					udpAddr, err := net.ResolveUDPAddr(conn.RemoteAddr().Network(), svr.Address)
					if err == nil {
						udpAddr.Port = int(svr.Port)
						svr.Address = udpAddr.String()
					}
					resp.List = append(resp.List, svr)
				}
				return true
			})
			if rpcEvent, ok := ev.(*rpc.RecvMsgEvent); ok {
				rpcEvent.Reply(resp)
			} else {
				ev.Session().Send(resp)
			}
		case *pb.UdpTos:
			session := s.peer.GetSession(msg.ToId)
			if session == nil {
				log.Warnln("cannot find vpn client, id: ", msg.ToId)
			} else {
				session.Send(&pb.UdpToc{FromId: ev.Session().ID(), Data: msg.Data})
			}
		}
	})
	s.peer.Start()
	go s.startRemoveTimeoutTimer()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	s.peer.Stop()
}

func (s *vpnServer) startRemoveTimeoutTimer() {
	ch := time.Tick(15 * time.Second)
	for range ch {
		if s.peer.SessionCount() == 0 {
			continue
		}
		now := time.Now()
		s.peer.VisitSession(func(session cellnet.Session) bool {
			lt, _ := session.(cellnet.ContextSet).GetContext(lastHeartTime)
			if lt.(time.Time).Add(s.timeout).Before(now) {
				session.Close()
			}
			return true
		})
	}
}

func init() {
	util.RegisterAllProtobuf()
}
