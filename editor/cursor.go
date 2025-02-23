package editor

// Cursor movement and manipulation
func (e *Editor) moveCursor(dx, dy int) {
    newX := e.cursorX + dx
    newY := e.cursorY + dy

    if newY >= 0 && newY < len(e.lines) {
        lineLength := len(e.lines[newY])
        if newX >= 0 && newX <= lineLength {
            e.cursorX = newX
            e.cursorY = newY
        }
    }
}

func (e *Editor) insertRune(r rune) {
    // Save state for undo
    e.addUndo(e.lines[e.cursorY])
    
    // Auto-close brackets
    autoClose := map[rune]rune{
        '(': ')',
        '[': ']',
        '{': '}',
        '"': '"',
        '\'': '\'',
    }
    
    line := e.lines[e.cursorY]
    if closing, ok := autoClose[r]; ok {
        // Insert both opening and closing brackets
        newLine := line[:e.cursorX] + string(r) + string(closing) + line[e.cursorX:]
        e.lines[e.cursorY] = newLine
        e.cursorX++
    } else {
        // Normal character insertion
        newLine := line[:e.cursorX] + string(r) + line[e.cursorX:]
        e.lines[e.cursorY] = newLine
        e.cursorX++
    }
    e.isDirty = true
}

func (e *Editor) insertNewLine() {
    e.addUndo(e.lines[e.cursorY])
    
    currentLine := e.lines[e.cursorY]
    
    // Calculate indentation of current line
    indent := ""
    for _, char := range currentLine {
        if char == ' ' || char == '\t' {
            indent += string(char)
        } else {
            break
        }
    }
    
    // Split line and maintain indentation
    remainingText := currentLine[e.cursorX:]
    e.lines[e.cursorY] = currentLine[:e.cursorX]
    e.lines = append(e.lines[:e.cursorY+1], append([]string{indent + remainingText}, e.lines[e.cursorY+1:]...)...)
    
    e.cursorY++
    e.cursorX = len(indent)
    e.isDirty = true
}

func (e *Editor) backspace() {
    if e.cursorX > 0 || e.cursorY > 0 {
        e.addUndo(e.lines[e.cursorY])
    }
    
    if e.cursorX > 0 {
        line := e.lines[e.cursorY]
        e.lines[e.cursorY] = line[:e.cursorX-1] + line[e.cursorX:]
        e.cursorX--
        e.isDirty = true
    } else if e.cursorY > 0 {
        // Join with previous line
        newX := len(e.lines[e.cursorY-1])
        e.lines[e.cursorY-1] += e.lines[e.cursorY]
        e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY+1:]...)
        e.cursorY--
        e.cursorX = newX
        e.isDirty = true
    }
} 