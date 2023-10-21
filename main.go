package main

import (
	"bytes"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/gen2brain/raylib-go/raylib"
	"net/http"
)

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

	for !rl.WindowShouldClose() {

		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			cam.Target.X -= rl.GetMouseDelta().X / cam.Zoom
			cam.Target.Y -= rl.GetMouseDelta().Y / cam.Zoom
		}

		scroll := rl.GetMouseWheelMoveV().Y

		// get before zoom
		mw_xp := rl.GetMousePosition().X/cam.Zoom + cam.Target.X
		mw_yp := rl.GetMousePosition().Y/cam.Zoom + cam.Target.Y
		if scroll > 0.0 {
			cam.Zoom *= 1.1
		} else if scroll < 0.0 {
			cam.Zoom *= 0.9
		}

		if scroll != 0.0 {
			mw_x := rl.GetMousePosition().X/cam.Zoom + cam.Target.X
			mw_y := rl.GetMousePosition().Y/cam.Zoom + cam.Target.Y

			cam.Target.X += mw_xp - mw_x
			cam.Target.Y += mw_yp - mw_y
		}

		rl.BeginDrawing()

		// draw canvas
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(cam)

		rl.DrawRectangle(400, 400, 200, 200, rl.Green)
		rl.EndMode2D()

		rl.EndDrawing()
	}
}
