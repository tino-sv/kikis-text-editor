package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text-editor/editor"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/fatih/color"
)

func highlightAndPrint(code, filename string) {
	// Split code into lines for numbering
	lines := strings.Split(code, "\n")
	
	// Calculate padding for line numbers based on total lines
	padding := len(fmt.Sprintf("%d", len(lines)))
	
	fmt.Println() // Add a newline before content
	
	// Print each line with its number
	for i, line := range lines {
		// Print line number with consistent padding
		lineNum := color.New(color.FgYellow).Sprintf("%*d", padding, i+1)
		fmt.Printf("%s │ ", lineNum)
		
		// Detect language and highlight the line
		var lexer chroma.Lexer
		if filename != "" && len(filename) > 0 {
			ext := strings.ToLower(filepath.Ext(filename))
			if ext != "" {
				lexer = lexers.Get(ext[1:])
			}
		}
		if lexer == nil {
			lexer = lexers.Fallback
		}

		iterator, err := lexer.Tokenise(nil, line)
		if err != nil {
			fmt.Println(line)
			continue
		}

		// Print highlighted tokens for this line
		for _, token := range iterator.Tokens() {
			c := color.New()
			switch token.Type.String() {
			case "Keyword", "KeywordType", "KeywordDeclaration":
				c.Add(color.FgMagenta, color.Bold)
			case "NameFunction", "NameClass":
				c.Add(color.FgGreen, color.Bold)
			case "LiteralString", "LiteralStringDouble", "LiteralStringSingle":
				c.Add(color.FgGreen)
			case "LiteralNumber", "LiteralNumberInteger":
				c.Add(color.FgYellow)
			case "Comment", "CommentSingle", "CommentMultiline":
				c.Add(color.FgCyan)
			case "Operator", "Punctuation":
				c.Add(color.FgBlue)
			case "NameBuiltin":
				c.Add(color.FgRed)
			default:
				c.Add(color.FgWhite)
			}
			c.Print(token.Value)
		}
		fmt.Println() // New line after each code line
	}
	fmt.Println() // Add a newline after content
}

func showFileOptions() {
	fmt.Println("\nFile operations:")
	fmt.Println("1. Open file")
	fmt.Println("2. Create new file")
	fmt.Println("3. Rename file")
	fmt.Println("4. Delete file")
	fmt.Println("5. Launch Editor")
	fmt.Println("6. Exit")
}

func ensureDirectory(path string) error {
	dir := filepath.Dir(path)
	if dir != "." {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func cli() {
	for {
		fmt.Println("\nWelcome to the TUI version of text editor (name pending)!")
		
		// List files in current directory
		files, err := os.ReadDir(".")
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		// Show available files
		fmt.Println("\nAvailable files:")
		for i, file := range files {
			fmt.Printf("%d. %s\n", i+1, file.Name())
		}

		showFileOptions()
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nChoose an option (1-6): ")
		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1": // Open file
			fmt.Print("Enter file number to open: ")
			fileNumStr, _ := reader.ReadString('\n')
			fileNumStr = strings.TrimSpace(fileNumStr)
			if fileNumStr != "" {
				fileNum := 0
				fmt.Sscan(fileNumStr, &fileNum)
				if fileNum > 0 && fileNum <= len(files) {
					currentFile := files[fileNum-1].Name()
					
					// Create and launch editor with file
					ed, err := editor.NewEditor()
					if err != nil {
						fmt.Printf("Error creating editor: %v\n", err)
						continue
					}

					ed.SetFilename(currentFile)
					ed.Run()
				}
			}

		case "2": // Create new file
			fmt.Print("Enter file path (e.g., folder/filename.txt): ")
			currentFile, _ := reader.ReadString('\n')
			currentFile = strings.TrimSpace(currentFile)
			
			// Ensure directory exists
			if err := ensureDirectory(currentFile); err != nil {
				fmt.Printf("Error creating directory: %v\n", err)
				continue
			}

			// Create and launch editor with new file
			ed, err := editor.NewEditor()
			if err != nil {
				fmt.Printf("Error creating editor: %v\n", err)
				continue
			}
			
			ed.SetFilename(currentFile)
			ed.Run()

		case "3": // Rename file
			fmt.Print("Enter file number to rename: ")
			fileNumStr, _ := reader.ReadString('\n')
			fileNumStr = strings.TrimSpace(fileNumStr)
			fileNum := 0
			fmt.Sscan(fileNumStr, &fileNum)
			if fileNum > 0 && fileNum <= len(files) {
				oldName := files[fileNum-1].Name()
				fmt.Printf("Renaming %s\n", oldName)
				fmt.Print("Enter new name: ")
				newName, _ := reader.ReadString('\n')
				newName = strings.TrimSpace(newName)
				err := os.Rename(oldName, newName)
				if err != nil {
					fmt.Printf("Error renaming file: %v\n", err)
				} else {
					fmt.Printf("Renamed %s to %s\n", oldName, newName)
				}
			} else {
				fmt.Printf("Invalid file number\n")
			}

		case "4": // Delete file
			fmt.Print("Enter file number to delete: ")
			fileNumStr, _ := reader.ReadString('\n')
			fileNumStr = strings.TrimSpace(fileNumStr)
			fileNum := 0
			fmt.Sscan(fileNumStr, &fileNum)
			if fileNum > 0 && fileNum <= len(files) {
				fileName := files[fileNum-1].Name()
				fmt.Printf("Are you sure you want to delete %s? (y/n): ", fileName)
				confirm, _ := reader.ReadString('\n')
				confirm = strings.TrimSpace(confirm)
				if confirm == "y" {
					err := os.Remove(fileName)
					if err != nil {
						fmt.Printf("Error deleting file: %v\n", err)
					} else {
						fmt.Printf("Deleted %s\n", fileName)
					}
				}
			}

		case "5": // Launch Editor
			editor, err := NewEditor()
			if err != nil {
				fmt.Printf("Error creating editor: %v\n", err)
				continue
			}
			if e, ok := editor.(interface{ Run() }); ok {
				e.Run()
			} else {
				fmt.Println("Error: editor does not implement Run method")
			}

		case "6": // Exit
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Invalid option")
		}

		fmt.Print("\nReturn to main menu? (y/n): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer != "y" {
			break
		}
	}
}

func NewEditor() (any, any) {
	panic("unimplemented")
}

