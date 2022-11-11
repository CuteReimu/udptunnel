package main

import (
	"fmt"
	_ "github.com/CuteReimu/cellnet-plus/kcp"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/cellnet/rpc"
	"net"
	"os"
	"os/signal"
	"time"
)

const lastHeartTime = "last_heart_time"

type vpnServer struct {
	timeout time.Duration
	peer    cellnet.TCPAcceptor
}

func (s *vpnServer) start() {
	queue := cellnet.NewEventQueue()
	queue.EnableCapturePanic(true)
	s.peer = peer.NewGenericPeer("kcp.Acceptor", "server", fmt.Sprint("0.0.0.0:", *tunnelPort), queue).(cellnet.TCPAcceptor)
	proc.BindProcessorHandler(s.peer, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			log.Debugln("server accepted: ", ev.Session().ID())
			ctx := ev.Session().(cellnet.ContextSet)
			ctx.SetContext(lastHeartTime, time.Now())
		case *cellnet.SessionClosed:
			log.Debugln("session closed: ", ev.Session().ID())
		case *HeartTos:
			ev.Session().(cellnet.ContextSet).SetContext(lastHeartTime, time.Now())
			if rpcEvent, ok := ev.(*rpc.RecvMsgEvent); ok {
				rpcEvent.Reply(&HeartToc{})
			} else {
				ev.Session().Send(&HeartToc{})
			}
		case *CreateServerTos:
			resp := &CreateServerToc{}
			if msg.Port < 65536 {
				ev.Session().(cellnet.ContextSet).SetContext("port", msg.Port)
				resp.Success = true
			}
			if rpcEvent, ok := ev.(*rpc.RecvMsgEvent); ok {
				rpcEvent.Reply(resp)
			} else {
				ev.Session().Send(resp)
			}
		case *GetAllServersTos:
			resp := &GetAllServersToc{}
			s.peer.VisitSession(func(session cellnet.Session) bool {
				conn := session.(interface{ Conn() net.Conn }).Conn()
				svr := &PbServer{Id: session.ID()}
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
		case *UdpTos:
			session := s.peer.GetSession(msg.ToId)
			if session == nil {
				log.Warnln("cannot find vpn client, id: ", msg.ToId)
			} else {
				session.Send(&UdpToc{FromId: ev.Session().ID(), Data: msg.Data})
			}
		}
	})
	s.peer.Start()
	queue.StartLoop()
	go s.startRemoveTimeoutTimer()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	queue.StopLoop()
	queue.Wait()
	s.peer.Stop()
}

func (s *vpnServer) startRemoveTimeoutTimer() {
	ch := time.Tick(15 * time.Second)
	for {
		<-ch
		s.peer.Queue().Post(s.removeTimeoutRoom)
	}
}

func (s *vpnServer) removeTimeoutRoom() {
	if s.peer.SessionCount() == 0 {
		return
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
