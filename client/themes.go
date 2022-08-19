package main

import "github.com/veandco/go-sdl2/sdl"

var (
	GRID_CURSOR_COLORS, GRID_CURSOR_GHOST_COLORS       []sdl.Color
	GRID_BACKGROUND, GRID_LINE_COLOR, GRID_WRONG_COLOR sdl.Color
)

func setDarkTheme() {
	GRID_CURSOR_COLORS = []sdl.Color{{0xB2, 0x1B, 0x1A, 255}, {0x2C, 0xAA, 0x18, 255}, {0x1E, 0x28, 0xA4, 255}, {0xF9, 0xC8, 0x02, 255}}
	GRID_CURSOR_GHOST_COLORS = []sdl.Color{{0xF7, 0x4F, 0x52, 127}, {0xB3, 0xFB, 0x3B, 127}, {0x89, 0xEC, 0xFC, 127}, {0xFC, 0xE5, 0x4F, 127}}

	GRID_BACKGROUND = sdl.Color{22, 22, 22, 255}
	GRID_LINE_COLOR = sdl.Color{44, 44, 44, 255}
	GRID_WRONG_COLOR = sdl.Color{127, 127, 127, 255}
}

func setLightTheme() {
	GRID_CURSOR_COLORS = []sdl.Color{{0xB2, 0x1B, 0x1A, 255}, {0x2C, 0xAA, 0x18, 255}, {0x1E, 0x28, 0xA4, 255}, {0xF9, 0xC8, 0x02, 255}}
	GRID_CURSOR_GHOST_COLORS = []sdl.Color{{0xF7, 0x4F, 0x52, 127}, {0xB3, 0xFB, 0x3B, 127}, {0x89, 0xEC, 0xFC, 127}, {0xFC, 0xE5, 0x4F, 127}}

	GRID_BACKGROUND = sdl.Color{233, 233, 233, 255}
	GRID_LINE_COLOR = sdl.Color{200, 200, 200, 255}
	GRID_WRONG_COLOR = sdl.Color{127, 127, 127, 255}
}

func init() {
	setLightTheme()
}
