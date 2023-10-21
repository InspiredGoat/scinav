package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gen2brain/raylib-go/raylib"
)

// store global state basically

type Workspace struct {
	Cam             rl.Camera2D
	StudiesExpanded []Study // root studies
	StudiesFiltered []Study // root studies with filtered removed

	RootStudies []*Study

	Hot    *Study
	Active *Study

	Tags []string // names of tags
}

type Reference struct {
	Key            string
	DOI            string
	Unstructured   string
	ArbitraryOrder string
}

var studyPreviewSize = rl.Vector2{X: 200, Y: 320}

type Study struct {
	// extracted
	Title             string // ^
	Subtitle          string // ^
	Journal           string // ^
	Authors           []string
	PublicationDate   time.Time // ^
	References        []Reference
	IsReferencedCount int64  // ^
	DOI               string // ^

	// user-defined
	Selected bool
	Expanded bool
	Tags     []int // stores the tags that are enabled

	// drawing
	TargetOff rl.Vector2 // moves to this target at x speed each frame
	Off       rl.Vector2 // offsets are relative to parent

	Children []*Study
	Parent   *Study // expanded
}

// all canvas code goes here

func drawCanvas(w *Workspace) {
	// draw studies as graph

	for _, root := range w.StudiesExpanded {
		// bfs traversal?
		root
		// rl.DrawRectangleRounded(rl.NewRectangle(400, 400, 200, 320), .2, 32, rl.White)
		// rl.DrawRectangleRounded(rl.NewRectangle(400, 400, 200, 20), .2, 32, rl.Gray)
	}
}

// draws everything on top of canvas
func drawInterface(w *Workspace) {
	// draw rest of UI
	// const searchbar_width int32 = 500
	// rl.DrawRectangle(int32(rl.GetScreenWidth())/2-(searchbar_width)/2, 10, searchbar_width, 30, rl.Gray)

	// // draw menu for selecting
	// rl.DrawRectangle(int32(rl.GetScreenWidth())/2-(400)/2, 300, 400, 300, rl.Gray)
}

func main() {
	// s2, _ := NewStudyFromDOI("10.2514/6.2005-4282")
	// // for _, r := range s2.References {
	// // 	// fmt.Println(r.Key)
	// // 	// fmt.Println(r.ArbitraryOrder)
	// // 	// fmt.Println(r.Unstructured)
	// // 	// fmt.Println(r.DOI)
	// // }
	// fmt.Println(s2.Title)

	// s2.ExpandChildren()

	// for _, r := range s2.Children {
	// 	fmt.Println(r.Title)
	// 	fmt.Println(r.Authors)
	// 	fmt.Println("---")
	// }

	// if true {
	// 	return
	// }

	fmt.Println("test")

	rl.SetTraceLogLevel(rl.LogNone)
	rl.InitWindow(800, 800, "testapp")
	rl.SetWindowState(rl.FlagWindowResizable)
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	var b strings.Builder
	b.Grow(1024)
	//search := b.String()

	var workspace Workspace
	var s Study
	s.Off.X = 400
	s.Off.Y = 400
	workspace.Cam.Zoom = 1

	workspace.StudiesExpanded = append(workspace.StudiesExpanded, s)

	for !rl.WindowShouldClose() {

		// check for collision, if no collision move background basically

		if rl.IsMouseButtonDown(rl.MouseRightButton) {
			workspace.Cam.Target.X -= rl.GetMouseDelta().X / workspace.Cam.Zoom
			workspace.Cam.Target.Y -= rl.GetMouseDelta().Y / workspace.Cam.Zoom
		}

		scroll := rl.GetMouseWheelMoveV().Y

		// get before zoom
		mw_xp := rl.GetMousePosition().X/workspace.Cam.Zoom + workspace.Cam.Target.X
		mw_yp := rl.GetMousePosition().Y/workspace.Cam.Zoom + workspace.Cam.Target.Y
		if scroll > 0.0 {
			workspace.Cam.Zoom *= 1.3
		} else if scroll < 0.0 {
			workspace.Cam.Zoom *= 0.7
		}
		workspace.Cam.Zoom = max(workspace.Cam.Zoom, .1)
		workspace.Cam.Zoom = min(workspace.Cam.Zoom, 2.5)

		if scroll != 0.0 {
			mw_x := rl.GetMousePosition().X/workspace.Cam.Zoom + workspace.Cam.Target.X
			mw_y := rl.GetMousePosition().Y/workspace.Cam.Zoom + workspace.Cam.Target.Y

			workspace.Cam.Target.X += mw_xp - mw_x
			workspace.Cam.Target.Y += mw_yp - mw_y
		}

		rl.BeginDrawing()

		// Canvas mode
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(workspace.Cam)

		// make new list with filtered only
		// draw it
		drawCanvas(&workspace)

		rl.EndMode2D()

		// Reference mode
		drawInterface(&workspace)

		rl.EndDrawing()
	}
}
