package editor

import "github.com/gdamore/tcell/v2"

// Handle all input-related functions
func (e *Editor) handleInput(ev *tcell.EventKey) {
    if ev.Key() == tcell.KeyEscape {
        if e.mode == "command" || e.mode == "search" {
            e.mode = "normal"
            e.commandBuffer = ""
            e.searchTerm = ""
        } else if e.mode == "insert" {
            e.mode = "normal"
            e.setStatusMessage("-- NORMAL MODE --")
        }
        return
    }

    switch e.mode {
    case "normal":
        e.handleNormalMode(ev)
    case "insert":
        e.handleInsertMode(ev)
    case "command":
        e.handleCommandMode(ev)
    case "search":
        e.handleSearchMode(ev)
    }
}

func (e *Editor) handleNormalMode(ev *tcell.EventKey) {
    if e.treeVisible {
        switch ev.Key() {
        case tcell.KeyRune:
            switch ev.Rune() {
            case 'j', 'k', 'h', 'l':
                e.handleTreeNavigation(ev)
                return
            case 't':  // Toggle file tree
                e.treeVisible = !e.treeVisible
                return
            }
        case tcell.KeyEnter:
            e.handleTreeNavigation(ev)
            return
        }
    }

    // Regular editor bindings when tree is not visible
    switch ev.Key() {
    case tcell.KeyRune:
        switch ev.Rune() {
        case 'i':
            e.mode = "insert"
            e.setStatusMessage("-- INSERT MODE --")
        case ':':
            e.mode = "command"
            e.commandBuffer = ""
            e.setStatusMessage(":")
        case '/':
            e.startSearch()
        case 'h':
            e.moveCursor(-1, 0)
        case 'l':
            e.moveCursor(1, 0)
        case 'j':
            e.moveCursor(0, 1)
        case 'k':
            e.moveCursor(0, -1)
        case 'u':
            e.undo()
        case 'r':
            e.redo()
        case 'n':
            e.nextMatch()
        case 'N':
            e.previousMatch()
        case 't':  // Toggle file tree
            e.treeVisible = !e.treeVisible
        case '?':
            e.showHelp()
        }
    }
}

func (e *Editor) handleInsertMode(ev *tcell.EventKey) {
    switch ev.Key() {
    case tcell.KeyTab:
        for i := 0; i < e.tabSize; i++ {
            e.insertRune(' ')
        }
    case tcell.KeyEnter:
        e.insertNewLine()
    case tcell.KeyBackspace, tcell.KeyBackspace2:
        e.backspace()
    case tcell.KeyRune:
        e.insertRune(ev.Rune())
    }
}

 