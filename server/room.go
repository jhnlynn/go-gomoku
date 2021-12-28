package server

import (
	"go-exp/common/message"
	"go-exp/features/game"
	"sync"
)

type Room struct {
	Clients map[*Client]bool
	ClientsMutex sync.Mutex
	Game *game.Game
	Register chan *Client
	Unregister chan *Client
	StartGame chan struct {}
	Broadcast chan *message.Message
	Rematch chan *Client
	rematchRequests map[*Client]bool
	rematchRequestsMutex sync.Mutex
}
