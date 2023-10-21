package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	// rg "github.com/gen2brain/raylib-go/raygui"
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

var studySize = rl.Vector2{X: 220, Y: 260}

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
	AiShorthand string
	Stage       string

	// user-defined
	Hearted    bool
	FilteredIn bool
	Expanded   bool
	Tags       []int // stores the tags that are enabled

	// drawing
	TargetOff    rl.Vector2 // moves to this target at x speed each frame
	Off          rl.Vector2 // offsets are relative to parent
	AbsPos       rl.Vector2 // offsets are relative to parent
	DragStartPos rl.Vector2 // offsets are relative to parent
	Selected     bool
	Dragging     bool

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

	if s.FilteredIn {
		rl.DrawRectangle(of_x-5, of_y-5, int32(studySize.X)+5, int32(studySize.Y)+5, rl.Green)
	}

	if w.Active == s {
		rl.DrawRectangle(of_x, of_y, int32(studySize.X), int32(studySize.Y), rl.NewColor(200, 200, 200, 255))
	} else if s.Selected {
		rl.DrawRectangle(of_x, of_y, int32(studySize.X), int32(studySize.Y), rl.NewColor(220, 220, 220, 255))
	} else if w.Hot == s {
		rl.DrawRectangle(of_x, of_y, int32(studySize.X), int32(studySize.Y), rl.NewColor(230, 230, 230, 255))
	} else {
		rl.DrawRectangle(of_x, of_y, int32(studySize.X), int32(studySize.Y), rl.White)
	}

	title := s.Title
	if len(s.AiShorthand) > 0 {
		title = s.AiShorthand
	}

	DrawTextBoxed(title, rl.NewRectangle(float32(of_x+5), float32(of_y+5), studySize.X-5, studySize.Y-5), 24, 1, true, false, 0, 0)

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
			// if c.FilteredIn {
			n_x := float32(of_x) - studySize.X*2 - 10 + float32(c.Off.X)
			n_y := float32(of_y) - (studySize.Y / 2) - float32(c_h) + float32(i)*(studySize.Y+15) + float32(c.Off.Y)

			color := rl.Gray
			if c.FilteredIn {
				color = rl.Green
			}

			drawStudy(w, c, int32(n_x), int32(n_y))
			rl.DrawLineBezier(
				rl.Vector2{X: float32(of_x), Y: float32(of_y) + float32(studySize.Y/2)},
				rl.Vector2{X: float32(n_x) + studySize.X, Y: float32(n_y) + float32(studySize.Y/2)}, 2, color)
			// }
		}
	}
}

var dragStartPos rl.Vector2

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

	if s == w.Active && !mouse_is_over && rl.IsMouseButtonReleased(rl.MouseLeftButton) && rl.GetMousePosition().X < float32(rl.GetScreenWidth()-600) {
		w.Active = nil
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

	if s.Selected && rl.IsKeyPressed(rl.KeyG) {
		dragStartPos = rl.GetScreenToWorld2D(rl.GetMousePosition(), w.Cam)
		s.DragStartPos = s.AbsPos
	}
	if s.Selected && rl.IsKeyDown(rl.KeyG) {
		s.TargetOff.X = m_world.X + s.Off.X - s.AbsPos.X - studySize.X/2 // - s.AbsPos.X - studySize.X
		s.TargetOff.Y = m_world.Y + s.Off.Y - s.AbsPos.Y - studySize.Y/2 // - studySize.Y
	}

	if mouse_is_over {
		if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			// active
			w.Active = s
		} else if w.Active != s {
			// hot
			w.Hot = s
		}
	} else if w.Hot == s {
		w.Hot = nil
	}

	// regardless of anything try approaching target
	s.Off.X += (s.TargetOff.X - s.Off.X) * (1 - float32(math.Pow(.5, float64(23.0*rl.GetFrameTime()))))
	s.Off.Y += (s.TargetOff.Y - s.Off.Y) * (1 - float32(math.Pow(.5, float64(23.0*rl.GetFrameTime()))))

	if s.Expanded {
		for _, c := range s.Children {
			updateStudyUI(w, c)
		}
	}
}

func betterTitles(w *Workspace, s *Study) {
	go func(s *Study) {
		response := AIPrompt("Come up with a more concise and practical title for this paper: \"" + s.Title + "\", it was published in " + s.PublicationDate.Format("2006-01-02") + " by the journal " + s.Journal)

		fmt.Println(response)
		s.AiShorthand = response
	}(s)
	if s.Expanded {
		for _, c := range s.Children {
			betterTitles(w, c)
		}
	}
}

func filterStudies(w *Workspace, s *Study, filter string, resetFilter bool) {

	if resetFilter {
		s.FilteredIn = true
	} else {
		go func(s *Study) {
			response := AIPrompt("Here's a filter, does the following data conform to it? \n#FILTER: " + filter + ".\n Your answer should only be the words yes or no. \n#DATA title: \"" + s.Title + "\"\n publish date: " + s.PublicationDate.Format("2006-01-02") + " \njournal author: " + s.Journal + "\nreference count:" + strconv.Itoa(len(s.References)) + "\n times it has been referenced: " + strconv.Itoa(s.IsReferencedCount)))

			fmt.Println(response)
			if strings.Contains(response, "Yes") {
				s.FilteredIn = true
			} else {
				s.FilteredIn = false
			}
		}(s)
	}

	if s.Expanded {
		for _, c := range s.Children {
			filterStudies(w, c, filter, resetFilter)
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
	const searchbarHeight int32 = 30
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
		if rl.IsKeyDown(rl.KeyLeftControl) && rl.IsKeyPressed(rl.KeyV) {
			searchString = searchString[:searchCurPos] + rl.GetClipboardText() + searchString[searchCurPos:]
		}

		// do interactive stuff
		if rl.IsKeyPressed(rl.KeyBackspace) {
			if len(searchString) > 0 && searchCurPos > 0 {
				bytes := []byte(searchString)
				searchString = string(append(bytes[:searchCurPos-1], bytes[searchCurPos:]...))

				searchCurPos -= 1
				searchCurPos = max(0, searchCurPos)
			}
			searchCool = float32(rl.GetTime()) + .1
		}
		if rl.IsKeyPressed(rl.KeyRight) {
			searchCurPos += 1
			searchCurPos = min(int32(len(searchString)), searchCurPos)
			searchCool = float32(rl.GetTime()) + .1
		}
		if rl.IsKeyPressed(rl.KeyLeft) {
			searchCurPos -= 1
			searchCurPos = max(0, searchCurPos)
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
			if rl.IsKeyDown(rl.KeyRight) {
				searchCurPos += 1
				searchCurPos = min(int32(len(searchString)), searchCurPos)
				searchCool = float32(rl.GetTime()) + .01
			}
			if rl.IsKeyDown(rl.KeyLeft) {
				searchCurPos -= 1
				searchCurPos = max(0, searchCurPos)
				searchCool = float32(rl.GetTime()) + .01
			}
		}

		if rl.IsKeyPressed(rl.KeyEnter) {
			// do something with chatgpt for real now

			// do ADD CODE
			if len(searchString) > 4 {
				if strings.Compare("ADD", searchString[0:3]) == 0 {
					fmt.Println(searchString[4:])
					NewStudyFromDOI(searchString[4:])
				} else {
					for _, s := range w.StudiesExpanded {
						if len(searchString) == 0 {
							filterStudies(w, s, searchString, true)
						} else {
							filterStudies(w, s, searchString, false)
						}
					}
				}
			}
			// do FILTERING

			// do ADD SEARCH

			// do TAG

			// give out tags?
			// filter out irrelevant
		}

		key := rl.GetCharPressed()
		for key > 0 {
			searchString += string(rune(key))
			searchCurPos += 1

			key = rl.GetCharPressed()
		}

		rl.DrawRectangleRec(rect, rl.White)
	} else if searchHot {
		rl.DrawRectangleRec(rect, rl.White)
	} else {
		rl.DrawRectangleRec(rect, rl.White)
	}

	if rl.IsKeyPressed(rl.KeyG) && rl.IsKeyDown(rl.KeyLeftControl) {
		for _, s := range w.StudiesExpanded {
			betterTitles(w, s)
		}
	}

	textRect := rect
	textRect.X += 5
	textRect.Y += 7
	textRect.Width -= 5
	textRect.Height -= 5
	DrawTextBoxed(searchString, textRect, 16, 1, false, searchActive, searchCurPos, 0)

	{
		// draw menu for selecting
		const panelWidth = 600
		var panelOffTarget float32 = panelWidth + 10

		if w.Active != nil {
			panelOffTarget = 0
		} else {
			panelOffTarget = panelWidth + 10
		}

		panelOff += float32(panelOffTarget-(panelOff)) * (1 - float32(math.Pow(.5, float64(25.0*rl.GetFrameTime()))))

		x := int32(rl.GetScreenWidth()) - panelWidth - 10 + int32(panelOff)
		y := 80
		rl.DrawRectangle(x, int32(y), panelWidth, int32(rl.GetScreenHeight())-160, rl.White)

		x += 10

		if w.Active != nil {

			title := w.Active.Title

			DrawTextBoxed(title, rl.NewRectangle(float32(x)+5, float32(y+5), panelWidth-10, 80), 24, 1, true, false, 0, 0)

			y += 65

			// tags

			authors := "by "
			for i, a := range w.Active.Authors {
				authors += a
				if i != len(w.Active.Authors)-1 {
					authors += ", "
				}
			}
			DrawTextBoxed(authors, rl.NewRectangle(float32(x)+5, float32(y+5), panelWidth-10, 80), 16, 1, true, false, 0, 0)

			{
				if len(w.Active.Journal) > 3 {
					y += 25
					text := "Published by: " + w.Active.Journal
					DrawTextBoxed(text, rl.NewRectangle(float32(x)+5, float32(y+5), panelWidth-10, 80), 16, 1, true, false, 0, 0)
				}
				{
					if len(w.Active.Journal) > 3 {
						y += 25
						text := "Published on: " + w.Active.PublicationDate.Format("2006-01-02")
						DrawTextBoxed(text, rl.NewRectangle(float32(x)+5, float32(y+5), panelWidth-10, 80), 16, 1, true, false, 0, 0)
					}
				}
				y += 25
				{
					text := "Cites: " + strconv.Itoa(len(w.Active.References)) + " other works"
					DrawTextBoxed(text, rl.NewRectangle(float32(x)+5, float32(y+5), panelWidth-10, 80), 16, 1, true, false, 0, 0)
				}
				y += 25
				{
					text := "Is cited by: " + strconv.Itoa(int(w.Active.IsReferencedCount)) + " others"
					DrawTextBoxed(text, rl.NewRectangle(float32(x)+5, float32(y+5), panelWidth-10, 80), 16, 1, true, false, 0, 0)
				}
				y += 32
				{
					if Button("Open in google scholar", x, int32(y), 16) {
						DOISanitized := strings.ReplaceAll(w.Active.DOI, "/", "%2F")
						rl.OpenURL("https://scholar.google.com/scholar?hl=en&as_sdt=0%2C5&q=" + DOISanitized + "&btnG=")
					}
				}
				y += 32
				{
					if Button("Open in scihub", x, int32(y), 16) {
						rl.OpenURL("https://sci-hub.hkvisa.net/" + w.Active.DOI)
					}
				}

				y += 32
				{
					if Button("Expand citations", x, int32(y), 16) {
						go w.Active.ExpandChildren()
					}
				}
			}
		}
	}
}

func main() {
	fmt.Println("test")

	rl.SetTraceLogLevel(rl.LogNone)
	rl.InitWindow(1640, 1024, "testapp")
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
	DefaultFont = rl.LoadFontEx("Roboto-Medium.ttf", 32, nil)
	rl.SetTextureFilter(DefaultFont.Texture, rl.FilterBilinear)

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {

		// check for collision, if no collision move background basically

		if rl.IsFileDropped() {

		}

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
