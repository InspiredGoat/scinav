package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// store global state basically

type Workspace struct {
	Cam             rl.Camera2D
	StudiesExpanded []*Study
	StudiesFiltered []*Study

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

var studySize = rl.Vector2{X: 200, Y: 160}

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

	// Ai generated
	Shorthand string

	// user-defined
	Hearted    bool
	FilteredIn bool
	Expanded   bool
	Tags       []int // stores the tags that are enabled

	// drawing
	TargetOff rl.Vector2 // moves to this target at x speed each frame
	Off       rl.Vector2 // offsets are relative to parent
	AbsPos    rl.Vector2 // offsets are relative to parent
	Selected  bool
	Dragging  bool

	Children []*Study
	Parent   *Study // expanded
}

// all canvas code goes here

func drawStudy(w *Workspace, s *Study, x int32, y int32) {
	// estimate position of children

	// draw self
	of_x := x
	of_y := y
	s.AbsPos.X = float32(x)
	s.AbsPos.Y = float32(y)

	if w.Active == s {
		rl.DrawRectangle(of_x, of_y, int32(studySize.X), int32(studySize.Y), rl.Blue)
	} else if s.Selected {
		rl.DrawRectangle(of_x, of_y, int32(studySize.X), int32(studySize.Y), rl.Pink)
	} else if w.Hot == s {
		rl.DrawRectangle(of_x, of_y, int32(studySize.X), int32(studySize.Y), rl.Orange)
	} else {
		rl.DrawRectangle(of_x, of_y, int32(studySize.X), int32(studySize.Y), rl.White)
	}

	DrawTextBoxed(s.Title, rl.NewRectangle(float32(of_x+5), float32(of_y+5), studySize.X-5, studySize.Y-5), 16, 1, true, false, 0, 0)

	// draw children
	if s.Expanded {
		count_valid := 0
		for _, c := range s.Children {
			if c.FilteredIn {
				count_valid++
			}
		}

		c_h := studySize.Y + 30*float32(count_valid)
		for i, c := range s.Children {
			n_x := float32(of_x) - studySize.X*2 - 10 + float32(c.Off.X)
			n_y := float32(of_y) - (studySize.Y / 2) - float32(c_h) + float32(i)*(studySize.Y+15) + float32(c.Off.Y)

			drawStudy(w, c, int32(n_x), int32(n_y))
			rl.DrawLineBezier(
				rl.Vector2{X: float32(of_x), Y: float32(of_y) + float32(studySize.Y/2)},
				rl.Vector2{X: float32(n_x) + studySize.X, Y: float32(n_y) + float32(studySize.Y/2)}, 2, rl.Gray)
		}
	}
}

func updateStudyUI(w *Workspace, s *Study) {
	m_p := rl.GetMousePosition()
	m_world := rl.GetScreenToWorld2D(m_p, w.Cam)

	if s.Selected && rl.IsKeyDown(rl.KeyLeftControl) && rl.IsKeyPressed(rl.KeyE) {
		go s.ExpandChildren()
	}

	mouse_is_over := rl.CheckCollisionPointRec(m_world, rl.NewRectangle(s.AbsPos.X, s.AbsPos.Y, studySize.X, studySize.Y))

	if mouse_is_over && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.Selected = true
	}
	if !mouse_is_over && rl.IsMouseButtonPressed(rl.MouseLeftButton) && !rl.IsKeyDown(rl.KeyLeftShift) {
		s.Selected = false
	}

	if s.Selected && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.Dragging = true
	}

	if s.Selected && rl.IsKeyPressed(rl.KeyD) && rl.IsKeyDown(rl.KeyLeftControl) {
		s.FilteredIn = false
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.Dragging = false
	}

	if s.Selected {
		if rl.IsKeyPressed(rl.KeyR) {
			s.TargetOff.X = 0
			s.TargetOff.Y = 0
		}
	}

	if s.Selected && rl.IsKeyDown(rl.KeyG) {
		s.TargetOff.X = m_world.X + s.Off.X - s.AbsPos.X - studySize.X/2 // - s.AbsPos.X - studySize.X
		s.TargetOff.Y = m_world.Y + s.Off.Y - s.AbsPos.Y - studySize.Y/2 // - studySize.Y
	}

	if mouse_is_over {
		if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			// active
			w.Active = s
			s.Selected = true
		} else if w.Active != s {
			// hot
			w.Hot = s
		}
	} else if w.Hot == s {
		w.Hot = nil
	}

	// regardless of anything try approaching target
	s.Off.X += (s.TargetOff.X - s.Off.X) * (1 - float32(math.Pow(.5, float64(20.0*rl.GetFrameTime()))))
	s.Off.Y += (s.TargetOff.Y - s.Off.Y) * (1 - float32(math.Pow(.5, float64(20.0*rl.GetFrameTime()))))

	if s.Expanded {
		for _, c := range s.Children {
			updateStudyUI(w, c)
		}
	}
}

func drawCanvas(w *Workspace) {
	for _, root := range w.StudiesExpanded {
		// bfs traversal?
		drawStudy(w, root, int32(root.Off.X), int32(root.Off.Y))
	}

	// check for hover/collision
	for _, root := range w.StudiesExpanded {
		updateStudyUI(w, root)
	}
}

// draws everything on top of canvas
var searchCurPos int32 = 0
var searchCurLength int32 = 0
var searchActive bool = false
var searchHot bool = false
var searchCool float32 = 0
var searchString string = ""
var panelOff float32 = 100

func drawInterface(w *Workspace) {
	// draw rest of UI
	const searchbarWidth int32 = 500
	const searchbarHeight int32 = 32
	// check for hot or actrive
	rect := rl.NewRectangle(float32(rl.GetScreenWidth())/2-float32(searchbarWidth)/2, 10, float32(searchbarWidth), float32(searchbarHeight))

	mbreleased := rl.IsMouseButtonReleased(rl.MouseLeftButton)
	if mbreleased {
		searchActive = false
	}

	if rl.CheckCollisionPointRec(rl.GetMousePosition(), rect) {
		if mbreleased {
			searchActive = true
			searchHot = false
		} else if searchActive == false {
			searchHot = true
		}
	} else {
		searchHot = false
	}

	rl.DrawRectangleRec(rect, rl.White)
	if rl.IsKeyDown(rl.KeyLeftControl) && rl.IsKeyPressed(rl.KeyL) {
		searchActive = true
		searchHot = false
		rl.GetKeyPressed()
	}
	if searchActive {
		if rl.IsKeyDown(rl.KeyLeftControl) && rl.IsKeyPressed(rl.KeyU) {
			searchString = ""
			rl.GetKeyPressed()
		}

		// do interactive stuff
		if rl.IsKeyPressed(rl.KeyBackspace) {
			searchCool = float32(rl.GetTime()) + .1
		}

		if searchCool < float32(rl.GetTime()) {
			if rl.IsKeyDown(rl.KeyBackspace) {
				if len(searchString) > 0 && searchCurPos > 0 {
					bytes := []byte(searchString)
					searchString = string(append(bytes[:searchCurPos-1], bytes[searchCurPos:]...))

					searchCurPos -= 1
					searchCurPos = max(0, searchCurPos)
				}
				searchCool = float32(rl.GetTime()) + .01
			}
		}

		if rl.IsKeyPressed(rl.KeyRight) {
			fmt.Println("LEFT")
			searchCurPos += 1
			searchCurPos = min(int32(len(searchString)), searchCurPos)
		}
		if rl.IsKeyPressed(rl.KeyLeft) {
			fmt.Println("LEFT")
			searchCurPos -= 1
			searchCurPos = max(0, searchCurPos)
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
			// do something with chatgpt for real now

			// give out tags?
			// filter out irrelevant
		}

		key := rl.GetCharPressed()
		for key > 0 {
			searchString += string(rune(key))
			searchCurPos += 1

			key = rl.GetCharPressed()
		}

		rl.DrawRectangleRec(rect, rl.Blue)
	} else if searchHot {
		rl.DrawRectangleRec(rect, rl.Orange)
	} else {
		rl.DrawRectangleRec(rect, rl.White)
	}

	textRect := rect
	textRect.X += 5
	textRect.Y += 5
	textRect.Width -= 5
	textRect.Height -= 5
	DrawTextBoxed(searchString, textRect, 24, 1, false, searchActive, searchCurPos, 0)

	{
		// draw menu for selecting
		const panelWidth = 400
		var panelOffTarget float32 = panelWidth + 100
		if w.Active != nil {
			panelOffTarget = 0
		}

		panelOff += float32(panelOffTarget-(panelOff)) * (1 - float32(math.Pow(.5, float64(15.0*rl.GetFrameTime()))))

		if w.Active != nil {
			x := int32(rl.GetScreenWidth()) - panelWidth - 10 + int32(panelOff)
			rl.DrawRectangle(x, 80, panelWidth, int32(rl.GetScreenHeight())-160, rl.White)

			DrawTextBoxed(w.Active.Title, rl.NewRectangle(float32(x)+5, 80+5, panelWidth-10, 60), 24, 1, true, false, 0, 0)

			// tags

			// DrawTextBoxed(w.Active.Authors, rl.NewRectangle(float32(x)+5, 80+5, panelWidth-10, 40), 24, 1, true, false, 0, 0)
		}
	}
}

func main() {
	fmt.Println("test")

	rl.SetTraceLogLevel(rl.LogNone)
	rl.InitWindow(800, 800, "testapp")
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.SetWindowState(rl.FlagWindowResizable)
	defer rl.CloseWindow()

	var b strings.Builder
	b.Grow(1024)
	//search := b.String()

	var workspace Workspace

	workspace.Cam.Zoom = 1

	go func(workspace *Workspace) {
		s, _ := NewStudyFromDOI("10.2514/6.2005-4282")
		s.Off.X = 0
		s.Off.Y = 0

		workspace.StudiesExpanded = append(workspace.StudiesExpanded, s)
	}(&workspace)
	DefaultFont = rl.LoadFontEx("Roboto-Medium.ttf", 48, nil)

	rl.SetTargetFPS(60)

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
