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
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"github.com/yuin/goldmark"
)

type Page struct {
	Title   string
	Content template.HTML
}

type Post struct {
	Title string
	Date  time.Time
	Slug  string
	Link  string
}

type IndexPage struct {
	Title string
	Posts []Post
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
	if _, err := os.Stat("dist/posts"); os.IsNotExist(err) {
		os.Mkdir("dist/posts", 0755)
	}

	// Parse the templates.
	postTmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		log.Fatal(err)
	}
	indexTmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	var posts []Post
	var markdownFiles []string

	// First walk: collect all markdown files
	err = filepath.Walk("blog", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			markdownFiles = append(markdownFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the title caser
	caser := cases.Title(language.Und, cases.NoLower)

	// Process all markdown files
	for _, path := range markdownFiles {
		info, _ := os.Stat(path)
		if info.Name() == "index.md" { // index.md is no longer processed as a separate file
			continue
		}

		// Read the markdown file.
		source, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Error reading %s: %v", path, err)
			continue
		}

		// Convert markdown to html.
		var content bytes.Buffer
		if err := goldmark.Convert(source, &content); err != nil {
			log.Printf("Error converting %s: %v", path, err)
			continue
		}

		var destPath string
		if strings.HasPrefix(path, "blog/posts/") {
			destPath = filepath.Join("dist/posts", strings.TrimSuffix(info.Name(), ".md")+".html")
		} else {
			destPath = filepath.Join("dist", strings.TrimSuffix(info.Name(), ".md")+".html")
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			log.Printf("Error creating %s: %v", destPath, err)
			continue
		}
		defer destFile.Close()

		// Create the data for the template.
		title := caser.String(strings.ReplaceAll(strings.TrimSuffix(info.Name(), ".md"), "-", " "))
		page := Page{
			Title:   title,
			Content: template.HTML(content.String()),
		}

		// Execute the template with the data.
		err = postTmpl.Execute(destFile, page)
		if err != nil {
			log.Printf("Error executing template for %s: %v", destPath, err)
			continue
		}

		fmt.Printf("Converted %s to %s\n", path, destPath)

		if strings.HasPrefix(path, "blog/posts/") {
			parts := strings.SplitN(info.Name(), "-", 4)
			if len(parts) == 4 {
				dateStr := strings.Join(parts[:3], "-")
				date, err := time.Parse("2006-01-02", dateStr)
				if err == nil {
					slug := strings.TrimSuffix(parts[3], ".md")
					posts = append(posts, Post{
						Title: caser.String(strings.ReplaceAll(slug, "-", " ")),
						Date:  date,
						Slug:  slug,
						Link:  "/" + filepath.Join("posts", strings.TrimSuffix(info.Name(), ".md")+".html"),
					})
				}
			}
		}
	}

	// Generate index.html
	destPath := filepath.Join("dist", "index.html")
	destFile, err := os.Create(destPath)
	if err != nil {
		log.Fatal(err)
	}
	defer destFile.Close()

	title := "Home"
	page := IndexPage{
		Title: title,
		Posts: posts,
	}

	err = indexTmpl.Execute(destFile, page)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated %s\n", destPath)
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
