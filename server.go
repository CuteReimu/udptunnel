package main

import (
	"fmt"
	_ "github.com/CuteReimu/cellnet-plus/kcp"
	"github.com/CuteReimu/udptunnel/pb"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/udp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"github.com/davyxu/cellnet/rpc"
	"os"
	"os/signal"
	"time"
)

type server struct {
	port    uint32
	timeout time.Duration
	cache   map[int64]*serverClient // from_client_id => server_client
	peer    cellnet.TCPConnector
}

func (s *server) start(address string) {
	ch := make(chan os.Signal, 1)
	s.cache = make(map[int64]*serverClient)
	queue := cellnet.NewEventQueue()
	queue.EnableCapturePanic(true)
	s.peer = peer.NewGenericPeer("kcp.Connector", "server", fmt.Sprint(address, ":", *tunnelPort), queue).(cellnet.TCPConnector)
	proc.BindProcessorHandler(s.peer, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			log.Debugln("server connected: ", ev.Session().ID())
			go s.heart()
			s.createServerRoom()
		case *cellnet.SessionConnectError:
			log.Errorln("server connect failed: ", msg.String())
			time.AfterFunc(3*time.Second, func() { ch <- os.Interrupt })
		case *cellnet.SessionClosed:
			log.Debugln("session closed: ", ev.Session().ID())
			time.AfterFunc(3*time.Second, func() { ch <- os.Interrupt })
		case *pb.UdpToc:
			cli, ok := s.cache[msg.FromId]
			if !ok {
				cli = &serverClient{server: s, id: msg.FromId}
				cli.start()
				s.cache[msg.FromId] = cli
			}
			cli.lastMsgTime = time.Now()
			cli.peer.Session().Send(&UDPMessage{Msg: msg.Data})
		}
	})
	s.peer.Start()
	queue.StartLoop()
	go s.startTimeoutTimer()
	signal.Notify(ch, os.Interrupt)
	<-ch
	queue.StopLoop()
	for _, c := range s.cache {
		c.peer.Queue().StopLoop()
	}
	queue.Wait()
	for _, c := range s.cache {
		c.peer.Queue().Wait()
	}
	for _, c := range s.cache {
		c.peer.Stop()
	}
	s.peer.Stop()
}

func (s *server) heart() {
	var count int
	ch := time.Tick(15 * time.Second)
	for {
		<-ch
		_, err := rpc.CallSync(s.peer.Session(), &pb.HeartTos{}, time.Second*3)
		if err != nil {
			count++
			if count >= 10 {
				log.Errorf("服务器已断开，程序即将关闭：%s", err.Error())
				s.peer.Session().Close()
				break
			}
			log.Errorf("服务器已断开，请检查网络连接：%s", err.Error())
		} else {
			count = 0
		}
	}
}

func (s *server) startTimeoutTimer() {
	ch := time.Tick(time.Minute)
	for {
		<-ch
		s.peer.Queue().Post(s.removeTimeoutClient)
	}
}

func (s *server) removeTimeoutClient() {
	if len(s.cache) == 0 {
		return
	}
	deleteClient := make([]int64, 0, len(s.cache))
	now := time.Now()
	for _, c := range s.cache {
		if c.lastMsgTime.Add(s.timeout).Before(now) {
			deleteClient = append(deleteClient, c.id)
			p := c.peer
			p.Queue().StopLoop()
			go func() {
				p.Queue().Wait()
				p.Stop()
			}()
		}
	}
	for _, id := range deleteClient {
		delete(s.cache, id)
	}
}

func (s *server) createServerRoom() {
	rpc.Call(s.peer.Session(), &pb.CreateServerTos{Port: s.port}, time.Second*3, func(raw interface{}) {
		if err, ok := raw.(error); ok {
			log.Errorln("创建服务器房间失败：", err)
			s.peer.Session().Close()
		} else if resp, ok := raw.(*pb.CreateServerToc); ok && !resp.Success {
			log.Errorln("创建服务器房间失败：", resp)
			s.peer.Session().Close()
		} else {
			log.Infof("创建服务器房间成功，等待客户端加入")
		}
	})
}

type serverClient struct {
	server      *server
	id          int64
	peer        cellnet.UDPConnector
	lastMsgTime time.Time
}

func (c *serverClient) start() {
	queue := cellnet.NewEventQueue()
	queue.EnableCapturePanic(true)
	c.peer = peer.NewGenericPeer("udp.Connector", "server", fmt.Sprint("127.0.0.1:", c.server.port), queue).(cellnet.UDPConnector)
	proc.BindProcessorHandler(c.peer, "udp.pure", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *UDPMessage:
			c.server.peer.Session().Send(&pb.UdpTos{ToId: c.id, Data: msg.Msg})
		}
	})
	c.peer.Start()
	queue.StartLoop()
}
