package main

import (
	"os"
	"os/signal"
	"strconv"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	_ "github.com/davyxu/cellnet/peer/udp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
)

type client struct {
	address string
}

func (c *client) start() {
	ch := make(chan os.Signal, 1)
	queue := cellnet.NewEventQueue()
	queue.EnableCapturePanic(true)
	p := peer.NewGenericPeer("tcp.Connector", "server", c.address+":"+strconv.Itoa(*tunnelPort), queue)
	var cs *clientServer
	proc.BindProcessorHandler(p, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			log.Debugln("client connected: ", ev.Session().ID())
		case *cellnet.SessionConnectError:
			log.Errorln("client connect failed: ", msg.String())
			ch <- os.Interrupt
		case *cellnet.SessionClosed:
			log.Debugln("session closed: ", ev.Session().ID())
			ch <- os.Interrupt
		case *GetPortToc:
			cs = &clientServer{client: c, tcpSession: ev.Session(), port: msg.Port}
			cs.start()
		case *UDPMessage:
			cs.send(msg)
		}
	})
	p.Start()
	queue.StartLoop()
	signal.Notify(ch, os.Interrupt)
	<-ch
	queue.StopLoop()
	queue.Wait()
	if cs != nil {
		cs.stop()
	}
	p.Stop()
}

type clientServer struct {
	port       int
	client     *client
	tcpSession cellnet.Session
	udpSession cellnet.Session
	queue      cellnet.EventQueue
	peer       cellnet.Peer
}

func (c *clientServer) start() {
	c.queue = cellnet.NewEventQueue()
	c.queue.EnableCapturePanic(true)
	c.peer = peer.NewGenericPeer("udp.Acceptor", "server", "0.0.0.0:"+strconv.Itoa(c.port), c.queue)
	proc.BindProcessorHandler(c.peer, "udp.pure", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			log.Debugln("server accepted: ", ev.Session().ID())
			c.udpSession = ev.Session()
		case *cellnet.SessionClosed:
			log.Debugln("session closed: ", ev.Session().ID())
		case *UDPMessage:
			c.tcpSession.Send(msg)
		}
	})
	c.peer.Start()
	c.queue.StartLoop()
	log.Infoln("udp start listen, port: ", &c)
}

func (c *clientServer) send(msg *UDPMessage) {
	c.queue.Post(func() {
		if c.udpSession != nil {
			c.udpSession.Send(msg)
		}
	})
}

func (c *clientServer) stop() {
	c.queue.StopLoop()
	c.queue.Wait()
	c.peer.Stop()
}
