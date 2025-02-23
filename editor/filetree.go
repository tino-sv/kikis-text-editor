package editor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type FileNode struct {
    name     string
    isDir    bool
    expanded bool
    children []*FileNode
    parent   *FileNode
}
func (e *Editor) initFileTree() {
    e.treeVisible = true
    e.treeWidth = 30
    e.currentPath, _ = os.Getwd()
    e.refreshFileTree()
}

func (e *Editor) refreshFileTree() {
    root := &FileNode{
        name:     e.currentPath,
        isDir:    true,
        expanded: true,
    }
    e.loadDirectory(root)
    e.fileTree = root
}

func (e *Editor) loadDirectory(node *FileNode) {
    entries, err := os.ReadDir(node.name)
    if err != nil {
        return
    }

    for _, entry := range entries {
        if entry.Name()[0] == '.' { // Skip hidden files
            continue
        }

        child := &FileNode{
            name:     filepath.Join(node.name, entry.Name()),
            isDir:    entry.IsDir(),
            expanded: false,
            parent:   node,
        }
        node.children = append(node.children, child)
    }

    // Sort directories first, then files
    sort.Slice(node.children, func(i, j int) bool {
        if node.children[i].isDir != node.children[j].isDir {
            return node.children[i].isDir
        }
        return node.children[i].name < node.children[j].name
    })
}

func (e *Editor) drawFileTree() {
    style := tcell.StyleDefault
    dirStyle := style.Foreground(tcell.ColorYellow)
    fileStyle := style.Foreground(tcell.ColorWhite)
    selectedStyle := style.Background(tcell.ColorDarkBlue)

    y := 0
    e.drawTreeNode(e.fileTree, 0, &y, dirStyle, fileStyle, selectedStyle)
}

func (e *Editor) drawTreeNode(node *FileNode, depth int, y *int, dirStyle, fileStyle, selectedStyle tcell.Style) {
    if *y >= e.screenHeight {
        return
    }

    // Draw the current node
    prefix := strings.Repeat("  ", depth)
    if node.isDir {
        if node.expanded {
            prefix += "▼ "
        } else {
            prefix += "▶ "
        }
    } else {
        prefix += "  "
    }

    name := filepath.Base(node.name)
    style := fileStyle
    if node.isDir {
        style = dirStyle
    }
    if *y == e.treeSelectedLine {
        style = selectedStyle
    }

    drawText(e.screen, 0, *y, style, prefix+name)
    *y++

    // Draw children if expanded
    if node.expanded {
        for _, child := range node.children {
            e.drawTreeNode(child, depth+1, y, dirStyle, fileStyle, selectedStyle)
        }
    }
}


// Add tree navigation methods
func (e *Editor) handleTreeNavigation(ev *tcell.EventKey) {
    switch ev.Key() {
    case tcell.KeyRune:
        switch ev.Rune() {
        case 'j': // Move down
            e.treeSelectedLine++
        case 'k': // Move up
            if e.treeSelectedLine > 0 {
                e.treeSelectedLine--
            }
        case 'h': // Collapse directory
            node := e.getSelectedNode()
            if node != nil && node.isDir {
                node.expanded = false
            }
        case 'l': // Expand directory or open file
            node := e.getSelectedNode()
            if node != nil {
                if node.isDir {
                    node.expanded = true
                    if len(node.children) == 0 {
                        e.loadDirectory(node)
                    }
                } else {
                    e.SetFilename(node.name)
                    if err := e.loadFile(node.name); err == nil {
                        e.treeVisible = false
                    }
                }
            }
        }
    case tcell.KeyEnter: // Open file or toggle directory
        node := e.getSelectedNode()
        if node != nil {
            if node.isDir {
                node.expanded = !node.expanded
                if node.expanded && len(node.children) == 0 {
                    e.loadDirectory(node)
                }
            } else {
                e.SetFilename(node.name)
                if err := e.loadFile(node.name); err == nil {
                    e.treeVisible = false
                }
            }
        }
    }
}

func (e *Editor) getSelectedNode() *FileNode {
    y := 0
    return e.findNodeAtLine(e.fileTree, &y)
}

func (e *Editor) findNodeAtLine(node *FileNode, y *int) *FileNode {
    if *y == e.treeSelectedLine {
        return node
    }
    *y++
    
    if node.expanded {
        for _, child := range node.children {
            if found := e.findNodeAtLine(child, y); found != nil {
                return found
            }
        }
    }
    return nil
}

// Add this new method to load file contents
func (e *Editor) loadFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        e.setStatusMessage(fmt.Sprintf("Error opening file: %v", err))
        return err
    }
    defer file.Close()

    e.lines = []string{}
    scanner := bufio.NewScanner(file)
    
    // Handle large files
    buf := make([]byte, 0, 64*1024)
    scanner.Buffer(buf, 1024*1024)
    
    for scanner.Scan() {
        e.lines = append(e.lines, scanner.Text())
    }
    
    if err := scanner.Err(); err != nil {
        e.setStatusMessage(fmt.Sprintf("Error reading file: %v", err))
        return err
    }
    
    if len(e.lines) == 0 {
        e.lines = append(e.lines, "")
    }
    
    e.cursorX = 0
    e.cursorY = 0
    e.isDirty = false
    e.setStatusMessage(fmt.Sprintf("Loaded %s (%d lines)", filename, len(e.lines)))
    return nil
} 