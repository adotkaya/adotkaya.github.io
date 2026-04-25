package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"
)

type PageData struct {
	Config     Config
	Pages      Pages
	Resume     Resume
	Now        Now
	Uses       Uses
	Experience []Experience
	Projects   []Project
	Books      []Book
	Tools      []Tool
	Posts      []BlogPost
	Post       *BlogPost
	PrevPost   *BlogPost
	NextPost   *BlogPost
	Year       int
	BaseURL    string
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

type Pages struct {
	Subtitles map[string]string `yaml:"subtitles"`
	Links     map[string]string `yaml:"links"`
}

type Resume struct {
	Summary   string       `yaml:"summary"`
	Skills    []SkillGroup `yaml:"skills"`
	Education []Education  `yaml:"education"`
}

type SkillGroup struct {
	Category string `yaml:"category"`
	Items    string `yaml:"items"`
}

type Education struct {
	Institution string   `yaml:"institution"`
	Degree      string   `yaml:"degree"`
	Period      string   `yaml:"period"`
	Notes       []string `yaml:"notes"`
}

type Now struct {
	LastUpdated  string    `yaml:"last_updated"`
	Items        []NowItem `yaml:"items"`
	Note         string    `yaml:"note"`
	NoteLink     string    `yaml:"note_link"`
	NoteLinkText string    `yaml:"note_link_text"`
}

type NowItem struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

type Uses struct {
	Sections []UsesSection `yaml:"sections"`
}

type UsesSection struct {
	Title string     `yaml:"title"`
	Items []UsesItem `yaml:"items"`
}

type UsesItem struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
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
	Status string `yaml:"status"`
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
	Title         string
	Date          string
	DateFormatted string
	Slug          string
	Excerpt       string
	Content       template.HTML
	URL           string
	ReadingTime   int
}

type blogFrontmatter struct {
	Title string `yaml:"title"`
	Date  string `yaml:"date"`
	Slug  string `yaml:"slug"`
	Draft bool   `yaml:"draft"`
}

type sitemapURL struct {
	Loc        string
	LastMod    string
	ChangeFreq string
	Priority   string
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
		if i > 0 {
			postData.PrevPost = &data.Posts[i-1]
		}
		if i < len(data.Posts)-1 {
			postData.NextPost = &data.Posts[i+1]
		}
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

	// Books
	os.MkdirAll("public/books", 0755)
	renderTemplate("templates/books.html", "public/books/index.html", data)

	// 404
	renderTemplate("templates/404.html", "public/404.html", data)

	// Sitemap
	generateSitemap(data)

	// RSS Feed
	generateFeed(data)

	// Robots.txt
	generateRobots()

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
	data.BaseURL = "https://" + data.Config.Username + ".github.io"
	data.Pages = mustParseYAML[Pages](contentPath("pages.yaml"))
	data.Resume = mustParseYAML[Resume](contentPath("resume.yaml"))
	data.Now = mustParseYAML[Now](contentPath("now.yaml"))
	data.Uses = mustParseYAML[Uses](contentPath("uses.yaml"))
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
		if post.Date == "" {
			continue // skip drafts or unparseable
		}
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

	if fm.Draft {
		return BlogPost{Date: ""}
	}

	// Calculate reading time (average 200 words per minute)
	wordCount := len(strings.Fields(content))
	readingTime := wordCount / 200
	if readingTime < 1 {
		readingTime = 1
	}

	// Convert markdown to HTML
	md := []byte(content)
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := mdhtml.CommonFlags | mdhtml.HrefTargetBlank
	opts := mdhtml.RendererOptions{Flags: htmlFlags}
	renderer := mdhtml.NewRenderer(opts)
	htmlContent := markdown.Render(doc, renderer)

	// Apply syntax highlighting to code blocks
	highlighted := highlightCodeBlocks(string(htmlContent))

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
		Title:         fm.Title,
		Date:          fm.Date,
		DateFormatted: formatDate(fm.Date),
		Slug:          fm.Slug,
		Excerpt:       excerpt,
		Content:       template.HTML(highlighted),
		URL:           fmt.Sprintf("/blog/%s/", fm.Slug),
		ReadingTime:   readingTime,
	}
}

func formatDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("January 2, 2006")
}

func highlightCodeBlocks(htmlContent string) string {
	// Regex to find <pre><code class="language-xxx">...</code></pre>
	re := regexp.MustCompile(`<pre><code(?: class="language-([^"]*)")?>([^<]*)</code></pre>`)

	return re.ReplaceAllStringFunc(htmlContent, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		lang := submatches[1]
		code := submatches[2]

		// Unescape HTML entities
		code = strings.ReplaceAll(code, "&lt;", "<")
		code = strings.ReplaceAll(code, "&gt;", ">")
		code = strings.ReplaceAll(code, "&amp;", "&")
		code = strings.ReplaceAll(code, "&quot;", "\"")

		// Get lexer
		lexer := lexers.Get(lang)
		if lexer == nil {
			lexer = lexers.Fallback
		}

		// Get style - using a dark theme that fits the site
		style := styles.Get("monokai")
		if style == nil {
			style = styles.Fallback
		}

		// Create formatter with inline styles (no external CSS needed)
		formatter := html.New(html.WithClasses(false), html.Standalone(false), html.TabWidth(4))

		// Tokenize and format
		iterator, err := lexer.Tokenise(nil, code)
		if err != nil {
			return match
		}

		var buf bytes.Buffer
		err = formatter.Format(&buf, style, iterator)
		if err != nil {
			return match
		}

		return fmt.Sprintf(`<div class="highlight"><pre class="chroma">%s</pre></div>`, buf.String())
	})
}

func generateFeed(data PageData) {
	baseURL := data.BaseURL
	now := time.Now().Format(time.RFC3339)

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<feed xmlns="http://www.w3.org/2005/Atom">` + "\n")
	buf.WriteString(fmt.Sprintf("  <title>%s</title>\n", data.Config.Name+"'s Blog"))
	buf.WriteString(fmt.Sprintf("  <link href=\"%s/blog/\"/>\n", baseURL))
	buf.WriteString(fmt.Sprintf("  <link rel=\"self\" href=\"%s/feed.xml\"/>\n", baseURL))
	buf.WriteString(fmt.Sprintf("  <updated>%s</updated>\n", now))
	buf.WriteString(fmt.Sprintf("  <id>%s/blog/</id>\n", baseURL))
	buf.WriteString(fmt.Sprintf("  <author><name>%s</name></author>\n", data.Config.Name))

	for _, post := range data.Posts {
		postDate, _ := time.Parse("2006-01-02", post.Date)
		buf.WriteString("  <entry>\n")
		buf.WriteString(fmt.Sprintf("    <title>%s</title>\n", template.HTMLEscapeString(post.Title)))
		buf.WriteString(fmt.Sprintf("    <link href=\"%s%s\"/>\n", baseURL, post.URL))
		buf.WriteString(fmt.Sprintf("    <id>%s%s</id>\n", baseURL, post.URL))
		buf.WriteString(fmt.Sprintf("    <updated>%s</updated>\n", postDate.Format(time.RFC3339)))
		buf.WriteString(fmt.Sprintf("    <summary>%s</summary>\n", template.HTMLEscapeString(post.Excerpt)))
		buf.WriteString("  </entry>\n")
	}

	buf.WriteString("</feed>\n")

	if err := os.WriteFile("public/feed.xml", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Error writing feed: %v", err)
	}
}

func generateSitemap(data PageData) {
	baseURL := "https://" + data.Config.Username + ".github.io"
	now := time.Now().Format("2006-01-02")

	urls := []sitemapURL{
		{baseURL + "/", now, "weekly", "1.0"},
		{baseURL + "/resume/", now, "monthly", "0.8"},
		{baseURL + "/projects/", now, "monthly", "0.8"},
		{baseURL + "/blog/", now, "weekly", "0.9"},
		{baseURL + "/now/", now, "weekly", "0.6"},
		{baseURL + "/uses/", now, "monthly", "0.6"},
		{baseURL + "/books/", now, "monthly", "0.6"},
	}

	for _, post := range data.Posts {
		urls = append(urls, sitemapURL{
			Loc:        baseURL + post.URL,
			LastMod:    post.Date,
			ChangeFreq: "never",
			Priority:   "0.7",
		})
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	for _, u := range urls {
		buf.WriteString("  <url>\n")
		buf.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", u.Loc))
		buf.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", u.LastMod))
		buf.WriteString(fmt.Sprintf("    <changefreq>%s</changefreq>\n", u.ChangeFreq))
		buf.WriteString(fmt.Sprintf("    <priority>%s</priority>\n", u.Priority))
		buf.WriteString("  </url>\n")
	}

	buf.WriteString("</urlset>\n")

	if err := os.WriteFile("public/sitemap.xml", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Error writing sitemap: %v", err)
	}
}

func generateRobots() {
	content := `User-agent: *
Allow: /
Sitemap: https://adotkaya.github.io/sitemap.xml
`
	if err := os.WriteFile("public/robots.txt", []byte(content), 0644); err != nil {
		log.Fatalf("Error writing robots.txt: %v", err)
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