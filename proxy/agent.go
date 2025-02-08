package proxy

import (
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// var (
// 	logExchanges      = true
// 	hexdumpPackets    = true
// 	logNewConnections = true
// )

type agent struct {
	sock         *net.UDPConn
	ws           *websocket.Conn
	addr         *net.UDPAddr
	running      bool
	udpData      chan []byte
	lastActivity time.Time
	mu           sync.Mutex
}

func (a *agent) ws2sock() {
	log := logrus.WithFields(logrus.Fields{
		"remote": a.addr.String(),
		"func":   "ws2sock",
	})
	defer func() {
		a.mu.Lock()
		a.running = false
		a.ws.Close()
		close(a.udpData)
		a.mu.Unlock()
	}()

	for {
		t, p, err := a.ws.ReadMessage()
		if err != nil {
			log.WithField("err", err).Error("Could not read WebSocket message")
			return
		}

		if logExchanges {
			log.Info("Got WS message")
		}

		if t != websocket.BinaryMessage {
			log.Error("Wrong message type")
			return
		}

		n, err := a.sock.WriteTo(p, a.addr)
		if err != nil {
			log.WithError(err).Error("Could not write to socket")
			return
		}

		if n != len(p) {
			log.WithFields(logrus.Fields{
				"written": n,
				"length":  len(p),
			}).Error("Did not write full message")
			return
		}

		if logExchanges {
			log.Info("Successfully sent WS message to socket")
		}

		if hexdumpPackets {
			fmt.Println(hex.Dump(p))
		}

		a.mu.Lock()
		a.lastActivity = time.Now()
		a.mu.Unlock()
	}
}

func (a *agent) sock2ws() {
	log := logrus.WithFields(logrus.Fields{
		"remote": a.addr.String(),
		"func":   "sock2ws",
	})

	defer func() {
		a.mu.Lock()
		a.running = false
		a.ws.Close()
		close(a.udpData)
		a.mu.Unlock()
	}()

	for p := range a.udpData {
		err := a.ws.WriteMessage(websocket.BinaryMessage, p)
		if err != nil {
			log.WithField("err", err).Error("Could not WriteMessage()")
			return
		}

		if logExchanges {
			log.Info("Successfully sent UDP message to WebSocket")
		}

		if hexdumpPackets {
			fmt.Println(hex.Dump(p))
		}

		a.mu.Lock()
		a.lastActivity = time.Now()
		a.mu.Unlock()
	}
}
