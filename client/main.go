package main

import (
	"flag"
	"fmt"
	"math"

	squares "github.com/iBug/Squares-go"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	GRID_WIDTH         = squares.BOARD_WIDTH
	GRID_HEIGHT        = squares.BOARD_HEIGHT
	GRID_CELL_SIZE     = 36
	SELECTOR_WIDTH     = 16
	SELECTOR_HEIGHT    = squares.BOARD_HEIGHT
	SELECTOR_CELL_SIZE = 24
	ROTATOR_WIDTH      = 6
	ROTATOR_HEIGHT     = 4
	ROTATOR_CELL_SIZE  = 16
	FPS_LIMIT          = 60
	SDL_TICKSPEED      = 1000
	INTERVAL           = float64(SDL_TICKSPEED) / float64(FPS_LIMIT)

	BOARD_AREA_WIDTH     = GRID_WIDTH*GRID_CELL_SIZE + 1
	SELECTOR_AREA_HEIGHT = SELECTOR_HEIGHT * SELECTOR_CELL_SIZE
	WINDOW_WIDTH         = BOARD_AREA_WIDTH + SELECTOR_WIDTH*SELECTOR_CELL_SIZE
	WINDOW_HEIGHT        = GRID_HEIGHT*GRID_CELL_SIZE + 1
)

var (
	GRID_CURSOR_COLORS       = []sdl.Color{{0xB2, 0x1B, 0x1A, 255}, {0x2C, 0xAA, 0x18, 255}, {0x1E, 0x28, 0xA4, 255}, {0xF9, 0xC8, 0x02, 255}}
	GRID_CURSOR_GHOST_COLORS = []sdl.Color{{0xF7, 0x4F, 0x52, 127}, {0xB3, 0xFB, 0x3B, 127}, {0x89, 0xEC, 0xFC, 127}, {0xFC, 0xE5, 0x4F, 127}}

	GRID_BACKGROUND  = sdl.Color{233, 233, 233, 255}
	GRID_LINE_COLOR  = sdl.Color{200, 200, 200, 255}
	GRID_WRONG_COLOR = sdl.Color{127, 127, 127, 255}

	SELECTOR_POS = []squares.Coord{{1, 1}, {1, 3}, {1, 5}, {1, 7}, {1, 10}, {1, 12}, {1, 15}, {1, 18}, {6, 1}, {10, 1}, {6, 4}, {6, 7}, {6, 10}, {6, 13}, {6, 17}, {12, 3}, {12, 10}, {12, 14}, {9, 15}, {12, 18}, {12, 7}}
)

var (
	game *squares.Game

	clientId     = 0
	activePlayer = 0
	firstRound   = true
	lostPlayers  = 0

	fServerAddr       = ""
	fLocalMultiplayer = false
)

func parseFlags() {
	flag.StringVar(&fServerAddr, "a", "", "(unsupported)")
	flag.Parse()

	// fLocalMultiplayer = fServerAddr == ""
	fLocalMultiplayer = true
}

func timeLeft(nextTime uint64) uint32 {
	now := sdl.GetTicks64()
	if now < nextTime {
		return uint32(nextTime - now)
	}
	return 0
}

func getSelection(x, y int) int {
	for i := 0; i < squares.NSHAPES; i++ {
		startX := SELECTOR_POS[i].X*SELECTOR_CELL_SIZE + BOARD_AREA_WIDTH
		endX := startX + squares.GetShape(i, 0).Width*SELECTOR_CELL_SIZE
		startY := SELECTOR_POS[i].Y * SELECTOR_CELL_SIZE
		endY := startY + squares.GetShape(i, 0).Height*SELECTOR_CELL_SIZE
		if x >= startX && x <= endX && y >= startY && y <= endY {
			return i
		}
	}
	return -1
}

func getRotation(x, y, shapeId int) int {
	rotations := squares.AvailableRotations(shapeId)
	for i := 0; i < squares.NROTATIONS; i++ {
		if rotations&(1<<i) == 0 {
			continue
		}
		startX := (i%4*ROTATOR_WIDTH+2)*ROTATOR_CELL_SIZE + BOARD_AREA_WIDTH
		endX := startX + squares.GetShape(shapeId, i).Width*ROTATOR_CELL_SIZE
		startY := WINDOW_HEIGHT - ((2-i/4)*ROTATOR_WIDTH+1)*ROTATOR_CELL_SIZE
		endY := startY + squares.GetShape(shapeId, i).Height*ROTATOR_CELL_SIZE
		if x >= startX && x <= endX && y >= startY && y <= endY {
			return i
		}
	}
	return -1
}

func renderBoard(renderer *sdl.Renderer) {
	for i := 0; i < squares.BOARD_HEIGHT; i++ {
		for j := 0; j < squares.BOARD_WIDTH; j++ {
			if game.At(j, i) >= 0 {
				gridColor := GRID_CURSOR_COLORS[game.At(j, i)]
				renderer.SetDrawColor(gridColor.R, gridColor.G, gridColor.B, gridColor.A)
				rect := sdl.Rect{int32(j * GRID_CELL_SIZE), int32(i * GRID_CELL_SIZE), int32(GRID_CELL_SIZE), int32(GRID_CELL_SIZE)}
				renderer.FillRect(&rect)
			}
		}
	}
}

// Extracted common code from renderSelection and renderRotator
func setColorForShape(renderer *sdl.Renderer, shapeId, playerId, activePlayer int, useGhostColor bool) {
	var gridColor sdl.Color
	if game.GetUsed(playerId, shapeId) {
		gridColor = GRID_WRONG_COLOR
	} else if activePlayer == playerId && useGhostColor {
		gridColor = GRID_CURSOR_GHOST_COLORS[clientId]
	} else {
		gridColor = GRID_CURSOR_COLORS[clientId]
	}
	renderer.SetDrawColor(gridColor.R, gridColor.G, gridColor.B, gridColor.A)
}

func renderShape(renderer *sdl.Renderer, shapeId, rotation int, topleft sdl.Rect, width, height int) {
	shape := squares.GetShape(shapeId, rotation)
	for i := 0; i < len(shape.Grids); i++ {
		tmp := topleft
		tmp.X += int32(shape.Grids[i].X * width)
		tmp.Y += int32(shape.Grids[i].Y * height)
		renderer.FillRect(&tmp)
	}
}

func renderSelector(renderer *sdl.Renderer, clientId, shapeId int) {
	for i := 0; i < squares.NSHAPES; i++ {
		setColorForShape(renderer, i, clientId, activePlayer, i == shapeId)
		base := sdl.Rect{
			X: int32(SELECTOR_POS[i].X*SELECTOR_CELL_SIZE + BOARD_AREA_WIDTH),
			Y: int32(SELECTOR_POS[i].Y * SELECTOR_CELL_SIZE),
			W: SELECTOR_CELL_SIZE,
			H: SELECTOR_CELL_SIZE,
		}
		renderShape(renderer, i, 0, base, SELECTOR_CELL_SIZE, SELECTOR_CELL_SIZE)
	}
}

func renderRotator(renderer *sdl.Renderer, clientId, shapeId, rotation int) {
	rotations := squares.AvailableRotations(shapeId)
	for i := 0; i < squares.NROTATIONS; i++ {
		if rotations&(1<<i) == 0 {
			continue
		}
		setColorForShape(renderer, i, clientId, activePlayer, i == rotation)
		base := sdl.Rect{
			X: int32((i%4*ROTATOR_WIDTH+2)*ROTATOR_CELL_SIZE + BOARD_AREA_WIDTH),
			Y: int32(WINDOW_HEIGHT - ((2-i/4)*ROTATOR_WIDTH+1)*ROTATOR_CELL_SIZE),
			W: ROTATOR_CELL_SIZE,
			H: ROTATOR_CELL_SIZE,
		}
		renderShape(renderer, shapeId, i, base, ROTATOR_CELL_SIZE, ROTATOR_CELL_SIZE)
	}
}

// was render_ghost() in original C++ code
func shouldRenderGhost(topleft sdl.Rect, shapeId, rotation int) bool {
	x := int(topleft.X)+squares.GetShape(shapeId, rotation).Width*GRID_CELL_SIZE <= WINDOW_HEIGHT
	y := int(topleft.Y)+squares.GetShape(shapeId, rotation).Height*GRID_CELL_SIZE <= BOARD_AREA_WIDTH
	return x && y
}

func isAnyPlayerAlive() bool {
	return lostPlayers != (1<<squares.NPLAYERS)-1
}

func afterMove(isLocalMultiplayerMode bool) bool {
	if !isLocalMultiplayerMode {
		firstRound = false
		return true
	}

	// 4 players on one client
	if !firstRound {
		if lp := game.GetLostPlayers(); lp != 0 {
			for i := 0; i < squares.NPLAYERS; i++ {
				if lp&(1<<i) != 0 {
					if !isAnyPlayerAlive() {
						fmt.Printf("Player %d won!\n", i+1)
					} else {
						fmt.Printf("Player %d lost!\n", i+1)
					}
				}
			}
			lostPlayers = lp
		}
	} else {
		if clientId == squares.NPLAYERS-1 {
			firstRound = false
		}
	}

	clientId = (clientId + 1) % 4
	if lostPlayers != 0 {
		if !isAnyPlayerAlive() {
			activePlayer = -1
			return false
		}
		for _ = clientId; lostPlayers&(1<<clientId) != 0; clientId = (clientId + 1) % 4 {
		}
	}
	activePlayer = clientId
	return true
}

func clientMain() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, renderer, err := sdl.CreateWindowAndRenderer(WINDOW_WIDTH, WINDOW_HEIGHT, 0)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()
	defer window.Destroy()

	window.SetTitle("Squares")

	// Initialize data
	game = squares.NewGame()
	shapeId := 0
	rotation := 0

	quit := false
	mouseActive := false
	mouseHover := false

	gridCursor := sdl.Rect{
		X: (GRID_WIDTH - 1) / 2 * GRID_CELL_SIZE,
		Y: (GRID_HEIGHT - 1) / 2 * GRID_CELL_SIZE,
		W: GRID_CELL_SIZE,
		H: GRID_CELL_SIZE,
	}
	gridCursorGhost := sdl.Rect{
		X: gridCursor.X,
		Y: gridCursor.Y,
		W: GRID_CELL_SIZE,
		H: GRID_CELL_SIZE,
	}

	nextTime := sdl.GetTicks64()
	for !quit {
		for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch event := e.(type) {
			case *sdl.KeyboardEvent:
				switch event.Keysym.Sym {
				case sdl.K_w, sdl.K_UP:
					gridCursor.Y -= GRID_CELL_SIZE
				case sdl.K_s, sdl.K_DOWN:
					gridCursor.Y += GRID_CELL_SIZE
				case sdl.K_a, sdl.K_LEFT:
					gridCursor.X -= GRID_CELL_SIZE
				case sdl.K_d, sdl.K_RIGHT:
					gridCursor.X += GRID_CELL_SIZE
				}
			case *sdl.MouseButtonEvent:
				if event.Type != sdl.MOUSEBUTTONDOWN || activePlayer != clientId {
					break
				}
				if event.Button == sdl.BUTTON_LEFT {
					// using event.motion does not make sense (as in original C++ code)
					if event.X < BOARD_AREA_WIDTH {
						gridCursor.X = event.X / GRID_CELL_SIZE * GRID_CELL_SIZE
						gridCursor.Y = event.Y / GRID_CELL_SIZE * GRID_CELL_SIZE
						insertPos := squares.Coord{int(gridCursor.X / GRID_CELL_SIZE), int(gridCursor.Y / GRID_CELL_SIZE)}
						if game.TryInsert(shapeId, rotation, insertPos, clientId, firstRound) {
							game.Insert(shapeId, rotation, insertPos, clientId, firstRound)
							if !afterMove(fLocalMultiplayer) {
								fmt.Println("Game over!")
								break
							}
						}
					} else if event.Y < SELECTOR_AREA_HEIGHT {
						if selShape := getSelection(int(event.X), int(event.Y)); selShape >= 0 {
							shapeId = selShape
							rotation = 0
						}
					} else {
						if rotShape := getRotation(int(event.X), int(event.Y), shapeId); rotShape >= 0 {
							rotation = rotShape
						}
					}
				} else if event.Button == sdl.BUTTON_RIGHT {
					rotation = squares.GetNextRotation(shapeId, rotation)
				}
			case *sdl.MouseMotionEvent:
				gridCursorGhost.X = event.X / GRID_CELL_SIZE * GRID_CELL_SIZE
				gridCursorGhost.Y = event.Y / GRID_CELL_SIZE * GRID_CELL_SIZE
				mouseActive = true
			case *sdl.WindowEvent:
				if event.Event == sdl.WINDOWEVENT_ENTER {
					mouseHover = true
				} else if event.Event == sdl.WINDOWEVENT_LEAVE {
					mouseHover = false
				}
			case *sdl.QuitEvent:
				quit = true
			}
		}

		// Draw grid background
		renderer.SetDrawColor(GRID_BACKGROUND.R, GRID_BACKGROUND.G, GRID_BACKGROUND.B, GRID_BACKGROUND.A)
		sdl.Delay(timeLeft(nextTime))
		nextTime = sdl.GetTicks64() + uint64(math.Round(INTERVAL))
		renderer.Clear()

		// Draw grid lines
		renderer.SetDrawColor(GRID_LINE_COLOR.R, GRID_LINE_COLOR.G, GRID_LINE_COLOR.B, GRID_LINE_COLOR.A)
		for x := int32(0); x < BOARD_AREA_WIDTH; x += GRID_CELL_SIZE {
			renderer.DrawLine(x, 0, x, WINDOW_HEIGHT)
		}
		for y := int32(0); y < WINDOW_HEIGHT; y += GRID_CELL_SIZE {
			renderer.DrawLine(0, y, BOARD_AREA_WIDTH-1, y)
		}

		// Draw selector separator line
		for i := int32(0); i < 3; i++ {
			renderer.DrawLine(BOARD_AREA_WIDTH, SELECTOR_AREA_HEIGHT+i, WINDOW_WIDTH, SELECTOR_AREA_HEIGHT+i)
		}

		// Draw grid ghost color
		if mouseActive && mouseHover && shouldRenderGhost(gridCursorGhost, shapeId, rotation) {
			var useColor sdl.Color
			if game.TryInsert(shapeId, rotation, squares.Coord{int(gridCursorGhost.X / GRID_CELL_SIZE), int(gridCursorGhost.Y / GRID_CELL_SIZE)}, clientId, firstRound) {
				useColor = GRID_CURSOR_GHOST_COLORS[clientId]
			} else {
				useColor = GRID_WRONG_COLOR
			}
			renderer.SetDrawColor(useColor.R, useColor.G, useColor.B, useColor.A)
			renderShape(renderer, shapeId, rotation, gridCursorGhost, GRID_CELL_SIZE, GRID_CELL_SIZE)
		}

		renderSelector(renderer, clientId, shapeId)
		renderRotator(renderer, clientId, shapeId, rotation)
		renderBoard(renderer)
		renderer.Present()
	}
}

func main() {
	parseFlags()
	clientMain()
}
