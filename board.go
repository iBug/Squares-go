package squares

import "fmt"

const (
	BOARD_SIZE   = 21
	BOARD_WIDTH  = BOARD_SIZE
	BOARD_HEIGHT = BOARD_SIZE
	GRID_SIZE    = 5

	NPLAYERS = 4 // max players
)

// The state of a game board
type Game struct {
	// Board[y][x] is the grid at Coord(x, y)
	Board     [squares.BOARD_HEIGHT][squares.BOARD_WIDTH]int `json:"board"`
	ChessUsed [squares.NPLAYERS][squares.NSHAPES]bool        `json:"chess_used"`

	ActivePlayer int  `json:"active_player"` // Who's next, -1 = game over
	FirstRound   bool `json:"first_round"`
	LostPlayers  int  `json:"lost_players"` // Cached value from GetLostPlayers()
}

/********
 * Game *
 ********/

func NewGame() *Game {
	game := &Game{}
	game.Reset()
	return game
}

func (game *Game) At(x, y int) int {
	return game.Board[y][x]
}

func (game *Game) GetUsed(playerId, shapeId int) bool {
	return game.ChessUsed[playerId][shapeId]
}

// was Squares::init in the original C++ version
func (game *Game) Reset() {
	game.ActivePlayer = 0
	game.FirstRound = true
	game.LostPlayers = 0

	for i := 0; i < BOARD_HEIGHT; i++ {
		for j := 0; j < BOARD_WIDTH; j++ {
			game.Board[i][j] = -1
		}
	}
	for p := 0; p < len(game.ChessUsed); p++ {
		for i := 0; i < len(game.ChessUsed[p]); i++ {
			game.ChessUsed[p][i] = false
		}
	}
}

func (game *Game) TryInsert(shapeId, rotation int, pos Coord, playerId int, firstRound bool) bool {
	if game.ChessUsed[playerId][shapeId] {
		return false
	}
	canPlace := false
	shape := GetShape(shapeId, rotation)

	if firstRound {
		corner := Coord{0, 0}
		switch playerId {
		case 1:
			corner.X = BOARD_WIDTH - 1
		case 3:
			corner.Y = BOARD_HEIGHT - 1
		case 2:
			corner.X = BOARD_WIDTH - 1
			corner.Y = BOARD_HEIGHT - 1
		}
		for _, grid := range shape.Grids {
			pos := pos.Add(grid)
			if !InRange(pos) {
				return false
			}
			if pos.Equals(corner) {
				canPlace = true
			}
		}
	} else {
		for i := range shape.Grids {
			shape.Grids[i] = shape.Grids[i].Add(pos)
			if !InRange(shape.Grids[i]) {
				// out of bounds
				return false
			}
			if game.Board[shape.Grids[i].Y][shape.Grids[i].X] > 0 {
				// already occupied
				return false
			}
		}
		for _, grid := range shape.Grids {
			// Edge rule: No adjacent pieces from the same player
			for _, edge := range EDGES {
				check := grid.Add(edge)
				if InRange(check) && game.Board[check.Y][check.X] == playerId {
					return false
				}
			}
			// Corner rule: One corner must be from the same player
			for _, corner := range CORNERS {
				check := grid.Add(corner)
				if InRange(check) && game.Board[check.Y][check.X] == playerId {
					canPlace = true
					break
				}
			}
		}
	}
	return canPlace
}

func (game *Game) Insert(shapeId, rotation int, pos Coord, playerId int) {
	shape := GetShape(shapeId, rotation)
	for _, grid := range shape.Grids {
		grid = grid.Add(pos)
		game.Board[grid.Y][grid.X] = playerId
	}
	game.ChessUsed[playerId][shapeId] = true

	game.LostPlayers = -1 // so it's calculated the next time GetLostPlayers() is called
}

func (game *Game) AfterMove() bool {
	if !game.FirstRound {
		if lp := game.GetLostPlayers(); lp != 0 {
			for i := 0; i < NPLAYERS; i++ {
				if lp&(1<<i) != 0 {
					if !game.IsAnyPlayerAlive() {
						fmt.Printf("Player %d won!\n", i+1)
					} else {
						fmt.Printf("Player %d lost!\n", i+1)
					}
				}
			}
		}
	} else if game.ActivePlayer == NPLAYERS-1 {
		game.FirstRound = false
	}

	activePlayer := (game.ActivePlayer + 1) % NPLAYERS
	if game.GetLostPlayers() != 0 {
		if !game.IsAnyPlayerAlive() {
			game.ActivePlayer = -1
			return false
		}
		for _ = activePlayer; game.LostPlayers&(1<<activePlayer) != 0; activePlayer = (activePlayer + 1) % NPLAYERS {
		}
	}
	game.ActivePlayer = activePlayer
	return true
}

// Check if a player has any valid move available
func (game *Game) CheckPlayer(playerId int) bool {
	/*Board status
	 * 0: This cell is not adjacent to any existing pieces
	 * 1: This cell is diagonally adjacent to an existing piece
	 * 2: This cell is orthogonally adjacent to, or already covered by, an existing piece
	 */
	var boardStatus [BOARD_HEIGHT][BOARD_WIDTH]int
	for y := 0; y < BOARD_HEIGHT; y++ {
		for x := 0; x < BOARD_WIDTH; x++ {
			// Mark occupied cells
			if game.Board[y][x] >= 0 {
				boardStatus[y][x] = 2
			}
			// Calculate adjacency only for same-color cells
			if game.Board[y][x] != playerId {
				continue
			}
			for _, edge := range EDGES {
				check := edge.AddXY(x, y)
				if InRange(check) {
					boardStatus[check.Y][check.X] = 2
				}
			}
			for _, corner := range CORNERS {
				check := corner.AddXY(x, y)
				if InRange(check) && boardStatus[check.Y][check.X] < 1 {
					boardStatus[check.Y][check.X] = 1
				}
			}
		}
	}

	musts := make([]Coord, 0)
	for y := 0; y < BOARD_HEIGHT; y++ {
		for x := 0; x < BOARD_WIDTH; x++ {
			if boardStatus[y][x] == 1 {
				musts = append(musts, Coord{x, y})
			}
		}
	}

	// Enumerate all remaining pieces over all "must cover" cells
	for i := 0; i < NSHAPES; i++ {
		if game.ChessUsed[playerId][i] {
			continue
		}
		for _, must := range musts {
			for rotation := 0; rotation < NROTATIONS; rotation++ {
				shape := GetShape(i, rotation)
				for _, grid := range shape.Grids {
					pos := must.Sub(grid)
					if InRange(pos) && game.TryInsert(i, rotation, pos, playerId, false) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (game *Game) FindLostPlayers() int {
	if game.FirstRound {
		return 0
	}

	ret := 0
	for i := 0; i < NPLAYERS; i++ {
		if !game.CheckPlayer(i) {
			ret |= 1 << i
		}
	}
	return ret
}

func (game *Game) GetLostPlayers() int {
	if game.LostPlayers >= 0 {
		return game.LostPlayers
	}
	game.LostPlayers = game.FindLostPlayers()
	return game.LostPlayers
}

func (game *Game) IsAnyPlayerAlive() bool {
	return game.GetLostPlayers() != (1<<NPLAYERS)-1
}

/**********************************
 * Game-related utility functions *
 **********************************/

func InRange(c Coord) bool {
	return c.X >= 0 && c.X < BOARD_WIDTH && c.Y >= 0 && c.Y < BOARD_HEIGHT
}
