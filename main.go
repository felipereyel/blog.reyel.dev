package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

func main() {
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
