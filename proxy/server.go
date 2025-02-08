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

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			s.mu.Lock()
			for addr, a := range s.agents {
				a.mu.Lock()
				if !a.running || time.Since(a.lastActivity) > 5*time.Minute {
					a.running = false
					if a.ws != nil {
						a.ws.Close()
					}
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
			logrus.WithField("err", err).Error("Could not ReadFromUDP() a connection")
			continue
		}

		data := make([]byte, n)
		copy(data, p[:n])

		a := s.getAgent(addr)
		if a == nil { // Prevent panic if getAgent() returns nil
			logrus.WithField("remote", addr.String()).Error("Agent creation failed, dropping packet")
			continue
		}

		// Send data to agent
		select {
		case a.udpData <- data:
		default:
			logrus.WithField("remote", addr.String()).Warn("UDP queue full, dropping packet")
		}
	}
}

func (s *SocketServer) getAgent(addr *net.UDPAddr) *agent {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if agent already exists
	a, ok := s.agents[addr.String()]
	if ok && a.running {
		return a
	}

	// Construct WebSocket URL
	u := url.URL{Scheme: "ws", Host: s.Destination, Path: "/"}
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	logrus.WithFields(logrus.Fields{
		"remote": addr.String(),
		"ws_url": u.String(),
	}).Info("Attempting WebSocket connection")

	// Try connecting with retries
	var ws *websocket.Conn
	var err error
	for i := 0; i < 5; i++ {
		ws, _, err = dialer.Dial(u.String(), nil)
		if err == nil {
			break
		}
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"remote": addr.String(),
			"retry":  i + 1,
		}).Warn("Could not dial WebSocket, retrying...")
		time.Sleep(3 * time.Second)
	}

	// If all attempts fail, return nil (fixes panic)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":    err,
			"remote": addr.String(),
		}).Error("WebSocket connection failed after retries")
		return nil
	}

	// Create a new agent
	a = &agent{
		ws:           ws,
		sock:         s.listener,
		addr:         addr,
		running:      true,
		udpData:      make(chan []byte, 10000),
		lastActivity: time.Now(),
	}

	// Start the WebSocket <-> UDP processing
	go a.ws2sock()
	go a.sock2ws()
	s.agents[addr.String()] = a

	if logNewConnections {
		logrus.WithField("remote", addr.String()).Info("New client connected")
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
