package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	PubDate     string   `xml:"pubDate"`
	Creator     string   `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Description string   `xml:"description"`
	Content     string   `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	GUID        string   `xml:"guid"`
	Categories  []string `xml:"category"`
}

func main() {
	feedURL := flag.String("feed", "", "RSS feed URL")
	sinceDays := flag.Int("since", 0, "Number of days to look back (0 = no limit)")
	enableAuthors := flag.Bool("authors", false, "Enable author filtering using ALLOWED_AUTHOR_LIST environment variable")
	format := flag.String("format", "rss", "Output format: 'rss' or 'markdown'")
	saveToDir := flag.String("save-to", "", "Directory to save individual article files")
	buildFromDir := flag.String("build-from", "", "Directory to read article files from and build combined feed")
	maxItems := flag.Int("max-items", 1000, "Maximum number of items in output feed")
	flag.Parse()

	if *format != "rss" && *format != "markdown" {
		fmt.Fprintf(os.Stderr, "Error: --format must be 'rss' or 'markdown'\n")
		flag.Usage()
		os.Exit(1)
	}

	// Mode: build combined feed from article files
	if *buildFromDir != "" {
		items, err := loadArticlesFromDir(*buildFromDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading articles: %v\n", err)
			os.Exit(1)
		}
		sortItemsByDate(items)
		if len(items) > *maxItems {
			items = items[:*maxItems]
		}
		fmt.Fprintf(os.Stderr, "Built feed from %d articles\n", len(items))
		if *format == "markdown" {
			outputMarkdown(items)
		} else {
			outputRSS(items)
		}
		return
	}

	// Mode: fetch, filter, and optionally save articles
	if *feedURL == "" {
		fmt.Fprintf(os.Stderr, "Error: --feed or --build-from is required\n")
		flag.Usage()
		os.Exit(1)
	}

	var allowedAuthors map[string]bool
	if *enableAuthors {
		allowedAuthorList := os.Getenv("ALLOWED_AUTHOR_LIST")
		if allowedAuthorList == "" {
			fmt.Fprintf(os.Stderr, "Error: --authors flag requires ALLOWED_AUTHOR_LIST environment variable to be set\n")
			os.Exit(1)
		}
		allowedAuthors = loadAllowedAuthorsFromEnv(allowedAuthorList)
	}

	resp, err := http.Get(*feedURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching feed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: received status code %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing RSS: %v\n", err)
		os.Exit(1)
	}

	var cutoffDate time.Time
	if *sinceDays > 0 {
		cutoffDate = time.Now().AddDate(0, 0, -*sinceDays)
	}

	var filteredItems []Item
	for _, item := range rss.Channel.Items {
		if allowedAuthors != nil && !allowedAuthors[item.Creator] {
			continue
		}
		if *sinceDays > 0 {
			pubDate, err := parseRSSDate(item.PubDate)
			if err != nil {
				continue
			}
			if pubDate.Before(cutoffDate) {
				continue
			}
		}
		filteredItems = append(filteredItems, item)
	}

	if *saveToDir != "" {
		saved, err := saveArticlesToDir(filteredItems, *saveToDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving articles: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Saved %d new articles to %s\n", saved, *saveToDir)
		return
	}

	if *format == "markdown" {
		outputMarkdown(filteredItems)
	} else {
		outputRSS(filteredItems)
	}
}

func articleFilename(item Item) string {
	key := item.GUID
	if key == "" {
		key = item.Link
	}
	hash := sha256.Sum256([]byte(key))

	datePrefix := "0000-00-00"
	if t, err := parseRSSDate(item.PubDate); err == nil {
		datePrefix = t.Format("2006-01-02")
	}

	return fmt.Sprintf("%s-%x.xml", datePrefix, hash[:4])
}

func saveArticlesToDir(items []Item, dir string) (int, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, err
	}

	saved := 0
	for _, item := range items {
		filename := filepath.Join(dir, articleFilename(item))
		if _, err := os.Stat(filename); err == nil {
			continue // already exists, skip
		}
		content := renderItemXML(item)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return saved, err
		}
		saved++
	}
	return saved, nil
}

func loadArticlesFromDir(dir string) ([]Item, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".xml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}
		var rss RSS
		if err := xml.Unmarshal(data, &rss); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}
		items = append(items, rss.Channel.Items...)
	}
	return items, nil
}

func renderItemXML(item Item) string {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	sb.WriteString("<rss version=\"2.0\" xmlns:dc=\"http://purl.org/dc/elements/1.1/\" xmlns:content=\"http://purl.org/rss/1.0/modules/content/\">\n")
	sb.WriteString("  <channel>\n")
	sb.WriteString("    <item>\n")
	sb.WriteString(fmt.Sprintf("      <title>%s</title>\n", escapeXML(item.Title)))
	sb.WriteString(fmt.Sprintf("      <link>%s</link>\n", escapeXML(item.Link)))
	if item.GUID != "" {
		sb.WriteString(fmt.Sprintf("      <guid>%s</guid>\n", escapeXML(item.GUID)))
	}
	if item.PubDate != "" {
		sb.WriteString(fmt.Sprintf("      <pubDate>%s</pubDate>\n", escapeXML(item.PubDate)))
	}
	if item.Creator != "" {
		sb.WriteString(fmt.Sprintf("      <dc:creator>%s</dc:creator>\n", escapeXML(item.Creator)))
	}
	if item.Description != "" {
		sb.WriteString(fmt.Sprintf("      <description>%s</description>\n", escapeXML(item.Description)))
	}
	if item.Content != "" {
		sb.WriteString(fmt.Sprintf("      <content:encoded><![CDATA[%s]]></content:encoded>\n", item.Content))
	}
	for _, category := range item.Categories {
		if category != "" {
			sb.WriteString(fmt.Sprintf("      <category>%s</category>\n", escapeXML(category)))
		}
	}
	sb.WriteString("    </item>\n")
	sb.WriteString("  </channel>\n")
	sb.WriteString("</rss>\n")
	return sb.String()
}

func parseRSSDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	}
	dateStr = strings.TrimSpace(dateStr)
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func loadAllowedAuthorsFromEnv(authorList string) map[string]bool {
	authors := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(authorList))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" && !strings.HasPrefix(name, "#") {
			authors[name] = true
		}
	}
	return authors
}

func outputMarkdown(items []Item) {
	for _, item := range items {
		author := item.Creator
		if author == "" {
			author = "Unknown"
		}
		fmt.Printf("- [%s](%s) - %s\n", item.Title, item.Link, author)
	}
}

func outputRSS(items []Item) {
	fmt.Println(`<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Println(`<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/">`)
	fmt.Println(`  <channel>`)
	fmt.Println(`    <title>Filtered Technical Blog Posts</title>`)
	fmt.Println(`    <link>https://xebia.com/blog/</link>`)
	fmt.Println(`    <description>Filtered feed of technical blog posts</description>`)
	fmt.Printf("    <lastBuildDate>%s</lastBuildDate>\n", time.Now().Format(time.RFC1123Z))

	for _, item := range items {
		fmt.Println(`    <item>`)
		fmt.Printf("      <title>%s</title>\n", escapeXML(item.Title))
		fmt.Printf("      <link>%s</link>\n", escapeXML(item.Link))
		if item.GUID != "" {
			fmt.Printf("      <guid>%s</guid>\n", escapeXML(item.GUID))
		}
		if item.PubDate != "" {
			fmt.Printf("      <pubDate>%s</pubDate>\n", escapeXML(item.PubDate))
		}
		if item.Creator != "" {
			fmt.Printf("      <dc:creator>%s</dc:creator>\n", escapeXML(item.Creator))
		}
		if item.Description != "" {
			fmt.Printf("      <description>%s</description>\n", escapeXML(item.Description))
		}
		if item.Content != "" {
			fmt.Printf("      <content:encoded><![CDATA[%s]]></content:encoded>\n", item.Content)
		}
		for _, category := range item.Categories {
			if category != "" {
				fmt.Printf("      <category>%s</category>\n", escapeXML(category))
			}
		}
		fmt.Println(`    </item>`)
	}

	fmt.Println(`  </channel>`)
	fmt.Println(`</rss>`)
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func sortItemsByDate(items []Item) {
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			date1, err1 := parseRSSDate(items[i].PubDate)
			date2, err2 := parseRSSDate(items[j].PubDate)
			if err1 == nil && err2 == nil && date2.After(date1) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
