package main

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	_ "github.com/davyxu/cellnet/peer/udp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"os"
	"os/signal"
)

type server struct {
	port int64
}

func (s *server) start() {
	cache := make(map[int64]*serverClient)
	queue := cellnet.NewEventQueue()
	queue.EnableCapturePanic(true)
	p := peer.NewGenericPeer("tcp.Acceptor", "server", fmt.Sprint("0.0.0.0:", *tunnelPort), queue)
	proc.BindProcessorHandler(p, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			log.Debugln("server accepted: ", ev.Session().ID())
			c := &serverClient{server: s, tcpSession: ev.Session()}
			c.start()
			cache[ev.Session().ID()] = c
		case *cellnet.SessionClosed:
			log.Debugln("session closed: ", ev.Session().ID())
			delete(cache, ev.Session().ID())
		case *GetPortTos:
			ev.Session().Send(&GetPortToc{Port: s.port})
		case *UDPMessageTos:
			cache[ev.Session().ID()].send(&UDPMessage{Msg: append(([]byte)(nil), msg.Msg...)})
		}
	})
	p.Start()
	queue.StartLoop()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	queue.StopLoop()
	queue.Wait()
	for _, c := range cache {
		c.stop()
	}
	p.Stop()
}

type serverClient struct {
	server     *server
	tcpSession cellnet.Session
	udpSession cellnet.Session
	queue      cellnet.EventQueue
	peer       cellnet.GenericPeer
}

func (c *serverClient) start() {
	c.queue = cellnet.NewEventQueue()
	c.queue.EnableCapturePanic(true)
	c.peer = peer.NewGenericPeer("udp.Connector", "server", fmt.Sprint("127.0.0.1:", c.server.port), c.queue)
	proc.BindProcessorHandler(c.peer, "udp.pure", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			c.udpSession = ev.Session()
		case *UDPMessage:
			c.tcpSession.Send(&UDPMessageToc{Msg: append(([]byte)(nil), msg.Msg...)})
		}
	})
	c.peer.Start()
	c.queue.StartLoop()
}

func (c *serverClient) send(msg *UDPMessage) {
	c.queue.Post(func() {
		if c.udpSession != nil {
			c.udpSession.Send(msg)
		}
	})
}

func (c *serverClient) stop() {
	c.queue.StopLoop()
	c.queue.Wait()
	c.peer.Stop()
}
