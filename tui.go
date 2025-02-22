package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/fatih/color"
)

func highlightAndPrint(code, filename string) {
	// Detect language from file extension or use fallback
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

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		fmt.Println(code)
		return
	}

	fmt.Println() // Add a newline before highlighted content
	for _, token := range iterator.Tokens() {
		c := color.New()
	
		// Map token types to terminal colors
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
	fmt.Println() // Add a newline after highlighted content
}

func showFileOptions() {
	fmt.Println("\nFile operations:")
	fmt.Println("1. Open file")
	fmt.Println("2. Create new file")
	fmt.Println("3. Rename file")
	fmt.Println("4. Delete file")
	fmt.Println("5. Exit")
}

func ensureDirectory(path string) error {
	dir := filepath.Dir(path)
	if dir != "." {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func tui() {
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
		fmt.Print("\nChoose an option (1-5): ")
		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1": // Open file
			fmt.Print("Enter file number to open: ")
			fileNumStr, _ := reader.ReadString('\n')
			fileNumStr = strings.TrimSpace(fileNumStr)
			if fileNumStr != "" {
				// Try to open existing file
				fileNum := 0
				fmt.Sscan(fileNumStr, &fileNum)
				if fileNum > 0 && fileNum <= len(files) {
					currentFile := files[fileNum-1].Name()
					data, err := os.ReadFile(currentFile)
					if err == nil {
						currentContent := string(data)
						fmt.Println("\nCurrent content:")
						highlightAndPrint(currentContent, currentFile)
						
						// Convert content to lines for editing
						texts := strings.Split(currentContent, "\n")
						
						fmt.Println("\nEnter new content (type 'save file' to save and quit):")
						for {
							fmt.Print("-> ")
							text, _ := reader.ReadString('\n')
							text = strings.TrimSpace(text)
							if text == "save file" {
								break
							}
							if text == "exit" {
								os.Exit(0)
							}
							texts = append(texts, text)
							highlightAndPrint(strings.Join(texts, "\n"), currentFile)
						}

						// Add save as option after editing
						fmt.Print("Save to different location? (y/n): ")
						saveAs, _ := reader.ReadString('\n')
						if strings.TrimSpace(saveAs) == "y" {
							fmt.Print("Enter new file path: ")
							newPath, _ := reader.ReadString('\n')
							newPath = strings.TrimSpace(newPath)
							
							if err := ensureDirectory(newPath); err != nil {
								fmt.Printf("Error creating directory: %v\n", err)
								continue
							}
							currentFile = newPath
						}

						// Save the changes
						file, err := os.Create(currentFile)
						if err != nil {
							fmt.Printf("Error creating file: %v\n", err)
							continue
						}
						defer file.Close()
						writer := bufio.NewWriter(file)
						for _, text := range texts {
							writer.WriteString(text + "\n")
						}
						writer.Flush()
						fmt.Println("saved:", currentFile)
					}
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

			fmt.Println("\nEnter content (type 'exit file' to save and quit):")
			var texts []string
			for {
				fmt.Print("-> ")
				text, _ := reader.ReadString('\n')
				text = strings.TrimSpace(text)
				if text == "exit file" {
					break
				}
				if text == "exit" {
					os.Exit(0)
				}
				texts = append(texts, text)
				highlightAndPrint(strings.Join(texts, "\n"), currentFile)
			}

			// Save the file
			file, err := os.Create(currentFile)
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				continue
			}
			defer file.Close()
			writer := bufio.NewWriter(file)
			for _, text := range texts {
				writer.WriteString(text + "\n")
			}
			writer.Flush()
			fmt.Println("saved:", currentFile)

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

		case "5": // Exit
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
