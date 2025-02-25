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
func (e *Editor) Draw() {
	// Clear screen only once per frame
	e.screen.Clear()
	
	// Calculate content area
	contentStartX := 0
	
	// Draw file tree if visible
	if e.treeVisible {
		e.drawFileTree()
		contentStartX = e.treeWidth + 1 // Add 1 for separator
	}
	
	// Calculate visible region based on scroll position
	startLine := e.scrollY
	endLine := startLine + e.screenHeight - 2 // Account for status bars
	
	if endLine > len(e.lines) {
		endLine = len(e.lines)
	}
	
	// Draw only visible content
	for y := startLine; y < endLine; y++ {
		screenY := y - startLine
		lineNum := y + 1
		
		// Draw line numbers if enabled
		if e.showLineNumbers {
			lineNumStr := fmt.Sprintf("%4d ", lineNum)
			lineNumStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
			drawText(e.screen, contentStartX, screenY, lineNumStyle, lineNumStr)
		}
		
		// Calculate line offset based on line numbers
		xOffset := contentStartX
		if e.showLineNumbers {
			xOffset += 5
		}
		
		// Draw the line content with syntax highlighting
		if y < len(e.lines) {
			line := e.lines[y]
			styles := e.highlightSyntax(line, y)
			
			// Draw each character with its style
			for x, r := range line {
				if x < len(styles) {
					e.screen.SetContent(xOffset+x, screenY, r, nil, styles[x])
				}
			}
		}
	}
	
	// Draw status bars
	e.drawStatusBar()
	e.drawMessageBar()
	
	// Position cursor
	cursorX := e.cursorX
	if e.showLineNumbers {
		cursorX += 5
	}
	if e.treeVisible {
		cursorX += e.treeWidth + 1
	}
	
	// Only show cursor if it's in the visible area
	if e.cursorY >= e.scrollY && e.cursorY < e.scrollY + e.screenHeight - 2 {
		e.screen.ShowCursor(cursorX, e.cursorY - e.scrollY)
	}
	
	// Update screen in one go
	e.screen.Show()
}

func drawText(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	for _, r := range text {
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

func (e *Editor) drawStatusBar() {
	// Create a more informative status bar
	statusStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	modeStyle := tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite).Bold(true)
	
	// Clear the status bar
	for x := 0; x < e.screenWidth; x++ {
		e.screen.SetContent(x, e.screenHeight-2, ' ', nil, statusStyle)
	}
	
	// Left side: filename, modified status, and mode
	filename := e.filename
	if filename == "" {
		filename = "[No Name]"
	}
	
	modified := ""
	if e.isDirty {
		modified = " [+]"
	}
	
	modeText := fmt.Sprintf(" %s ", strings.ToUpper(e.mode))
	
	// Right side: line/column info
	lineInfo := fmt.Sprintf("Ln %d, Col %d ", e.cursorY+1, e.cursorX+1)
	
	// Draw mode with special highlighting
	drawText(e.screen, 0, e.screenHeight-2, modeStyle, modeText)
	
	// Draw filename and modified status
	fileStatus := fmt.Sprintf(" %s%s ", filename, modified)
	drawText(e.screen, len(modeText), e.screenHeight-2, statusStyle, fileStatus)
	
	// Draw line info on the right
	drawText(e.screen, e.screenWidth-len(lineInfo), e.screenHeight-2, statusStyle, lineInfo)
	
	// Draw language type if known
	langNames := []string{"Go", "Python", "JavaScript", "Rust", "Unknown"}
	if e.filename != "" {
		lang := e.detectLanguage()
		if lang >= 0 && lang < len(langNames) {
			langText := fmt.Sprintf(" %s ", langNames[lang])
			drawText(e.screen, e.screenWidth-len(lineInfo)-len(langText), e.screenHeight-2, statusStyle, langText)
		}
	}
}

func (e *Editor) showHelp() {
	helpText := []string{
		"Kiki's Text Editor Help",
		"---------------------",
		"",
		"Getting Started:",
		"  i       - Enter insert mode (for typing)",
		"  Esc     - Return to normal mode",
		"  :       - Enter command mode",
		"  ?       - Show this help",
		"",
		"Navigation:",
		"  h,j,k,l - Move cursor (left, down, up, right)",
		"  t       - Toggle file tree",
		"",
		"Editing:",
		"  i       - Start typing (insert mode)",
		"  Tab     - Show code completions (in insert mode)",
		"  u       - Undo",
		"  r       - Redo",
		"",
		"File Tree:",
		"  j,k     - Move up/down",
		"  Enter   - Open file/folder",
		"  n       - Create new file",
		"  D       - Delete file",
		"  r       - Rename file",
		"",
		"File Operations:",
		"  :w      - Save file",
		"  :q      - Quit",
		"  :wq     - Save and quit",
		"",
		"Search:",
		"  /       - Start search",
		"  n       - Next match",
		"  N       - Previous match",
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

func (e *Editor) drawMessageBar() {
	// Clear the message bar
	for x := 0; x < e.screenWidth; x++ {
		e.screen.SetContent(x, e.screenHeight-1, ' ', nil, tcell.StyleDefault)
	}
	
	// Show status message if it exists
	if e.statusMessage != "" && time.Now().Before(e.statusTimeout) {
		drawText(e.screen, 0, e.screenHeight-1, tcell.StyleDefault, e.statusMessage)
		return
	}
	
	// Otherwise show context-sensitive key hints
	hintStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
	var hints string
	
	switch e.mode {
	case "normal":
		if e.treeVisible {
			hints = "j/k:navigate  Enter:open  n:new  D:delete  r:rename  t:hide  ?:help"
		} else {
			hints = "i:insert  /:search  t:files  :w:save  :q:quit  ?:help"
		}
	case "insert":
		hints = "Tab:complete  Esc:normal mode"
	case "command":
		hints = "Enter:execute  Esc:cancel"
	case "search":
		hints = "Enter:find  n:next  N:prev  Esc:cancel"
	case "filename":
		hints = "Enter:create  Esc:cancel"
	case "rename":
		hints = "Enter:rename  Esc:cancel"
	}
	
	// Truncate if too long
	if len(hints) > e.screenWidth && e.screenWidth > 3 {
		hints = hints[:e.screenWidth-3] + "..."
	}
	
	drawText(e.screen, 0, e.screenHeight-1, hintStyle, hints)
} 