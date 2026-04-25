package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"
)

type PageData struct {
	Config     Config
	Experience []Experience
	Projects   []Project
	Books      []Book
	Tools      []Tool
	Posts      []BlogPost
	Post       *BlogPost
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

type BlogPost struct {
	Title   string
	Date    string
	Slug    string
	Excerpt string
	Content template.HTML
	URL     string
}

type blogFrontmatter struct {
	Title string `yaml:"title"`
	Date  string `yaml:"date"`
	Slug  string `yaml:"slug"`
}

func main() {
	data := loadPageData()

	if err := os.MkdirAll("public", 0755); err != nil {
		log.Fatalf("Error creating public dir: %v", err)
	}

	// Home
	renderTemplate("templates/index.html", "public/index.html", data)

	// Resume
	os.MkdirAll("public/resume", 0755)
	renderTemplate("templates/resume.html", "public/resume/index.html", data)

	// Projects
	os.MkdirAll("public/projects", 0755)
	renderTemplate("templates/projects.html", "public/projects/index.html", data)

	// Blog list
	os.MkdirAll("public/blog", 0755)
	renderTemplate("templates/blog-list.html", "public/blog/index.html", data)

	// Blog posts
	for i := range data.Posts {
		postData := data
		postData.Post = &data.Posts[i]
		postDir := filepath.Join("public/blog", data.Posts[i].Slug)
		os.MkdirAll(postDir, 0755)
		renderTemplate("templates/blog-post.html", filepath.Join(postDir, "index.html"), postData)
	}

	// Now
	os.MkdirAll("public/now", 0755)
	renderTemplate("templates/now.html", "public/now/index.html", data)

	// Uses
	os.MkdirAll("public/uses", 0755)
	renderTemplate("templates/uses.html", "public/uses/index.html", data)

	// Static files
	copyStatic("static", "public")

	log.Println("Site generated successfully in public/")
}

func renderTemplate(src, dst string, data PageData) {
	tmpl, err := template.ParseFiles(src)
	if err != nil {
		log.Fatalf("Error parsing %s: %v", src, err)
	}

	f, err := os.Create(dst)
	if err != nil {
		log.Fatalf("Error creating %s: %v", dst, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		log.Fatalf("Error executing %s: %v", src, err)
	}
}

func loadPageData() PageData {
	var data PageData
	data.Year = time.Now().Year()

	data.Config = mustParseYAML[Config](contentPath("config.yaml"))
	data.Experience = mustParseYAML[[]Experience](contentPath("experience.yaml"))
	data.Projects = mustParseYAML[[]Project](contentPath("projects.yaml"))
	data.Books = mustParseYAML[[]Book](contentPath("books.yaml"))
	data.Tools = mustParseYAML[[]Tool](contentPath("tools.yaml"))
	data.Posts = loadBlogPosts()

	return data
}

func loadBlogPosts() []BlogPost {
	entries, err := os.ReadDir("content/blog")
	if err != nil {
		return nil
	}

	var posts []BlogPost
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		post := parseBlogPost(filepath.Join("content/blog", entry.Name()))
		posts = append(posts, post)
	}

	// Sort by date descending (newest first)
	for i := 0; i < len(posts)-1; i++ {
		for j := i + 1; j < len(posts); j++ {
			if posts[i].Date < posts[j].Date {
				posts[i], posts[j] = posts[j], posts[i]
			}
		}
	}

	return posts
}

func parseBlogPost(path string) BlogPost {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading blog post %s: %v", path, err)
	}

	content := string(data)

	// Parse frontmatter
	var fm blogFrontmatter
	if strings.HasPrefix(content, "---") {
		end := strings.Index(content[3:], "---")
		if end != -1 {
			fmData := content[3 : end+3]
			if err := yaml.Unmarshal([]byte(fmData), &fm); err != nil {
				log.Printf("Warning: failed to parse frontmatter in %s: %v", path, err)
			}
			content = strings.TrimSpace(content[end+6:])
		}
	}

	// Convert markdown to HTML
	md := []byte(content)
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	htmlContent := markdown.Render(doc, renderer)

	// Generate excerpt (first 150 chars of plain text)
	plain := strings.TrimSpace(content)
	plain = strings.ReplaceAll(plain, "#", "")
	plain = strings.ReplaceAll(plain, "*", "")
	plain = strings.ReplaceAll(plain, "`", "")
	excerpt := plain
	if len(excerpt) > 150 {
		excerpt = excerpt[:150] + "..."
	}
	excerpt = strings.ReplaceAll(excerpt, "\n", " ")

	return BlogPost{
		Title:   fm.Title,
		Date:    fm.Date,
		Slug:    fm.Slug,
		Excerpt: excerpt,
		Content: template.HTML(htmlContent),
		URL:     fmt.Sprintf("/blog/%s/", fm.Slug),
	}
}

func contentPath(filename string) string {
	primary := filepath.Join("content", filename)
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	return filepath.Join("content-example", filename)
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