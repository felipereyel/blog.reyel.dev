package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

type Page struct {
	Title   string
	Content template.HTML
}

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

	// Parse the template.
	tmpl, err := template.ParseFiles("templates/layout.html")
	if err != nil {
		log.Fatal(err)
	}

	// Walk through the blog directory.
	err = filepath.Walk("blog", func(path string, info os.FileInfo, err error) error {
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
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Convert markdown to html.
			var content bytes.Buffer
			if err := goldmark.Convert(source, &content); err != nil {
				return err
			}

			// Create the new html file in the dist directory.
			destPath := filepath.Join("dist", strings.TrimSuffix(info.Name(), ".md")+".html")
			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			// Create the data for the template.
			title := strings.Title(strings.ReplaceAll(strings.TrimSuffix(info.Name(), ".md"), "-", " "))
			page := Page{
				Title:   title,
				Content: template.HTML(content.String()),
			}

			// Execute the template with the data.
			err = tmpl.Execute(destFile, page)
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join("dist", "index.html"))
			return
		}
		http.FileServer(http.Dir("dist")).ServeHTTP(w, r)
	})

	fmt.Println("Serving files from dist/ on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
