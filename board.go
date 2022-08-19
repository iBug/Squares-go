package squares

const (
	BOARD_SIZE   = 21
	BOARD_WIDTH  = BOARD_SIZE
	BOARD_HEIGHT = BOARD_SIZE
	GRID_SIZE    = 5

	NPLAYERS   = 4  // max players
	NSHAPES    = 21 // number of shapes
	NROTATIONS = 8  // number of rotations
)

type Shape struct {
	Grids          []Coord
	Width, Height  int
	MirrorSymmetry bool
	RotateSymmetry bool // 180deg rotational symmetry
}

type Game struct {
	// Board[y][x] is the grid at Coord(x, y)
	Board     [BOARD_HEIGHT][BOARD_WIDTH]int `json:"board"`
	ChessUsed [NPLAYERS][NSHAPES]bool        `json:"chess_used"`
}

var gameShapes = []Shape{
	{[]Coord{{0, 0}}, 1, 1, true, true}, // id = 0
	{[]Coord{{0, 0}, {1, 0}}, 2, 1, true, true},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}}, 3, 1, true, true},
	{[]Coord{{0, 0}, {1, 0}, {0, 1}}, 2, 2, true, false},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, 4, 1, true, true},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {0, 1}}, 3, 2, false, false},
	{[]Coord{{0, 0}, {1, 0}, {1, 1}, {2, 1}}, 3, 2, false, true},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {1, 1}}, 3, 2, true, false},
	{[]Coord{{0, 0}, {1, 0}, {0, 1}, {1, 1}}, 2, 2, true, true}, // id = 8
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}}, 5, 1, true, true},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {0, 1}}, 4, 2, false, false},
	{[]Coord{{0, 0}, {1, 0}, {1, 1}, {2, 1}, {3, 1}}, 4, 2, false, false},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {1, 1}}, 4, 2, false, false},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {0, 2}}, 3, 3, true, false},
	{[]Coord{{0, 0}, {1, 0}, {1, 1}, {1, 2}, {2, 2}}, 3, 3, false, true},
	{[]Coord{{0, 0}, {0, 1}, {1, 1}, {2, 1}, {1, 2}}, 3, 3, false, false},
	{[]Coord{{1, 0}, {0, 1}, {1, 1}, {2, 1}, {1, 2}}, 3, 3, true, true}, // id = 16
	{[]Coord{{0, 0}, {1, 0}, {1, 1}, {2, 1}, {2, 2}}, 3, 3, true, false},
	{[]Coord{{0, 0}, {0, 1}, {1, 1}, {2, 1}, {0, 2}}, 3, 3, true, false},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {1, 1}}, 3, 2, false, false},
	{[]Coord{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {2, 1}}, 3, 2, true, false},
}

/*********
 * Shape *
 *********/

func GetShape(num, rotation int) Shape {
	return gameShapes[num].Rotate(rotation)
}

func (s Shape) Rotate(rotation int) Shape {
	res := Shape{Grids: make([]Coord, len(s.Grids))}
	if rotation%2 == 0 {
		res.Width, res.Height = s.Width, s.Height
	} else {
		res.Width, res.Height = s.Height, s.Width
	}
	for i := 0; i < len(s.Grids); i++ {
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
		res.Grids[i] = dst
	}
	return res
}

func GetNextRotation(shapeId, rotation int) int {
	switch shapeId {
	case 0, 8, 16:
		return 0
	}
	rotation++
	if gameShapes[shapeId].RotateSymmetry && rotation&2 != 0 {
		rotation += 2
	}
	if gameShapes[shapeId].MirrorSymmetry {
		rotation %= 4
	} else {
		rotation %= 8
	}
	return rotation
}

func GetPrevRotation(shapeId, rotation int) int {
	switch shapeId {
	case 0, 8, 16:
		return 0
	}
	rotation = (rotation + 7) % 8
	if gameShapes[shapeId].RotateSymmetry {
		rotation &= ^2
	}
	if gameShapes[shapeId].MirrorSymmetry {
		rotation &= ^4
	}
	return rotation
}

func AvailableRotations(shapeId int) int {
	switch shapeId {
	case 0, 8, 16:
		return 1
	}
	rotations := 3
	if !gameShapes[shapeId].RotateSymmetry {
		rotations |= rotations << 2
	}
	if !gameShapes[shapeId].MirrorSymmetry {
		rotations |= rotations << 4
	}
	return rotations
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

func (game *Game) Insert(shapeId, rotation int, pos Coord, playerId int, firstRound bool) {
	shape := GetShape(shapeId, rotation)
	for _, grid := range shape.Grids {
		grid = grid.Add(pos)
		game.Board[grid.Y][grid.X] = playerId
	}
	game.ChessUsed[playerId][shapeId] = true
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
