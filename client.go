package main

import (
	"fmt"
	_ "github.com/CuteReimu/cellnet-plus/kcp"
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

type client struct {
	timeout time.Duration
	peer    cellnet.TCPConnector
	cs      *clientServer
}

func (c *client) start(address string) {
	ch := make(chan os.Signal, 1)
	queue := cellnet.NewEventQueue()
	queue.EnableCapturePanic(true)
	c.peer = peer.NewGenericPeer("kcp.Connector", "server", fmt.Sprint(address+":", *tunnelPort), queue).(cellnet.TCPConnector)
	proc.BindProcessorHandler(c.peer, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			log.Debugln("client connected: ", ev.Session().ID())
			go c.heart()
			rpc.Call(ev.Session(), &GetAllServersTos{}, time.Second*3, c.waitForChooseServer)
		case *cellnet.SessionConnectError:
			log.Errorln("client connect failed: ", msg.String())
			time.AfterFunc(3*time.Second, func() { ch <- os.Interrupt })
		case *cellnet.SessionClosed:
			log.Debugln("session closed: ", ev.Session().ID())
			time.AfterFunc(3*time.Second, func() { ch <- os.Interrupt })
		case *UdpToc:
			if c.cs != nil {
				c.cs.send(&UDPMessage{Msg: msg.Data})
			}
		}
	})
	c.peer.Start()
	queue.StartLoop()
	signal.Notify(ch, os.Interrupt)
	<-ch
	queue.StopLoop()
	if c.cs != nil {
		c.cs.peer.Queue().StopLoop()
	}
	queue.Wait()
	if c.cs != nil {
		c.cs.peer.Queue().Wait()
		c.cs.peer.Stop()
	}
	c.peer.Stop()
}

func (c *client) waitForChooseServer(raw interface{}) {
	switch msg := raw.(type) {
	case error:
		fmt.Println("获取服务器列表超时，程序将在3秒后关闭")
		time.AfterFunc(3*time.Second, func() { c.peer.Session().Close() })
	case *GetAllServersToc:
		serverList := msg.List
		if len(serverList) == 0 {
			fmt.Println("没有找到服务器房间，程序将在3秒后关闭")
			time.AfterFunc(3*time.Second, func() { c.peer.Session().Close() })
			break
		}
		for {
			fmt.Println("当前服务器列表：")
			for _, svr := range serverList {
				fmt.Println(svr.Id, "  ", svr.Address)
			}
			fmt.Println("请选择你要连接的服务器：")
			id := input[int64]("请输入你要连接的服务器ID：")
			for _, svr := range serverList {
				if svr.Id == id {
					c.cs = &clientServer{serverId: id, client: c, port: svr.Port}
					c.cs.start()
					return
				}
			}
			fmt.Println("你输入的服务器ID不存在")
		}
	}
}

func (c *client) heart() {
	var count int
	ch := time.Tick(15 * time.Second)
	for {
		<-ch
		_, err := rpc.CallSync(c.peer.Session(), &HeartTos{}, time.Second*3)
		if err != nil {
			count++
			if count >= 10 {
				log.Errorf("服务器已断开，程序即将关闭：%s", err.Error())
				c.peer.Session().Close()
				break
			}
			log.Errorf("服务器已断开，请检查网络连接：%s", err.Error())
		} else {
			count = 0
		}
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
			c.client.peer.Session().Send(&UdpTos{ToId: c.serverId, Data: msg.Msg})
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
