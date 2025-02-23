package editor

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

const VERSION = "v0.1"

// Display-related functionality
func (e *Editor) draw() {
	e.updateScreenSize()
	e.screen.Clear()
	
	// Calculate visible range
	startLine := 0
	if e.cursorY >= e.screenHeight {
		startLine = e.cursorY - e.screenHeight + 2
	}
	endLine := startLine + e.screenHeight - 1
	if endLine > len(e.lines) {
		endLine = len(e.lines)
	}

	// Draw visible lines
	contentOffset := 0
	if e.treeVisible {
		e.drawFileTree()
		contentOffset = e.treeWidth + 1
		// Draw vertical separator
		for y := 0; y < e.screenHeight; y++ {
			e.screen.SetContent(e.treeWidth, y, '│', nil, tcell.StyleDefault)
		}
	}

	for i, y := startLine, 0; i < endLine; i, y = i+1, y+1 {
		if e.showLineNumbers {
			lineNum := fmt.Sprintf("%3d ", i+1)
			drawText(e.screen, contentOffset, y, tcell.StyleDefault.Foreground(tcell.ColorYellow), lineNum)
			contentOffset += 4
		}

		if i < len(e.lines) {
			line := e.lines[i]
			styles := e.syntaxStyle(line)
			for x, char := range line {
				style := styles[x]
				
				// Highlight search matches
				if len(e.searchTerm) > 0 {
					if x < len(line)-len(e.searchTerm)+1 && strings.HasPrefix(line[x:], e.searchTerm) {
						if e.currentMatch < len(e.searchMatches) &&
							e.searchMatches[e.currentMatch].y == y &&
							e.searchMatches[e.currentMatch].x == x {
							style = style.Background(tcell.ColorDarkGreen)
						} else {
							style = style.Background(tcell.ColorDarkGray)
						}
					}
				}
				
				e.screen.SetContent(x+contentOffset, y, char, nil, style)
			}
		}
		
		if e.showLineNumbers {
			contentOffset -= 4
		}
	}

	// Draw scrollbar if needed
	if len(e.lines) > e.screenHeight {
		scrollbarPos := (startLine * e.screenHeight) / len(e.lines)
		scrollbarHeight := (e.screenHeight * e.screenHeight) / len(e.lines)
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}
		
		for y := 0; y < e.screenHeight; y++ {
			char := '│'
			if y >= scrollbarPos && y < scrollbarPos+scrollbarHeight {
				char = '█'
			}
			e.screen.SetContent(e.screenWidth-1, y, char, nil, tcell.StyleDefault)
		}
	}

	// Draw status line
	e.drawStatus()

	// Adjust cursor position for scrolling
	cursorY := e.cursorY - startLine
	cursorOffset := contentOffset
	if e.showLineNumbers {
		cursorOffset += 4
	}
	e.screen.ShowCursor(e.cursorX+cursorOffset, cursorY)
	e.screen.Show()
}

func drawText(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	for _, r := range []rune(text) {
		screen.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
}

func (e *Editor) setStatusMessage(msg string) {
	e.statusMessage = msg
	e.statusTimeout = time.Now().Add(3 * time.Second)
}

func (e *Editor) syntaxStyle(line string) []tcell.Style {
	styles := make([]tcell.Style, len(line))
	defaultStyle := tcell.StyleDefault
	
	// Define syntax styles
	keywordStyle := defaultStyle.Foreground(tcell.ColorPurple)
	typeStyle := defaultStyle.Foreground(tcell.ColorTeal)
	stringStyle := defaultStyle.Foreground(tcell.ColorGreen)
	numberStyle := defaultStyle.Foreground(tcell.ColorRed)
	commentStyle := defaultStyle.Foreground(tcell.ColorGray)
	functionStyle := defaultStyle.Foreground(tcell.ColorYellow)
	operatorStyle := defaultStyle.Foreground(tcell.ColorOrange)

	// Language keywords
	keywords := map[string]bool{
		"func": true, "return": true, "if": true, "else": true,
		"for": true, "range": true, "break": true, "continue": true,
		"switch": true, "case": true, "default": true,
		"package": true, "import": true, "type": true,
		"var": true, "const": true, "struct": true, "interface": true,
		"map": true, "chan": true, "go": true, "defer": true,
	}

	// Types
	types := map[string]bool{
		"string": true, "int": true, "bool": true, "float": true,
		"byte": true, "rune": true, "error": true, "uint": true,
		"int8": true, "int16": true, "int32": true, "int64": true,
		"uint8": true, "uint16": true, "uint32": true, "uint64": true,
	}

	// Fill with default style first
	for i := range styles {
		styles[i] = defaultStyle
	}

	inString := false
	inComment := false
	word := ""

	for i, char := range line {
		if inComment {
			styles[i] = commentStyle
			continue
		}

		if i < len(line)-1 && line[i:i+2] == "//" {
			inComment = true
			styles[i] = commentStyle
			continue
		}

		if char == '"' && (i == 0 || line[i-1] != '\\') {
			inString = !inString
			styles[i] = stringStyle
			continue
		}

		if inString {
			styles[i] = stringStyle
			continue
		}

		// Handle numbers
		if unicode.IsDigit(char) || (char == '.' && i+1 < len(line) && unicode.IsDigit(rune(line[i+1]))) {
			styles[i] = numberStyle
			continue
		}

		// Handle operators
		if strings.ContainsRune("+-*/%=<>!&|^~:;,.", char) {
			styles[i] = operatorStyle
			continue
		}

		// Handle words (keywords, types, functions)
		if unicode.IsLetter(char) || char == '_' {
			word += string(char)
		} else if word != "" {
			style := defaultStyle
			if keywords[word] {
				style = keywordStyle
			} else if types[word] {
				style = typeStyle
			} else if i < len(line) && line[i] == '(' {
				style = functionStyle
			}

			// Apply style to the whole word
			for j := i - len(word); j < i; j++ {
				styles[j] = style
			}
			word = ""
		}
	}

	return styles
}

func (e *Editor) drawStatus() {
	style := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	
	// Get status text
	statusText := fmt.Sprintf("-- %s -- %s %s", e.mode, e.filename, VERSION)
	if e.isDirty {
		statusText += " [+]"
	}
	
	// Show temporary status message if active
	if e.statusMessage != "" && time.Now().Before(e.statusTimeout) {
		statusText = e.statusMessage
	}
	
	// Draw status at bottom of screen
	drawText(e.screen, 0, e.screenHeight-1, style, statusText)
}

func (e *Editor) showHelp() {
	helpText := []string{
		"Editor Help",
		"-----------",
		"Normal Mode:",
		"  h,j,k,l - Move cursor",
		"  i       - Enter insert mode",
		"  :       - Enter command mode",
		"  /       - Search",
		"  n/N     - Next/Previous search match",
		"  u       - Undo",
		"  r       - Redo",
		"  t       - Toggle file tree",
		"",
		"File Tree:",
		"  h,j,k,l - Navigate",
		"  Enter   - Open file/Toggle directory",
		"",
		"Command Mode:",
		"  :w      - Save file",
		"  :q      - Quit",
		"  :wq     - Save and quit",
		"  :q!     - Force quit",
		"  :set number/nonumber - Toggle line numbers",
		"",
		"Press any key to close help",
	}

	// Save current screen
	e.screen.Clear()
	
	// Center help text
	for i, line := range helpText {
		x := (e.screenWidth - len(line)) / 2
		if x < 0 {
			x = 0
		}
		drawText(e.screen, x, i, tcell.StyleDefault.Foreground(tcell.ColorWhite), line)
	}
	
	e.screen.Show()
	
	// Wait for keypress
	for {
		ev := e.screen.PollEvent()
		switch ev.(type) {
		case *tcell.EventKey:
			return
		}
	}
} 