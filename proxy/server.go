package proxy

import (
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type SocketServer struct {
	ListenAddress string
	ListenPort    int
	Destination   string

	listener *net.UDPConn
	agents   map[string]*agent
	mu       sync.Mutex
}

func New(listen, dest string) *SocketServer {
	return &SocketServer{
		ListenAddress: listen,
		ListenPort:    27960,
		Destination:   dest,
		agents:        make(map[string]*agent),
	}
}

func (s *SocketServer) listen() error {
	addr := net.UDPAddr{
		Port: s.ListenPort,
		IP:   net.ParseIP(s.ListenAddress),
	}
	listener, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		return err
	}

	s.listener = listener
	return nil
}

func (s *SocketServer) Start() error {
	err := s.listen()
	if err != nil {
		return err
	}

	// Таймер для проверки активности агентов
	go func() {
		for {
			time.Sleep(1 * time.Minute) // Проверяем каждую минуту
			s.mu.Lock()
			for addr, a := range s.agents {
				a.mu.Lock()
				if !a.running || time.Since(a.lastActivity) > 5*time.Minute {
					a.running = false
					a.ws.Close()
					close(a.udpData)
					delete(s.agents, addr)
					if logNewConnections {
						logrus.WithField("remote", addr).Info("Client disconnected due to inactivity")
					}
				}
				a.mu.Unlock()
			}
			s.mu.Unlock()
		}
	}()

	p := make([]byte, 65535)

	for {
		n, addr, err := s.listener.ReadFromUDP(p)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Error("Could not ReadFromUDP() a connection")
			continue
		}

		data := make([]byte, n)
		copy(data, p[:n])

		a := s.getAgent(addr)
		if a != nil {
			a.udpData <- data
		}
	}
}

func (s *SocketServer) getAgent(addr *net.UDPAddr) *agent {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.agents[addr.String()]
	if ok && a.running {
		return a
	}

	// WebSocket client
	u := url.URL{Scheme: "ws", Host: s.Destination, Path: "/"}
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.WithField("err", err).Error("Could not dial WebSocket")
		return nil
	}

	a = &agent{
		ws:           ws,
		sock:         s.listener,
		addr:         addr,
		running:      true,
		udpData:      make(chan []byte, 10000),
		lastActivity: time.Now(),
	}
	go a.ws2sock()
	go a.sock2ws()
	s.agents[addr.String()] = a

	if logNewConnections {
		logrus.WithField("remote", addr.String()).Info("New client")
	}
	return a
}

func (s *SocketServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.agents {
		a.running = false
		a.ws.Close()
		close(a.udpData)
	}
	return s.listener.Close()
}
