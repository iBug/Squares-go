package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net"

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
	SELECTOR_POS = []squares.Coord{{1, 1}, {1, 3}, {1, 5}, {1, 7}, {1, 10}, {1, 12}, {1, 15}, {1, 18}, {6, 1}, {10, 1}, {6, 4}, {6, 7}, {6, 10}, {6, 13}, {6, 17}, {12, 3}, {12, 10}, {12, 14}, {9, 15}, {12, 18}, {12, 7}}
)

var (
	game         = squares.NewGame()
	clientId     = 0
	clientPlayer = 0 // which player this client represents

	// Global event channel, as a complement for sdl.PushEvent
	chEvent = make(chan any, 8)

	fServerAddr       = ""
	fIsServer         = false
	fLocalMultiplayer = false
	fUseDarkTheme     = false
)

type ConnectionLost struct {
	err error
}

func parseFlags() {
	flag.StringVar(&fServerAddr, "a", "", "server address")
	flag.BoolVar(&fUseDarkTheme, "d", false, "use dark theme")
	flag.BoolVar(&fIsServer, "s", false, "run as server")
	flag.IntVar(&clientId, "i", 0, "client id (for reconnection)")
	flag.Parse()

	fLocalMultiplayer = fServerAddr == ""

	if fUseDarkTheme {
		setDarkTheme()
	}
}

func getSelection(x, y int) int {
	for i := 0; i < squares.NSHAPES; i++ {
		shape := squares.GetShape(i, 0)
		startX := SELECTOR_POS[i].X*SELECTOR_CELL_SIZE + BOARD_AREA_WIDTH
		endX := startX + shape.Width*SELECTOR_CELL_SIZE
		startY := SELECTOR_POS[i].Y * SELECTOR_CELL_SIZE
		endY := startY + shape.Height*SELECTOR_CELL_SIZE
		if x >= startX && x < endX && y >= startY && y < endY {
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
		shape := squares.GetShape(shapeId, i)
		startX := (i%4*ROTATOR_WIDTH+2)*ROTATOR_CELL_SIZE + BOARD_AREA_WIDTH
		endX := startX + shape.Width*ROTATOR_CELL_SIZE
		startY := WINDOW_HEIGHT - ((2-i/4)*ROTATOR_WIDTH+1)*ROTATOR_CELL_SIZE
		endY := startY + shape.Height*ROTATOR_CELL_SIZE
		if x >= startX && x < endX && y >= startY && y < endY {
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
func setColorForShape(renderer *sdl.Renderer, shapeId int, useGhostColor bool) {
	var gridColor sdl.Color
	if game.GetUsed(clientPlayer, shapeId) {
		gridColor = GRID_WRONG_COLOR
	} else if game.ActivePlayer == clientPlayer && useGhostColor {
		gridColor = GRID_CURSOR_GHOST_COLORS[clientPlayer]
	} else {
		gridColor = GRID_CURSOR_COLORS[clientPlayer]
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

func renderSelector(renderer *sdl.Renderer, clientPlayer, shapeId int) {
	for i := 0; i < squares.NSHAPES; i++ {
		setColorForShape(renderer, i, i == shapeId)
		base := sdl.Rect{
			X: int32(SELECTOR_POS[i].X*SELECTOR_CELL_SIZE + BOARD_AREA_WIDTH),
			Y: int32(SELECTOR_POS[i].Y * SELECTOR_CELL_SIZE),
			W: SELECTOR_CELL_SIZE,
			H: SELECTOR_CELL_SIZE,
		}
		renderShape(renderer, i, 0, base, SELECTOR_CELL_SIZE, SELECTOR_CELL_SIZE)
	}
}

func renderRotator(renderer *sdl.Renderer, clientPlayer, shapeId, rotation int) {
	rotations := squares.AvailableRotations(shapeId)
	for i := 0; i < squares.NROTATIONS; i++ {
		if rotations&(1<<i) == 0 {
			continue
		}
		setColorForShape(renderer, i, i == rotation)
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
	shape := squares.GetShape(shapeId, rotation)
	x := int(topleft.X)+shape.Width*GRID_CELL_SIZE <= WINDOW_HEIGHT
	y := int(topleft.Y)+shape.Height*GRID_CELL_SIZE <= BOARD_AREA_WIDTH
	return x && y
}

func pushSdlEvent(code uint32, windowId uint32, e any) (bool, error) {
	chEvent <- e
	return sdl.PushEvent(&sdl.UserEvent{
		Type:      code,
		Timestamp: sdl.GetTicks(), //lint:ignore SA1019 compat
		WindowID:  windowId,
	})
}

func clientNetThread(conn net.Conn, windowId uint32) {
	defer conn.Close()
	for {
		msg, err := RecvMsg(conn)
		if err != nil {
			pushSdlEvent(sdl.USEREVENT, windowId, ConnectionLost{err})
			break
		}
		pushSdlEvent(sdl.USEREVENT, windowId, msg)
	}
}

func setupClientNetThread(window *sdl.Window) (net.Conn, error) {
	conn, err := net.Dial("tcp", fServerAddr)
	if err != nil {
		return nil, err
	}
	windowID, _ := window.GetID()
	go clientNetThread(conn, windowID)
	return conn, SendMsg(conn, ConnectReq{Id: clientId})
}

func clientMain() {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	window, renderer, err := sdl.CreateWindowAndRenderer(WINDOW_WIDTH, WINDOW_HEIGHT, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Destroy()
	defer window.Destroy()
	window.SetTitle(fmt.Sprintf("Squares (Player %d)", clientPlayer+1))

	var conn net.Conn
	if !fLocalMultiplayer {
		conn, err = setupClientNetThread(window)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Initialize data
	game.Reset()
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
		for e := sdl.WaitEvent(); e != nil; e = sdl.PollEvent() {
			switch event := e.(type) {
			case *sdl.KeyboardEvent:
				if event.State != sdl.PRESSED {
					continue
				}
				switch event.Keysym.Sym {
				case sdl.K_w, sdl.K_UP:
					gridCursor.Y -= GRID_CELL_SIZE
				case sdl.K_s, sdl.K_DOWN:
					gridCursor.Y += GRID_CELL_SIZE
				case sdl.K_a, sdl.K_LEFT:
					gridCursor.X -= GRID_CELL_SIZE
				case sdl.K_d, sdl.K_RIGHT:
					gridCursor.X += GRID_CELL_SIZE
				case sdl.K_q:
					rotation = squares.GetPrevRotation(shapeId, rotation)
				case sdl.K_e, sdl.K_SPACE:
					rotation = squares.GetNextRotation(shapeId, rotation)
				case sdl.K_r:
					if !fLocalMultiplayer {
						SendMsg(conn, ConnectReq{Id: clientId})
					}
				}
			case *sdl.MouseWheelEvent:
				if event.Y > 0 {
					// Scroll up
					rotation = squares.GetPrevRotation(shapeId, rotation)
				} else if event.Y < 0 {
					// Scroll down
					rotation = squares.GetNextRotation(shapeId, rotation)
				}
			case *sdl.MouseButtonEvent:
				if event.Type != sdl.MOUSEBUTTONDOWN || game.ActivePlayer != clientPlayer {
					break
				}
				if event.Button == sdl.BUTTON_LEFT {
					// using event.motion does not make sense (as in original C++ code)
					if event.X < BOARD_AREA_WIDTH {
						gridCursor.X = event.X / GRID_CELL_SIZE * GRID_CELL_SIZE
						gridCursor.Y = event.Y / GRID_CELL_SIZE * GRID_CELL_SIZE
						insertPos := squares.Coord{int(gridCursor.X / GRID_CELL_SIZE), int(gridCursor.Y / GRID_CELL_SIZE)}
						if game.TryInsert(shapeId, rotation, insertPos, clientPlayer, game.FirstRound) {
							if fLocalMultiplayer {
								game.Insert(shapeId, rotation, insertPos, clientPlayer)
								if !game.AfterMove() {
									log.Println("Game over!")
									break
								}
								clientPlayer = game.ActivePlayer
							} else {
								SendMsg(conn, MoveReq{
									Id:       clientId,
									ShapeId:  shapeId,
									Pos:      [2]int{insertPos.X, insertPos.Y},
									Rotation: rotation,
								})
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

			case *sdl.UserEvent:
				switch event := (<-chEvent).(type) {
				case ConnectRes:
					if event.Id == 0 {
						log.Fatal("Connection failed")
						break
					}
					game = &event.Game
					if clientId != event.Id {
						log.Printf("Updated client ID: %d\n", event.Id)
						clientId = event.Id
					}
					if clientPlayer != event.PlayerId {
						log.Printf("Updated client player: %d\n", event.PlayerId)
						clientPlayer = event.PlayerId
						window.SetTitle(fmt.Sprintf("Squares (Player %d)", clientPlayer+1))
					}
				case MoveRes:
					break
				case OtherMoveRes:
					pos := squares.Coord{event.Pos[0], event.Pos[1]}
					game.Insert(event.ShapeId, event.Rotation, pos, event.PlayerId)
					game.ActivePlayer = event.ActivePlayer
					if event.PlayerId == squares.NPLAYERS-1 {
						game.FirstRound = false
					}
				case ServerRes:
					log.Printf("Server message: [%d] %s\n", event.Code, ServerResString(event.Code))

				case ConnectionLost:
					errorS := fmt.Sprintf("Connection lost: %s", event.err)
					log.Println(errorS)
					sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_ERROR, "Error", errorS, nil)

					// Retry connection
					conn, err = setupClientNetThread(window)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}

		// Draw grid background
		renderer.SetDrawColor(GRID_BACKGROUND.R, GRID_BACKGROUND.G, GRID_BACKGROUND.B, GRID_BACKGROUND.A)
		sdlNow := sdl.GetTicks64()
		if sdlNow < nextTime {
			sdl.Delay(uint32(nextTime - sdlNow))
			nextTime = sdlNow + uint64(math.Round(INTERVAL))
		}
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
			if game.TryInsert(shapeId, rotation, squares.Coord{int(gridCursorGhost.X / GRID_CELL_SIZE), int(gridCursorGhost.Y / GRID_CELL_SIZE)}, clientPlayer, game.FirstRound) {
				useColor = GRID_CURSOR_GHOST_COLORS[clientPlayer]
			} else {
				useColor = GRID_WRONG_COLOR
			}
			renderer.SetDrawColor(useColor.R, useColor.G, useColor.B, useColor.A)
			renderShape(renderer, shapeId, rotation, gridCursorGhost, GRID_CELL_SIZE, GRID_CELL_SIZE)
		}

		renderSelector(renderer, clientPlayer, shapeId)
		renderRotator(renderer, clientPlayer, shapeId, rotation)
		renderBoard(renderer)
		renderer.Present()
	}
}

func main() {
	parseFlags()
	if fIsServer {
		serverMain()
	} else {
		clientMain()
	}
}
