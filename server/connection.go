package server

import (
	"net/http"
	"time"

	"github.com/dev-appmonsters/dicemix-light-server/dc"
	"github.com/dev-appmonsters/dicemix-light-server/utils"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// using expose interfaces
var iDcNet dc.DC

type connection struct {
	hub *hub
	Server
}

// NewConnection creates a new Server instance
func NewConnection() Server {
	iDcNet = dc.NewDCNetwork()

	hub := newHub()
	go hub.listener()

	return &connection{hub: hub}
}

// Register handles websocket requests from the peer.
func (s *connection) Register(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error:- ", err)
		return
	}
	client := &client{hub: s.hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writeMessage()
	go client.readMessage()
}

// readMessage pumps messages from the websocket connection to the hub.
//
// The application runs readMessage in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *client) readMessage() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadDeadline(time.Now().Add(utils.PongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(utils.PongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("Error:- ", err)
			}
			break
		}
		c.hub.request <- message
	}
}

// writeMessage pumps messages from the hub to the websocket connection.
//
// A goroutine running writeMessage is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *client) writeMessage() {
	ticker := time.NewTicker(utils.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(utils.WriteWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(utils.Newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(utils.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// removs all peers from run and terminates run
func terminate(h *hub, sessionID uint64) {
	// if session exists
	if _, ok := h.runs[sessionID]; !ok {
		return
	}

	// remove all peers from run
	for _, peer := range h.runs[sessionID].peers {
		removePeer(h, peer.Id)
	}

	// remove run info
	delete(h.runs, sessionID)
}

// remove a peer from set of all peers
func removePeer(h *hub, id int32) {
	// if client is offline and not submitted response
	if client, ok := getClient(h.clients, id); ok {
		// remove offline peers from clients
		log.Info("USER UN-REGISTRATION - ", id)
		delete(h.clients, client)
		close(client.send)
	}
}

// checks for any potential errors
func checkError(err error) bool {
	if err != nil {
		log.Error("Error Occured:", err)
		return true
	}
	return false
}
