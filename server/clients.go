package server

import (
	"encoding/json"
	ws "github.com/gorilla/websocket"
	"go-exp/common/message"
	"go-exp/features/game"
	"go-exp/invitation/invitationCode"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2048
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
}

type Client struct {
	Conn         *ws.Conn
	disconnected bool
	Player       game.Player
	Server       *Server
	Room         *Room
	Code         int
	Send         chan []byte
}

func newClient(conn *ws.Conn, server *Server) *Client {
	return &Client{
		Conn:   conn,
		Server: server,
		Send:   make(chan []byte),
	}
}

func ServeWs(s *Server, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		return
	}

	path := r.URL.Path[len("ws"):]
	params := strings.Split(path, "/")
	client := newClient(conn, s)

	if len(params) < 2 || params[1] == "" {
		code, m := client.invite()
		if m != nil {
			client.error(m)
			return
		}
		data, _ = json.Marshal(message.NewInvitationCode(strconv.Itoa(code)))
		client.Conn.WriteMessage(ws.TextMessage, data)
	} else {
		m := client.accept(params[1])
		if m != nil {
			client.error(m)
			return
		}
	}

	go client.write()
	go client.read()
}

func (c *Client) read() {
	defer c.disconnect()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		c.handleMessage(data)
	}
}

func (c *Client) write() {
	defer c.disconnect()

	ticker := time.NewTimer(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case m, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}

			w.Write(m)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(data []byte) {
	if len(data) == 0 {
		return
	}

	m := &message.Message{}

	err := json.Unmarshal(data, m)

	if err != nil {
		m, _ := json.Marshal(message.NewInvalidMessageFormat())
		c.Conn.WriteMessage(ws.TextMessage, m)
		return
	}

	switch m.Type {
	case message.Chat:
		c.handleChatMessage(m)
	case message.Move:
		c.handleMoveMessage(m)
	case message.NextGame:
		c.handleNextGame()
	}
}

func (c *Client) disconnect() {
	if c.disconnected {
		return
	}

	c.disconnected = true

	c.Conn.Close()

	room := c.Room
	if room == nil {
		return
	}

	room.Unregister <- c
	room.ClientsMutex.Lock()
	defer room.ClientsMutex.Unlock()

	for client := range room.Clients {
		m, _ := json.Marshal(message.NewOpponentLeft())
		client.Conn.WriteMessage(ws.TextMessage, m)
	}

	c.Server.InvitationsMutex.Lock()
	defer c.Server.InvitationsMutex.Unlock()

	if _, ok := c.Server.Invitation[c.Code]; ok {
		delete(c.Server.Invitation, c.Code)
		invitationCode.Return(c.Code)
	}
}

func (c *Client) error(m *message.Message) {
	data, _ := json.Marshal(m)
	c.Conn.WriteMessage(ws.TextMessage, data)
}

func (c *Client) invite() (int, *message.Message) {
	code, err := invitationCode.Get()
	if err != nil {
		return 0, message.NewInsufficientInvitationCode()
	}

	c.Code = code
	c.Server.Invite <- c
	return code, nil
}

func (c *Client) accept(codeString string) *message.Message {
	code, err := strconv.Atoi(codeString)
	if err != nil || code < 0 || code > invitationCode.GetMaxSeed() {
		return message.NewInvalidInvitationCode(codeString)
	}

	c.Server.InvitationsMutex.Lock()
	defer c.Server.InvitationsMutex.Unlock()

	_, ok := c.Server.Invitation[code]
	if !ok {
		return message.NewInvalidInvitationCode(codeString)
	}

	c.Code = code
	c.Server.Accept <- c
	invitationCode.Return(code)
	return nil
}
