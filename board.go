package squares

const (
	BOARD_SIZE   = 21
	BOARD_WIDTH  = BOARD_SIZE
	BOARD_HEIGHT = BOARD_SIZE
	GRID_SIZE    = 5

	NPLAYERS   = 4  // max players
	NCHESS     = 21 // numper of chess pieces
	NROTATIONS = 8  // number of possible rotations
)

type Shape struct {
	Grids         []Coord
	Width, Height int
}

type Game struct {
	// board[y][x] is the grid at Coord(x, y)
	board     [BOARD_HEIGHT][BOARD_WIDTH]int
	chessUsed [NPLAYERS][NCHESS]bool
}

var SHAPES = []Shape{
	{[]Coord{{0, 0}}, 1, 1},
	{[]Coord{{0, 0}, {1, 0}}, 2, 1},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}}, 3, 1},
	{[]Coord{{0, 0}, {1, 0}, {0, 1}}, 2, 2},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, 4, 1},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {0, 1}}, 3, 2},
	{[]Coord{{0, 0}, {1, 0}, {1, 1}, {2, 1}}, 3, 2},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {1, 1}}, 3, 2},
	{[]Coord{{0, 0}, {1, 0}, {0, 1}, {1, 1}}, 2, 2},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}}, 5, 1},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {0, 1}}, 4, 2},
	{[]Coord{{0, 0}, {1, 0}, {1, 1}, {2, 1}, {3, 1}}, 4, 2},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {1, 1}}, 4, 2},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {0, 2}}, 3, 3},
	{[]Coord{{0, 0}, {1, 0}, {1, 1}, {1, 2}, {2, 2}}, 3, 3},
	{[]Coord{{0, 0}, {0, 1}, {1, 1}, {2, 1}, {1, 2}}, 3, 3},
	{[]Coord{{1, 0}, {0, 1}, {1, 1}, {2, 1}, {1, 2}}, 3, 3},
	{[]Coord{{0, 0}, {1, 0}, {1, 1}, {2, 1}, {2, 2}}, 3, 3},
	{[]Coord{{0, 0}, {0, 1}, {1, 1}, {2, 1}, {0, 2}}, 3, 3},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {1, 1}}, 3, 2},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {2, 1}}, 3, 2},
}

/*********
 * Shape *
 *********/

func GetShape(num, rotation int) Shape {
	return SHAPES[num].Rotate(rotation)
}

func (s Shape) Size() int {
	return len(s.Grids)
}

func (s Shape) Rotate(rotation int) Shape {
	var res Shape
	if rotation%2 == 0 {
		res.Width = s.Width
		res.Height = s.Height
	} else {
		res.Width = s.Height
		res.Height = s.Width
	}
	for i := 0; i < s.Size(); i++ {
		src := s.Grids[i]
		var dst Coord
		switch rotation % NROTATIONS {
		default:
			dst = src
		case 1:
			dst.X = res.Width - src.Y - 1
			dst.Y = src.X
		case 2:
			dst.X = res.Width - src.X - 1
			dst.Y = res.Height - src.Y - 1
		case 3:
			dst.X = src.Y
			dst.Y = res.Height - src.X - 1
		case 4: // mirrored
			dst.X = res.Width - src.X - 1
			dst.Y = src.Y
		case 5:
			dst.X = res.Width - src.Y - 1
			dst.Y = res.Height - src.X - 1
		case 6:
			dst.X = src.X
			dst.Y = res.Height - src.Y - 1
		case 7:
			dst.X = src.Y
			dst.Y = src.X
		}
		res.Grids = append(res.Grids, dst)
	}
	return res
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
	return game.board[y][x]
}

func (game *Game) GetUsed(cmnum, playerId int) bool {
	return game.chessUsed[playerId][cmnum]
}

// was Squares::init in the original C++ version
func (game *Game) Reset() {
	for i := 0; i < BOARD_HEIGHT; i++ {
		for j := 0; j < BOARD_WIDTH; j++ {
			game.board[i][j] = -1
		}
	}
	for p := 0; p < len(game.chessUsed); p++ {
		for i := 0; i < len(game.chessUsed[p]); i++ {
			game.chessUsed[p][i] = false
		}
	}
}

func (game *Game) TryInsert(cmnum, rotation int, coord Coord, np int, firstRound bool) bool {
	if game.chessUsed[np][cmnum] {
		return false
	}
	canPlace := false
	shape := SHAPES[cmnum].Rotate(rotation)

	if firstRound {
		corner := Coord{}
		switch np {
		case 1:
			corner.X = BOARD_WIDTH - 1
		case 3:
			corner.Y = BOARD_WIDTH - 1
		case 2:
			corner.X = BOARD_WIDTH - 1
			corner.Y = BOARD_WIDTH - 1
		}
		for _, grid := range shape.Grids {
			pos := coord.Add(grid)
			if !InRange(pos) {
				return false
			}
			if pos.Equals(corner) {
				canPlace = true
			}
		}
	} else {
		for i := range shape.Grids {
			shape.Grids[i] = shape.Grids[i].Add(coord)
			if !InRange(shape.Grids[i]) {
				// out of bounds
				return false
			}
			if game.board[shape.Grids[i].Y][shape.Grids[i].X] > 0 {
				// already occupied
				return false
			}
		}
		for _, grid := range shape.Grids {
			// Edge rule: No adjacent pieces from the same player
			for _, edge := range EDGES {
				check := grid.Add(edge)
				if InRange(check) && game.board[check.Y][check.X] == np {
					return false
				}
			}
			// Corner rule: One corner must be from the same player
			for _, corner := range CORNERS {
				check := grid.Add(corner)
				if InRange(check) && game.board[check.Y][check.X] == np {
					canPlace = true
					break
				}
			}
		}
	}
	return canPlace
}

func (game *Game) Insert(cmnum, rotation int, coord Coord, np int, firstRound bool) {
	shape := SHAPES[cmnum].Rotate(rotation)
	for _, grid := range shape.Grids {
		grid = grid.Add(coord)
		game.board[grid.Y][grid.X] = np
	}
	game.chessUsed[np][cmnum] = true
}

// Check if a player has any valid move available
func (game *Game) CheckPlayer(np int) bool {
	/*Board status
	 *  0: This cell is not adjacent to any existing pieces
	 *  1: This cell is diagonally adjacent to an existing piece
	 *  2: This cell is orthogonally adjacent to, or already covered by, an existing piece
	 */
	var boardStatus [BOARD_HEIGHT][BOARD_WIDTH]int
	for y := 0; y < BOARD_HEIGHT; y++ {
		for x := 0; x < BOARD_WIDTH; x++ {
			// Mark occupied cells
			if game.board[y][x] >= 0 {
				boardStatus[y][x] = 2
			}
			// Calculate adjacency only for same-color cells
			if game.board[y][x] != np {
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
	for i := 0; i < NCHESS; i++ {
		if game.chessUsed[np][i] {
			continue
		}
		for _, must := range musts {
			for rotation := 0; rotation < NROTATIONS; rotation++ {
				shape := GetShape(i, rotation)
				for _, grid := range shape.Grids {
					pos := must.Sub(grid)
					if InRange(pos) && game.TryInsert(i, rotation, pos, np, false) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (game *Game) GetLostPlayers() int {
	ret := 0
	for i := 0; i < NPLAYERS; i++ {
		if !game.CheckPlayer(i) {
			ret |= 1 << i
		}
	}
	return ret
}

/**********************************
 * Game-related utility functions *
 **********************************/

func InRange(c Coord) bool {
	return c.X >= 0 && c.X < BOARD_WIDTH && c.Y >= 0 && c.Y < BOARD_HEIGHT
}
