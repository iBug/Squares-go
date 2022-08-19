package squares

const (
	NSHAPES    = 21 // number of shapes
	NROTATIONS = 8  // number of rotations
)

type Shape struct {
	Grids          []Coord
	Width, Height  int
	MirrorSymmetry bool
	RotateSymmetry bool // 180deg rotational symmetry
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
