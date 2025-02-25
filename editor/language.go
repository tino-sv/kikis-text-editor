package editor

import (
	"path/filepath"
	"strings"
)

// Language constants
const (
	LangGo = iota
	LangPython
	LangJavaScript
	LangRust
	LangUnknown
)

func (e *Editor) detectLanguage() int {
	if e.filename == "" {
		return LangGo // default to Go
	}
	
	ext := strings.ToLower(filepath.Ext(e.filename))
	switch ext {
	case ".go":
		return LangGo
	case ".py":
		return LangPython
	case ".js", ".jsx", ".ts", ".tsx":
		return LangJavaScript
	case ".rs":
		return LangRust
	default:
		return LangUnknown
	}
} 