package main

import rl "github.com/gen2brain/raylib-go/raylib"
import "unsafe"

var searchbar_text string

func draw_searchbar() {

}

var DefaultFont rl.Font
var SelectedTint rl.Color = rl.Black
var SelectedBackTint rl.Color = rl.Blue

func DrawText(text string, x float32, y float32, size float32) {
	rl.DrawTextEx(DefaultFont, text, rl.Vector2{X: x, Y: y}, size, 1.0, rl.Black)
	// rg.Text
}

func getCodepoint(text []byte, off int, ret_size *int) int32 {
	var res int32

	if text[off]&0b11110000 > 0 {
		res = int32(text[off]) << 24
		res = int32(text[off+1]) << 16
		res = int32(text[off+2]) << 8
		res = int32(text[off+3]) << 0
		*ret_size = 4
	} else if text[off]&0b11100000 > 0 {
		res = int32(text[off]) << 16
		res = int32(text[off+1]) << 8
		res = int32(text[off+2]) << 0
		*ret_size = 3
	} else if text[off]&0b11000000 > 0 {
		res = int32(text[off]) << 8
		res = int32(text[off+1]) << 0
		*ret_size = 2
	} else {
		*ret_size = 1
		res = int32(text[off])
	}

	return res
}

func DrawTextBoxed(text string, rec rl.Rectangle, size float32, spacing float32, wordWrap bool, drawCursor bool, selectStart int32, selectLength int32) {
	length := len(text) // Total length in bytes of the text, scanned by codepoints in loop

	var textOffsetY float32 // Offset between lines (on line break '\n')
	var textOffsetX float32 // Offset X to next character to draw

	scaleFactor := size / float32(DefaultFont.BaseSize) // Character rectangle scaling factor

	// Word/character wrapping mechanism variables
	state := false
	if wordWrap == true {
		// measuring
		state = false
	} else {
		// drawing
		state = true
	}

	startLine := -1 // Index where to begin drawing (where a line begins)
	endLine := -1   // Index where to stop drawing (where a line ends)
	lastk := -1     // Holds last value of the character position

	k := 0
	for i := 0; i < length; {
		// Get next codepoint from byte string and glyph index in font
		codepointByteCount := 0
		codepoint := int32(text[i]) //getCodepoint([]byte(text), i, &codepointByteCount)
		codepointByteCount = 1

		index := rl.GetGlyphIndex(DefaultFont, codepoint)

		// NOTE: Normally we exit the decoding sequence as soon as a bad byte is found (and return 0x3f)
		// but we need to draw all of the bad bytes using the '?' symbol moving one byte
		if codepoint == 0x3f {
			codepointByteCount = 1
		}
		i += (codepointByteCount - 1)

		var glyphWidth float32 = 0
		if codepoint != '\n' {

			if unsafe.Slice(DefaultFont.Chars, DefaultFont.CharsCount)[index].AdvanceX == 0 {
				glyphWidth = unsafe.Slice(DefaultFont.Recs, DefaultFont.CharsCount)[index].Width * scaleFactor
			} else {
				glyphWidth = float32(unsafe.Slice(DefaultFont.Chars, DefaultFont.CharsCount)[index].AdvanceX) * scaleFactor
			}

			if i+1 < length {
				glyphWidth = glyphWidth + spacing
			}
		}

		// NOTE: When wordWrap is ON we first measure how much of the text we can draw before going outside of the rec container
		// We store this info in startLine and endLine, then we change states, draw the text between those two variables
		// and change states again and again recursively until the end of the text (or until we get outside of the container).
		// When wordWrap is OFF we don't need the measure state so we go to the drawing state immediately
		// and begin drawing on the next line before we can get outside the container.
		if state == false {
			// TODO: There are multiple types of spaces in UNICODE, maybe it's a good idea to add support for more
			// Ref: http://jkorpela.fi/chars/spaces.html
			if (codepoint == ' ') || (codepoint == '\t') || (codepoint == '\n') {
				endLine = i
			}

			if (textOffsetX + glyphWidth) > rec.Width {
				if endLine < 1 {
					endLine = i
				} else {
					endLine = endLine
				}

				if i == endLine {
					endLine -= codepointByteCount
				}
				if (startLine + codepointByteCount) == endLine {
					endLine = (i - codepointByteCount)
				}

				state = !state
			} else if (i + 1) == length {
				endLine = i
				state = !state
			} else if codepoint == '\n' {
				state = !state
			}

			if state == true {
				textOffsetX = 0
				i = startLine
				glyphWidth = 0

				// Save character position when we switch states
				tmp := lastk
				lastk = k - 1
				k = tmp
			}
		} else {
			if codepoint == '\n' {
				if !wordWrap {
					textOffsetY += float32(DefaultFont.BaseSize+DefaultFont.BaseSize/2) * scaleFactor
					textOffsetX = 0
				}
			} else {
				if !wordWrap && ((textOffsetX + glyphWidth) > rec.Width) {
					textOffsetY += float32(DefaultFont.BaseSize+DefaultFont.BaseSize/2) * scaleFactor
					textOffsetX = 0
				}

				// When text overflows rectangle height limit, just stop drawing
				if (textOffsetY + float32(DefaultFont.BaseSize)*scaleFactor) > rec.Height {
					break
				}

				// Draw selection background
				// isGlyphSelected := false
				if (selectStart >= 0) && (int32(k) >= selectStart) && (int32(k) < (selectStart + selectLength)) {
					rl.DrawRectangleRec(rl.NewRectangle(rec.X+textOffsetX-1, rec.Y+textOffsetY, glyphWidth, float32(DefaultFont.BaseSize)*scaleFactor), SelectedBackTint)
					// isGlyphSelected = true
				}

				if drawCursor && int32(k) == selectStart {
					rl.DrawRectangleRec(rl.NewRectangle(rec.X+textOffsetX-1-(spacing/2), rec.Y+textOffsetY, spacing*2, float32(DefaultFont.BaseSize)*scaleFactor), rl.Black)
				}

				// Draw current character glyph
				if (codepoint != ' ') && (codepoint != '\t') {
					selcol := rl.Black
					// isGlyphSelected ? selectTint : tint

					rl.DrawTextEx(DefaultFont, string(rune(codepoint)), rl.Vector2{X: rec.X + textOffsetX, Y: rec.Y + textOffsetY}, size, spacing, selcol)
					// rl.DrawTextCodepoint(DefaultFont, codepoint, rl.Vector2{X: rec.X + textOffsetX, Y: rec.Y + textOffsetY}, size, selcol)

				}
			}

			if wordWrap && (i == endLine) {
				textOffsetY += float32(DefaultFont.BaseSize+DefaultFont.BaseSize/2) * scaleFactor
				textOffsetX = 0
				startLine = endLine
				endLine = -1
				glyphWidth = 0
				selectStart += int32(lastk - k)
				k = lastk

				state = !state
			}
		}

		if (textOffsetX != 0) || (codepoint != ' ') {
			textOffsetX += glyphWidth // avoid leading spaces
		}

		i++
		k++
	}
	if drawCursor && selectStart == int32(len(text)) {
		rl.DrawRectangleRec(rl.NewRectangle(rec.X+textOffsetX-1, rec.Y+textOffsetY, spacing*2, float32(DefaultFont.BaseSize)*scaleFactor), rl.Black)
	}
}
