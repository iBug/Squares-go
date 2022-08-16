package squares

type Coord struct {
	X, Y int
}

var (
	CORNERS = []Coord{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	EDGES   = []Coord{{1, 0}, {0, -1}, {-1, 0}, {0, 1}}
)

// Negate the coordinate
func (c Coord) Neg() Coord {
	return Coord{-c.X, -c.Y}
}

// Add the coordinates
func (c Coord) Add(c2 Coord) Coord {
	return Coord{c.X + c2.X, c.Y + c2.Y}
}

// Add to the coordinate by values
func (c Coord) AddXY(x, y int) Coord {
	return Coord{c.X + x, c.Y + y}
}

// Subtract the coordinates
func (c Coord) Sub(c2 Coord) Coord {
	return c.Add(c2.Neg())
}

// Subtract the coordinate by values
func (c Coord) SubXY(x, y int) Coord {
	return Coord{c.X - x, c.Y - y}
}

// Check if the coordinates are equal
func (c Coord) Equals(c2 Coord) bool {
	return c.X == c2.X && c.Y == c2.Y
}

// Transpose the coordinate
func (c Coord) T() Coord {
	return Coord{c.Y, c.X}
}
