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
	foldedLines      map[int]int
	treeVisible      bool
	treeWidth        int
	currentPath      string
	fileTree         *FileNode
	treeSelectedLine int
	screenWidth      int
	screenHeight     int
}

func (e *Editor) SetFilename(name string) {
	e.filename = name
}

func NewEditor() (*Editor, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	ed := &Editor{
		screen:          screen,
		lines:           []string{""},
		mode:            "normal",
		tabSize:         4,
		showLineNumbers: true,
		treeVisible:     true,
		treeWidth:       30,
		undoStack:       make([]Action, 0),
		redoStack:       make([]Action, 0),
	}
	
	// Initialize file tree
	ed.initFileTree()
	
	return ed, nil
}

func (e *Editor) Run() {
	defer e.screen.Fini()

	for {
		e.draw()
		if e.quit {
			return
		}

		switch ev := e.screen.PollEvent().(type) {
		case *tcell.EventKey:
			e.handleInput(ev)
		case *tcell.EventResize:
			e.screen.Sync()
		}
	}
}

func (e *Editor) updateScreenSize() {
	e.screenWidth, e.screenHeight = e.screen.Size()
}
