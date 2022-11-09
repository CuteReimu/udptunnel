package main

import (
	"fmt"
	"github.com/CuteReimu/udptunnel/pb"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	_ "github.com/davyxu/cellnet/peer/udp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"os"
	"os/signal"
	"time"
)

type client struct {
	timeout time.Duration
	peer    cellnet.TCPConnector
}

func (c *client) start(address string) {
	ch := make(chan os.Signal, 1)
	queue := cellnet.NewEventQueue()
	queue.EnableCapturePanic(true)
	c.peer = peer.NewGenericPeer("tcp.Connector", "server", fmt.Sprint(address+":", *tunnelPort), queue).(cellnet.TCPConnector)
	var cs *clientServer
	proc.BindProcessorHandler(c.peer, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			log.Debugln("client connected: ", ev.Session().ID())
			go c.heart()
			ev.Session().Send(&pb.GetAllServersTos{})
		case *cellnet.SessionConnectError:
			log.Errorln("client connect failed: ", msg.String())
			time.AfterFunc(3*time.Second, func() { ch <- os.Interrupt })
		case *cellnet.SessionClosed:
			log.Debugln("session closed: ", ev.Session().ID())
			ch <- os.Interrupt
		case *pb.GetAllServersToc:
			if len(msg.List) == 0 {
				fmt.Println("没有找到服务器房间，程序将在3秒后关闭")
				time.AfterFunc(3*time.Second, func() { ch <- os.Interrupt })
			} else {
				go c.waitForChooseServer(msg.List)
			}
		case *pb.UdpToc:
			cs.send(&UDPMessage{Msg: msg.Data})
		}
	})
	c.peer.Start()
	queue.StartLoop()
	signal.Notify(ch, os.Interrupt)
	<-ch
	queue.StopLoop()
	if cs != nil {
		c.peer.Queue().StopLoop()
	}
	queue.Wait()
	if cs != nil {
		c.peer.Queue().Wait()
		cs.peer.Stop()
	}
	c.peer.Stop()
}

func (c *client) waitForChooseServer(serverList []*pb.PbServer) {
	for {
		fmt.Println("当前服务器列表：")
		for _, svr := range serverList {
			fmt.Println(svr.Id, "  ", svr.Address)
		}
		fmt.Println("请选择你要连接的服务器：")
		id := input[int64]("请输入你要连接的服务器ID：")
		for _, svr := range serverList {
			if svr.Id == id {
				(&clientServer{serverId: id, client: c, port: svr.Port}).start()
				return
			}
		}
		fmt.Println("你输入的服务器ID不存在")
	}
}

func (c *client) heart() {
	ch := time.Tick(15 * time.Second)
	for {
		<-ch
		c.peer.Session().Send(&pb.HeartTos{})
	}
}

type clientServer struct {
	port       uint32
	client     *client
	serverId   int64
	udpSession cellnet.Session
	peer       cellnet.GenericPeer
}

func (c *clientServer) start() {
	queue := cellnet.NewEventQueue()
	queue.EnableCapturePanic(true)
	c.peer = peer.NewGenericPeer("udp.Acceptor", "server", fmt.Sprint("0.0.0.0:", c.port), queue)
	proc.BindProcessorHandler(c.peer, "udp.pure", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *UDPMessage:
			c.udpSession = ev.Session()
			c.client.peer.Session().Send(&pb.UdpTos{ToId: c.serverId, Data: msg.Msg})
		}
	})
	c.peer.Start()
	queue.StartLoop()
	log.Infoln("udp start listen, port: ", c.port)
}

func (c *clientServer) send(msg *UDPMessage) {
	c.peer.Queue().Post(func() {
		if c.udpSession != nil {
			c.udpSession.Send(msg)
		}
	})
}
