package game

import "errors"

const Size = 15

/*
	Player
	corresponds with Black & White, amazing choice
*/
const (
	None = iota
	Black
	White
	Draw
)

type Player uint8

type Board [Size][Size]Player

/*
	Piece
	includes:
	1. which player place it,
	2. col & row
*/
type Piece struct {
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	Player Player `json:"player"`
}

/*
	Game
	Board: the chess board,
	Player: who should place the next chess,
	LastMove:
	WinningPieces: the pieces of which makes the winner win
*/
type Game struct {
	Board         Board    `json:"board"`
	Player        Player   `json:"player"`
	Winner        Player   `json:"winner,omitempty"`
	LastMove      *Piece   `json:"lastMove,omitempty"`
	WinningPieces []*Piece `json:"winningPieces,omitempty"`
}

/*
	New:
	Initialize all grids as 0, which is neither Black nor White, amazing choice
	Let the game begin.
*/
func New() *Game {
	board := [Size][Size]Player {}

	for i := range board {
		board[i] = [Size]Player {}
	}

	return &Game {
		Board: board,
		Player: Black,
	}
}

func (g *Game) Move(p *Piece) error {
	if g.Winner != None {
		return errors.New("the game has ended")
	}

	r, c := p.Row, p.Col

	if r < 0 || r >= Size ||
		c < 0 || c > Size ||
		g.Board[r][c] != 0 ||
		p.Player != g.Player {
		return errors.New("invalid position")
	}

	g.LastMove = p
	// turn around to play
	player := g.Player
	if player == Black {
		g.Player = White
	} else {
		g.Player = Black
	}
	g.Board[r][c] = player
	g.checkWinning(p)

	if g.Winner == None {
		g.testDraw()
	}

	return nil
}

func (g *Game) testDraw() {
	for i := range g.Board {
		for j := range g.Board[i] {
			if g.Board[i][j] == None {
				return
			}
		}
	}

	g.Winner = Draw
}

func (g *Game) checkWinning(p *Piece) {
	var beg, end int

	// # check the vertical #
	for beg = p.Row; beg >= 0; beg -= 1 {
		if g.Board[beg][p.Col] != p.Player {
			break
		}
	}

	for end = p.Row; end < Size; end += 1 {
		if g.Board[end][p.Col] != p.Player {
			break
		}
	}

	if end - beg - 1 >= 5 {
		g.Winner = p.Player
		for i := beg + 1; i < end; i += 1 {
			// lock the seres of the winning pieces in the board
			g.WinningPieces = append(g.WinningPieces, &Piece{Row: i, Col: p.Col, Player: p.Player})
		}
	}

	// # end region

	// # check the horizontal #
	for beg = p.Col; beg >= 0; beg -= 1 {
		if g.Board[p.Row][beg] != p.Player {
			break
		}
	}

	for end = p.Col; end < Size; end += 1 {
		if g.Board[p.Row][end] != p.Player {
			break
		}
	}

	if end - beg >= 5 {
		for i := beg + 1; i < end; i += 1 {
			g.Winner = p.Player
			g.WinningPieces = append(g.WinningPieces, &Piece{Row: p.Row, Col: i, Player: p.Player})
		}
	}

	// # end region

	var begR, begC, endR, endC int

	// # check forward diagonal #
	for begR, begC = p.Row, p.Col; begR >= 0 && begC < Size; begR, begC = begR-1, begC+1 {
		if g.Board[begR][begC] != p.Player {
			break
		}
	}
	for endR, endC = p.Row, p.Col; endR < Size && endC >= 0; endR, endC = endR+1, endC-1 {
		if g.Board[endR][endC] != p.Player {
			break
		}
	}
	if endR-begR-1 >= 5 {
		for i, j := begR+1, begC-1; i < endR && j > endC; i, j = i+1, j-1 {
			g.Winner = p.Player
			g.WinningPieces = append(g.WinningPieces, &Piece{Row: i, Col: j, Player: p.Player})
		}
	}

	// end region

	// # check back diagonal #
	for begR, begC = p.Row, p.Col; begR >= 0 && begC >= 0; begR, begC = begR-1, begC-1 {
		if g.Board[begR][begC] != p.Player {
			break
		}
	}
	for endR, endC = p.Row, p.Col; endR < Size && endC < Size; endR, endC = endR+1, endC+1 {
		if g.Board[endR][endC] != p.Player {
			break
		}
	}
	if endR-begR-1 >= 5 {
		for i, j := begR+1, begC+1; i < endR && j < endC; i, j = i+1, j+1 {
			g.Winner = p.Player
			g.WinningPieces = append(g.WinningPieces, &Piece{Row: i, Col: j, Player: p.Player})
		}
	}
}