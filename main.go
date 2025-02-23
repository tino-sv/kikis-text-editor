package main

import (
	"fmt"
	"os"
	"text-editor/editor"
)

func main() {
	ed, err := editor.NewEditor()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating editor: %v\n", err)
		os.Exit(1)
	}

	// If a file path is provided, open it
	if len(os.Args) > 1 {
		ed.SetFilename(os.Args[1])
	}

	// Run the editor (will show file tree by default)
	ed.Run()
}







