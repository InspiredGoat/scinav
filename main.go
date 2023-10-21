package main

import (
	"bytes"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/gen2brain/raylib-go/raylib"
	"net/http"
	rg "scinav/raygui"
	"strings"
	"time"
)

// store global state basically

type Workspace struct {
	Camera          rl.Camera2D
	StudiesExpanded []Study
	StudiesFiltered []Study

	Tags []string // names of tags
}

type Study struct {
	// extracted
	Name            string
	Authors         []string
	PublicationDate time.Time

	// user-defined
	Enabled bool
	Tags    []int // stores the tags that are enabled

	Children *[]Study
	Parents  *[]Study // expanded
}

func crossref(req string) *http.Response {
	res, _ := http.Get(req + "&mailto=tomd@airmail.cc")
	return res
}

func main() {
	res := crossref("https://api.crossref.org/works?query=petrol&rows=100")

	buf := new(bytes.Buffer)

	buf.ReadFrom(res.Body)
	defer res.Body.Close()

	jsonparser.ArrayEach(buf.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		fmt.Println(jsonparser.GetString(value, "abstract"))

		jsonparser.ArrayEach(buf.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			fmt.Println(jsonparser.GetString(value, "abstract"))
		}, "author")
	}, "message", "items")

	var cam rl.Camera2D
	cam.Zoom = 1

	rl.SetTraceLogLevel(rl.LogNone)
	rl.InitWindow(800, 800, "testapp")
	rl.SetWindowState(rl.FlagWindowResizable)
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// var prev_size_x float32 = float32(rl.GetScreenWidth())
	// var prev_size_y float32 = float32(rl.GetScreenHeight())

	var b strings.Builder
	b.Grow(1024)
	search := b.String()

	for !rl.WindowShouldClose() {

		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			cam.Target.X -= rl.GetMouseDelta().X / cam.Zoom
			cam.Target.Y -= rl.GetMouseDelta().Y / cam.Zoom
		}

		if rl.IsWindowResized() {
			// cam.Target.X -= rl.GetMouseDelta().X / cam.Zoom
			// cam.Target.Y -= rl.GetMouseDelta().Y / cam.Zoom
		}

		scroll := rl.GetMouseWheelMoveV().Y

		// get before zoom
		mw_xp := rl.GetMousePosition().X/cam.Zoom + cam.Target.X
		mw_yp := rl.GetMousePosition().Y/cam.Zoom + cam.Target.Y
		if scroll > 0.0 {
			cam.Zoom *= 1.3
		} else if scroll < 0.0 {
			cam.Zoom *= 0.7
		}
		cam.Zoom = max(cam.Zoom, .1)
		cam.Zoom = min(cam.Zoom, 2.5)

		if scroll != 0.0 {
			mw_x := rl.GetMousePosition().X/cam.Zoom + cam.Target.X
			mw_y := rl.GetMousePosition().Y/cam.Zoom + cam.Target.Y

			cam.Target.X += mw_xp - mw_x
			cam.Target.Y += mw_yp - mw_y
		}

		rl.BeginDrawing()

		// Canvas mode
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(cam)

		// make new list with filtered only
		// draw it
		rl.DrawRectangle(400, 400, 200, 200, rl.Green)

		rl.EndMode2D()

		// Reference mode

		// draw rest of UI
		// const searchbar_width int32 = 500
		// rl.DrawRectangle(int32(rl.GetScreenWidth())/2-(searchbar_width)/2, 10, searchbar_width, 30, rl.Gray)

		// // draw menu for selecting
		// rl.DrawRectangle(int32(rl.GetScreenWidth())/2-(400)/2, 300, 400, 300, rl.Gray)

		rl.EndDrawing()
	}
}
