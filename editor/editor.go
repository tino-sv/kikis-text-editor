// Main editor package
package editor

import (
	"time"

	"github.com/gdamore/tcell/v2"
)




type Editor struct {
	screen           tcell.Screen
	lines            []string
	cursorX, cursorY int
	mode             string
	filename         string
	statusMessage    string
	statusTimeout    time.Time
	isDirty          bool
	tabSize          int
	searchTerm       string
	searchMatches    []struct{ y, x int }
	currentMatch     int
	undoStack        []Action
	redoStack        []Action
	commandBuffer    string
	showLineNumbers  bool
	quit             bool
	treeVisible      bool
	treeWidth        int
	currentPath      string
	fileTree         *FileNode
	treeSelectedLine int
	screenWidth      int
	screenHeight     int
	newFileDir       string
	isWelcomeScreen  bool
	confirmAction    func()
	scrollY          int // Vertical scroll position
}

func (e *Editor) SetFilename(name string) {
	e.filename = name
}

func NewEditor() (*Editor, error) {
	// Initialize screen
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	
	// Enable mouse support
	screen.EnableMouse()
	
	// Get screen dimensions
	width, height := screen.Size()
	
	// Create editor instance
	ed := &Editor{
		screen:          screen,
		lines:           []string{""},
		mode:            "normal",
		tabSize:         4,
		showLineNumbers: true,
		treeVisible:     true,
		treeWidth:       30,
		screenWidth:     width,
		screenHeight:    height,
		undoStack:       make([]Action, 0),
		redoStack:       make([]Action, 0),
		isWelcomeScreen: true,
	}
	
	ed.initFileTree()
	ed.SetStatusMessage("Welcome! Press '?' for help, 'i' for insert mode, ':' for commands")
	
	// Show welcome screen
	ed.showWelcomeScreen()
	
	return ed, nil
}

func (e *Editor) Run() {
	defer e.screen.Fini()

	for {
		e.updateScreenSize()
		e.Draw()
		if e.quit {
			return
		}

		switch ev := e.screen.PollEvent().(type) {
		case *tcell.EventKey:
			e.handleInput(ev)
		case *tcell.EventMouse:
			e.handleMouse(ev)
		case *tcell.EventResize:
			e.screen.Sync()
			e.updateScreenSize()
		}
	}
}

func (e *Editor) updateScreenSize() {
	e.screenWidth, e.screenHeight = e.screen.Size()
}

func (e *Editor) SetStatusMessage(msg string) {
	e.statusMessage = msg
	e.statusTimeout = time.Now().Add(3 * time.Second)
}

func (e *Editor) deleteChar() {
	if e.cursorY >= len(e.lines) || e.cursorX <= 0 || e.cursorX > len(e.lines[e.cursorY]) {
		return
	}
	
	line := e.lines[e.cursorY]
	e.lines[e.cursorY] = line[:e.cursorX-1] + line[e.cursorX:]
	e.cursorX--
	e.isDirty = true
	
	e.undoStack = append(e.undoStack, Action{
		Type: ActionDelete,
		lineNum: e.cursorY,
		oldLine: line,
		newLine: e.lines[e.cursorY],
		cursorX: e.cursorX,
		cursorY: e.cursorY,
		text: string(line[e.cursorX]),
	})
	e.redoStack = nil
}

func (e *Editor) joinLines() {
	if e.cursorY > 0 {
		currentLine := e.lines[e.cursorY]
		prevLine := e.lines[e.cursorY-1]
		e.cursorX = len(prevLine)
		e.lines[e.cursorY-1] = prevLine + currentLine
		e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY+1:]...)
		e.cursorY--
		e.isDirty = true
		
		// Record action for undo
		e.undoStack = append(e.undoStack, Action{
			Type: ActionJoinLines,
			lineNum: e.cursorY,
			oldLine: currentLine,
			newLine: prevLine + currentLine,
			cursorX: e.cursorX,
			cursorY: e.cursorY,
			text: "\n",
		})
		e.redoStack = nil
	}
}

func (e *Editor) showWelcomeScreen() {
	// Only show welcome screen if no file is loaded
	if e.filename != "" {
		return
	}
	
	// Create welcome message with ASCII art logo
	welcome := []string{
		"  _  __ _  _    _  _         ",
		" | |/ /(_)| |_ (_)( )___     ",
		" | ' / | || __|| ||// __|    ",
		" | . \\ | || |_ | |  \\__ \\    ",
		" |_|\\_\\|_| \\__||_|  |___/    ",
		"  _____         _     ______    _ _ _             ",
		" |_   _|       | |   |  ____|  | (_) |            ",
		"   | |  ___ __ | |_  | |__   __| |_| |_ ___  _ __ ",
		"   | | / _ \\\\ \\/ / __| |  __| / _` | | __/ _ \\| '__|",
		"   | ||  __/ >  <\\__ \\| |___| (_| | | || (_) | |   ",
		"   |_| \\___|/_/\\_\\___/|______\\__,_|_|\\__\\___/|_|   ",
		"",
		"Version 0.2",
		"-----------------------------",
		"",
		"Welcome to Kiki's Text Editor!",
		"",
		"Quick Start Guide:",
		"  • Press 'i' to start typing (insert mode)",
		"  • Press 'Esc' to return to normal mode",
		"  • Press ':w filename' to save",
		"  • Press ':q' to quit",
		"  • Press '?' for full help",
		"",
		"The file tree is shown on the left. Press 't' to toggle it.",
		"Press 'n' in the file tree to create a new file.",
		"",
		"Happy editing with Kiki!",
	}
	
	// Set the welcome message as the content
	e.lines = welcome
}


