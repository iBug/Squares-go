package main

import (
	"fmt"
	"math"

	squares "github.com/iBug/Squares-go"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	GRID_SIZE     = 36
	W_SELECTOR    = 16
	H_ROTATOR     = 4
	S_SELECELL    = 24
	S_ROTACELL    = 16
	W_ROTATOR     = 6
	FPS_LIMIT     = 60
	SDL_TICKSPEED = 1000
	INTERVAL      = float64(SDL_TICKSPEED) / float64(FPS_LIMIT)

	GRID_CELL_SIZE = GRID_SIZE
	GRID_WIDTH     = squares.BOARD_WIDTH
	GRID_HEIGHT    = squares.BOARD_HEIGHT

	BOARD_WIDTH     = GRID_WIDTH*GRID_CELL_SIZE + 1
	WINDOW_WIDTH    = BOARD_WIDTH + W_SELECTOR*S_SELECELL
	WINDOW_HEIGHT   = GRID_HEIGHT*GRID_CELL_SIZE + 1
	SELECTOR_HEIGHT = squares.BOARD_SIZE * S_SELECELL
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

	id           = 0
	activePlayer = 0
	firstRound   = true
	lostPlayers  = 0
)

func timeLeft(nextTime uint32) uint32 {
	now := sdl.GetTicks()
	if now < nextTime {
		return nextTime - now
	}
	return 0
}

func isSelector(x, y, xOffset int) int {
	for i := 0; i < squares.NCHESS; i++ {
		startX := SELECTOR_POS[i].X*S_SELECELL + xOffset
		endX := startX + squares.GetShape(i, 0).Width*S_SELECELL
		startY := SELECTOR_POS[i].Y * S_SELECELL
		endY := startY + squares.GetShape(i, 0).Height*S_SELECELL
		if x >= startX && x <= endX && y >= startY && y <= endY {
			return i
		}
	}
	return -1
}

func isRotator(x, y, xOffset, yOffset, cmnum int) int {
	for i := 0; i < squares.NROTATIONS; i++ {
		startX := (i%4*W_ROTATOR+2)*S_ROTACELL + xOffset
		endX := startX + squares.GetShape(cmnum, i).Width*S_ROTACELL
		startY := yOffset - ((2-i/4)*W_ROTATOR+1)*S_ROTACELL
		endY := startY + squares.GetShape(cmnum, i).Height*S_ROTACELL
		if x >= startX && x <= endX && y >= startY && y <= endY {
			return i
		}
	}
	return -1
}

func insert(renderer *sdl.Renderer, cmnum int, topleft sdl.Rect, gridColor sdl.Color, rotation int) {
	renderer.SetDrawColor(gridColor.R, gridColor.G, gridColor.B, gridColor.A)
	chessman := squares.GetShape(cmnum, rotation)
	for i := 0; i < chessman.Size(); i++ {
		tmp := topleft
		tmp.X += int32(chessman.Grids[i].X * GRID_SIZE)
		tmp.Y += int32(chessman.Grids[i].Y * GRID_SIZE)
		renderer.FillRect(&tmp)
	}
}

func renderBoard(renderer *sdl.Renderer) {
	for i := 0; i < squares.BOARD_HEIGHT; i++ {
		for j := 0; j < squares.BOARD_WIDTH; j++ {
			if game.At(j, i) >= 0 {
				gridColor := GRID_CURSOR_COLORS[game.At(j, i)]
				renderer.SetDrawColor(gridColor.R, gridColor.G, gridColor.B, gridColor.A)
				rect := sdl.Rect{int32(j * GRID_SIZE), int32(i * GRID_SIZE), int32(GRID_SIZE), int32(GRID_SIZE)}
				renderer.FillRect(&rect)
			}
		}
	}
}

// Extracted common code from renderSelection and renderRotator
func setColorForShape(renderer *sdl.Renderer, shapeId, playerId, activePlayer int, useGhostColor bool) {
	var gridColor sdl.Color
	if game.GetUsed(shapeId, playerId) {
		gridColor = GRID_WRONG_COLOR
	} else if activePlayer == playerId && useGhostColor {
		gridColor = GRID_CURSOR_GHOST_COLORS[id]
	} else {
		gridColor = GRID_CURSOR_COLORS[id]
	}
	renderer.SetDrawColor(gridColor.R, gridColor.G, gridColor.B, gridColor.A)
}

func renderShape(renderer *sdl.Renderer, shapeId, rotation int, base sdl.Rect, width, height int) {
	shape := squares.GetShape(shapeId, rotation)
	for j := 0; j < shape.Size(); j++ {
		tmp := base
		tmp.X += int32(shape.Grids[j].X * width)
		tmp.Y += int32(shape.Grids[j].Y * height)
		renderer.FillRect(&tmp)
	}
}

func renderSelector(renderer *sdl.Renderer, xOffset, id, cmnum int) {
	for i := 0; i < squares.NCHESS; i++ {
		setColorForShape(renderer, i, id, activePlayer, i == cmnum)
		base := sdl.Rect{
			X: int32(SELECTOR_POS[i].X*S_SELECELL + xOffset),
			Y: int32(SELECTOR_POS[i].Y * S_SELECELL),
			W: S_SELECELL,
			H: S_SELECELL,
		}
		renderShape(renderer, i, 0, base, S_SELECELL, S_SELECELL)
	}
}

func renderRotator(renderer *sdl.Renderer, yOffset, xOffset, cmnum, id, rotation int) {
	for i := 0; i < squares.NROTATIONS; i++ {
		setColorForShape(renderer, i, id, activePlayer, i == rotation)
		base := sdl.Rect{
			X: int32((i%4*W_ROTATOR+2)*S_ROTACELL + xOffset),
			Y: int32(yOffset - ((2-i/4)*W_ROTATOR+1)*S_ROTACELL),
			W: S_ROTACELL,
			H: S_ROTACELL,
		}
		renderShape(renderer, cmnum, i, base, S_ROTACELL, S_ROTACELL)
	}
}

// was render_ghost() in original C++ code
func shouldRenderGhost(topleft sdl.Rect, yOffset, xOffset, cmnum, rotation int) bool {
	x := int(topleft.X)+squares.GetShape(cmnum, rotation).Width*GRID_SIZE <= xOffset
	y := int(topleft.Y)+squares.GetShape(cmnum, rotation).Height*GRID_SIZE <= yOffset
	return x && y
}

func isAnyPlayerAlive() bool {
	return lostPlayers != (1<<squares.NPLAYERS)-1
}

func afterMove(singleClient bool) bool {
	if !singleClient {
		firstRound = false
		return true
	}

	// 4 players on one client
	if !firstRound {
		if lp := game.GetLostPlayers(); lp != 0 {
			for i := 0; i < squares.NPLAYERS; i++ {
				if lp&(1<<i) != 0 {
					lostPlayers |= 1 << i
					if !isAnyPlayerAlive() {
						fmt.Printf("Player %d won!\n", i+1)
					} else {
						fmt.Printf("Player %d lost!\n", i+1)
					}
				}
			}
		}
	} else {
		if id == squares.NPLAYERS-1 {
			firstRound = false
		}
	}

	id = (id + 1) % 4
	if lostPlayers != 0 {
		if !isAnyPlayerAlive() {
			activePlayer = -1
			return false
		}
		for _ = id; lostPlayers&(1<<id) != 0; id = (id + 1) % 4 {
		}
	}
	activePlayer = id
	return true
}

func main() {
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
	chessman := 0
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

	nextTime := sdl.GetTicks()
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
				if event.Type != sdl.MOUSEBUTTONDOWN || activePlayer != id {
					break
				}
				if event.Button == sdl.BUTTON_LEFT {
					// using event.motion does not make sense (as in original C++ code)
					if event.X < BOARD_WIDTH {
						gridCursor.X = event.X / GRID_CELL_SIZE * GRID_CELL_SIZE
						gridCursor.Y = event.Y / GRID_CELL_SIZE * GRID_CELL_SIZE
						if game.TryInsert(chessman, rotation, squares.Coord{X: int(gridCursor.X), Y: int(gridCursor.Y)}, id, firstRound) {
							game.Insert(chessman, rotation, squares.Coord{X: int(gridCursor.X), Y: int(gridCursor.Y)}, id, firstRound)
							if !afterMove(false) {
								break
							}
						}
					} else if event.Y < SELECTOR_HEIGHT {
						if cmnum := isSelector(int(event.X), int(event.Y), BOARD_WIDTH); cmnum >= 0 {
							chessman = cmnum
							rotation = 0
						}
					} else {
						if rot := isRotator(int(event.X), int(event.Y), BOARD_WIDTH, WINDOW_HEIGHT, chessman); rot >= 0 {
							rotation = rot
						}
					}
				} else if event.Button == sdl.BUTTON_RIGHT {
					rotation = (rotation + 1) % 8
				}
			case *sdl.MouseMotionEvent:
				gridCursorGhost.X = event.X / GRID_CELL_SIZE * GRID_CELL_SIZE
				gridCursorGhost.Y = event.Y / GRID_CELL_SIZE * GRID_CELL_SIZE
				if !mouseActive {
					mouseActive = true
				}
			case *sdl.WindowEvent:
				if event.Event == sdl.WINDOWEVENT_ENTER && !mouseHover {
					mouseHover = true
				} else if event.Event == sdl.WINDOWEVENT_LEAVE && mouseHover {
					mouseHover = false
				}
			case *sdl.QuitEvent:
				quit = true
			}
		}

		// Draw grid background
		renderer.SetDrawColor(GRID_BACKGROUND.R, GRID_BACKGROUND.G, GRID_BACKGROUND.B, GRID_BACKGROUND.A)
		sdl.Delay(timeLeft(nextTime))
		nextTime += uint32(math.Round(INTERVAL))
		renderer.Clear()

		// Draw grid lines
		renderer.SetDrawColor(GRID_LINE_COLOR.R, GRID_LINE_COLOR.G, GRID_LINE_COLOR.B, GRID_LINE_COLOR.A)
		for x := int32(0); x < GRID_WIDTH*GRID_CELL_SIZE+1; x += GRID_CELL_SIZE {
			renderer.DrawLine(x, 0, x, WINDOW_HEIGHT)
		}
		for y := int32(0); y < GRID_HEIGHT*GRID_CELL_SIZE+1; y += GRID_CELL_SIZE {
			renderer.DrawLine(0, y, BOARD_WIDTH, y)
		}

		// Draw selector separator line
		for i := int32(0); i < 3; i++ {
			renderer.DrawLine(BOARD_WIDTH, SELECTOR_HEIGHT+i, WINDOW_WIDTH, SELECTOR_HEIGHT+1)
		}

		// Draw grid ghost color
		if mouseActive && mouseHover && shouldRenderGhost(gridCursorGhost, WINDOW_HEIGHT, BOARD_WIDTH, chessman, rotation) {
			var useColor sdl.Color
			if game.TryInsert(chessman, rotation, squares.Coord{int(gridCursorGhost.X / GRID_SIZE), int(gridCursorGhost.Y / GRID_SIZE)}, id, firstRound) {
				useColor = GRID_CURSOR_GHOST_COLORS[id]
			} else {
				useColor = GRID_WRONG_COLOR
			}
			insert(renderer, chessman, gridCursorGhost, useColor, rotation)
		}

		renderSelector(renderer, BOARD_WIDTH, id, chessman)
		renderRotator(renderer, WINDOW_HEIGHT, BOARD_WIDTH, chessman, id, rotation)
		renderBoard(renderer)

		renderer.Present()
	}
}
