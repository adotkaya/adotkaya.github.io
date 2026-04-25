package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Site struct {
	Config     Config
	Experience []Experience
	Projects   []Project
	Books      []Book
	Tools      []Tool
	Year       int
}

type Config struct {
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Bio      string `yaml:"bio"`
	GitHub   string `yaml:"github"`
	LinkedIn string `yaml:"linkedin"`
	Location string `yaml:"location"`
	About    string `yaml:"about"`
}

type Project struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	URL         string   `yaml:"url"`
	TechStack   []string `yaml:"tech_stack"`
}

type Book struct {
	Title  string `yaml:"title"`
	Author string `yaml:"author"`
}

type Experience struct {
	Company string   `yaml:"company"`
	Role    string   `yaml:"role"`
	Period  string   `yaml:"period"`
	Points  []string `yaml:"points"`
}

type Tool struct {
	Name string `yaml:"name"`
	Icon string `yaml:"icon"`
}

func contentPath(filename string) string {
	primary := filepath.Join("content", filename)
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	return filepath.Join("content-example", filename)
}

func main() {
	var site Site
	site.Year = time.Now().Year()

	site.Config = mustParseYAML[Config](contentPath("config.yaml"))
	site.Experience = mustParseYAML[[]Experience](contentPath("experience.yaml"))
	site.Projects = mustParseYAML[[]Project](contentPath("projects.yaml"))
	site.Books = mustParseYAML[[]Book](contentPath("books.yaml"))
	site.Tools = mustParseYAML[[]Tool](contentPath("tools.yaml"))

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	if err := os.MkdirAll("public", 0755); err != nil {
		log.Fatalf("Error creating public dir: %v", err)
	}

	out, err := os.Create("public/index.html")
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, site); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	if err := copyStatic("static", "public"); err != nil {
		log.Fatalf("Error copying static files: %v", err)
	}

	log.Println("Site generated successfully in public/")
}

func mustParseYAML[T any](path string) T {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading %s: %v", path, err)
	}

	var v T
	if err := yaml.Unmarshal(data, &v); err != nil {
		log.Fatalf("Error parsing %s: %v", path, err)
	}
	return v
}

func copyStatic(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}