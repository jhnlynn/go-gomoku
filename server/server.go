package server

import "sync"

type Server struct {
	Invitation map[int]*Room
	InvitationsMutex sync.Mutex
	Invite chan *Client
	Accept chan *Client
}

func New() *Server {
	return &Server {

	}
}

func Run(s *Server) {
	for {
		select {
			case c := <-
		}
	}
}

