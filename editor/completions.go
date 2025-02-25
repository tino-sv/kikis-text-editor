package editor

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type Completion struct {
	Text        string
	Description string
}



func (e *Editor) getCompletions() []Completion {
	// Get the word under cursor
	if e.cursorY >= len(e.lines) {
		return nil
	}
	
	line := e.lines[e.cursorY]
	if e.cursorX > len(line) {
		return nil
	}
	
	start := e.cursorX
	for start > 0 && isIdentChar(rune(line[start-1])) {
		start--
	}
	prefix := line[start:e.cursorX]

	// Get language-specific completions
	var completions []Completion
	switch e.detectLanguage() {
	case LangGo:
		completions = goCompletions
	case LangPython:
		completions = pythonCompletions
	case LangJavaScript:
		completions = jsCompletions
	case LangRust:
		completions = rustCompletions
	}

	// Filter completions based on prefix
	if prefix != "" {
		filtered := []Completion{}
		for _, c := range completions {
			if strings.HasPrefix(strings.ToLower(c.Text), strings.ToLower(prefix)) {
				filtered = append(filtered, c)
			}
		}
		return filtered
	}

	return completions
}

func (e *Editor) showCompletions() {
	completions := e.getCompletions()
	if len(completions) == 0 {
		return
	}

	// Draw completion box
	style := tcell.StyleDefault
	boxStyle := style.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	
	// Calculate box dimensions
	maxWidth := 0
	for _, c := range completions {
		width := len(c.Text) + len(c.Description) + 3 // +3 for spacing
		if width > maxWidth {
			maxWidth = width
		}
	}
	
	// Draw box at cursor position
	x := e.cursorX
	y := e.cursorY + 1
	
	// Ensure box stays within screen bounds
	if y+len(completions) >= e.screenHeight {
		y = e.cursorY - len(completions)
	}
	if x+maxWidth >= e.screenWidth {
		x = e.screenWidth - maxWidth - 1
	}
	
	// Draw completions
	for i, c := range completions {
		if y+i >= e.screenHeight {
			break
		}
		text := fmt.Sprintf("%-*s %s", maxWidth-len(c.Description)-1, c.Text, c.Description)
		drawText(e.screen, x, y+i, boxStyle, text)
	}
}

func isIdentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// Language-specific completions
var goCompletions = []Completion{
	// Go keywords
	{Text: "func", Description: "function declaration"},
	{Text: "type", Description: "type declaration"},
	{Text: "struct", Description: "struct declaration"},
	{Text: "interface", Description: "interface declaration"},
	{Text: "map", Description: "map type"},
	{Text: "chan", Description: "channel type"},
	{Text: "if", Description: "if statement"},
	{Text: "else", Description: "else statement"},
	{Text: "for", Description: "for loop"},
	{Text: "range", Description: "range clause"},
	{Text: "return", Description: "return statement"},
	{Text: "package", Description: "package declaration"},
	{Text: "import", Description: "import declaration"},
	{Text: "var", Description: "variable declaration"},
	{Text: "const", Description: "constant declaration"},
	
	// Common types
	{Text: "string", Description: "string type"},
	{Text: "int", Description: "integer type"},
	{Text: "bool", Description: "boolean type"},
	{Text: "error", Description: "error type"},
	{Text: "float64", Description: "64-bit float"},
	
	// Common snippets
	{Text: "fmt.Println()", Description: "print line"},
	{Text: "fmt.Printf()", Description: "print formatted"},
	{Text: "make()", Description: "make builtin"},
	{Text: "append()", Description: "append builtin"},
	{Text: "len()", Description: "length builtin"},
}

var pythonCompletions = []Completion{
	// Python keywords
	{Text: "def", Description: "function definition"},
	{Text: "class", Description: "class definition"},
	{Text: "if", Description: "if statement"},
	{Text: "elif", Description: "else if statement"},
	{Text: "else", Description: "else statement"},
	{Text: "for", Description: "for loop"},
	{Text: "while", Description: "while loop"},
	{Text: "try", Description: "try block"},
	{Text: "except", Description: "except block"},
	{Text: "finally", Description: "finally block"},
	{Text: "import", Description: "import statement"},
	{Text: "from", Description: "from import"},
	{Text: "return", Description: "return statement"},
	
	// Common functions
	{Text: "print()", Description: "print function"},
	{Text: "len()", Description: "length function"},
	{Text: "range()", Description: "range function"},
	{Text: "list()", Description: "list constructor"},
	{Text: "dict()", Description: "dictionary constructor"},
	{Text: "set()", Description: "set constructor"},
}

var jsCompletions = []Completion{
	// JavaScript keywords
	{Text: "function", Description: "function declaration"},
	{Text: "const", Description: "constant declaration"},
	{Text: "let", Description: "block-scoped variable"},
	{Text: "var", Description: "function-scoped variable"},
	{Text: "if", Description: "if statement"},
	{Text: "else", Description: "else statement"},
	{Text: "for", Description: "for loop"},
	{Text: "while", Description: "while loop"},
	{Text: "class", Description: "class declaration"},
	{Text: "return", Description: "return statement"},
	{Text: "import", Description: "import statement"},
	{Text: "export", Description: "export statement"},
	
	// Common methods
	{Text: "console.log()", Description: "console log"},
	{Text: "console.error()", Description: "console error"},
	{Text: "Array.from()", Description: "array from iterable"},
	{Text: "Object.keys()", Description: "object keys"},
	{Text: "Promise", Description: "Promise constructor"},
	{Text: "async", Description: "async function"},
	{Text: "await", Description: "await expression"},
}

var rustCompletions = []Completion{
	// Rust keywords
	{Text: "fn", Description: "function declaration"},
	{Text: "struct", Description: "struct declaration"},
	{Text: "enum", Description: "enum declaration"},
	{Text: "impl", Description: "implementation block"},
	{Text: "trait", Description: "trait declaration"},
	{Text: "let", Description: "variable binding"},
	{Text: "mut", Description: "mutable binding"},
	{Text: "if", Description: "if expression"},
	{Text: "else", Description: "else expression"},
	{Text: "match", Description: "match expression"},
	{Text: "loop", Description: "loop expression"},
	{Text: "while", Description: "while loop"},
	{Text: "for", Description: "for loop"},
	
	// Common macros
	{Text: "println!()", Description: "print line macro"},
	{Text: "vec![]", Description: "vector macro"},
	{Text: "Some()", Description: "Some variant"},
	{Text: "None", Description: "None variant"},
	{Text: "Ok()", Description: "Ok variant"},
	{Text: "Err()", Description: "Err variant"},
}
