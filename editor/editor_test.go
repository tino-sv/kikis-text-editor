package editor

import (
	"testing"
)

func TestNewEditor(t *testing.T) {
	ed, err := NewEditor()
	if err != nil {
		t.Fatalf("Failed to create editor: %v", err)
	}
	if ed == nil {
		t.Fatal("Editor is nil")
	}
	if ed.mode != "normal" {
		t.Errorf("Expected initial mode to be 'normal', got '%s'", ed.mode)
	}
}

func TestInsertRune(t *testing.T) {
	ed, _ := NewEditor()
	ed.insertRune('a')
	if len(ed.lines) == 0 || ed.lines[0] != "a" {
		t.Errorf("Expected line to contain 'a', got '%s'", ed.lines[0])
	}
}

func TestUndo(t *testing.T) {
	ed, _ := NewEditor()
	ed.insertRune('a')
	ed.undo()
	if len(ed.lines) == 0 || ed.lines[0] != "" {
		t.Errorf("Expected empty line after undo, got '%s'", ed.lines[0])
	}
} 