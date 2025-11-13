package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "generate":
		generate()
	case "serve":
		serve()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: go run main.go <command>")
	fmt.Println("Commands:")
	fmt.Println("  generate - Convert markdown files to html")
	fmt.Println("  serve    - Serve the html files in dist/")
}

func generate() {
	// Ensure the dist directory exists.
	if _, err := os.Stat("dist"); os.IsNotExist(err) {
		os.Mkdir("dist", 0755)
	}

	// Walk through the blog directory.
	err := filepath.Walk("blog", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories.
		if info.IsDir() {
			return nil
		}

		// Check for markdown files.
		if strings.HasSuffix(info.Name(), ".md") {
			// Read the markdown file.
			source, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			// Convert markdown to html.
			var buf bytes.Buffer
			if err := goldmark.Convert(source, &buf); err != nil {
				return err
			}

			// Create the new html file in the dist directory.
			destPath := filepath.Join("dist", strings.TrimSuffix(info.Name(), ".md")+ ".html")
			err = ioutil.WriteFile(destPath, buf.Bytes(), 0644)
			if err != nil {
				return err
			}

			fmt.Printf("Converted %s to %s\n", path, destPath)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func serve() {
	fs := http.FileServer(http.Dir("dist"))
	http.Handle("/", fs)

	fmt.Println("Serving files from dist/ on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
