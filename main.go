package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
	feedURL := flag.String("feed", "", "RSS feed URL (required)")
	sinceDays := flag.Int("since", 0, "Number of days to look back (0 = no limit)")
	enableAuthors := flag.Bool("authors", false, "Enable author filtering using ALLOWED_AUTHOR_LIST environment variable")
	format := flag.String("format", "rss", "Output format: 'rss' or 'markdown'")
	mergeExisting := flag.String("merge-existing", "", "URL to existing RSS feed to merge with (optional)")
	maxItems := flag.Int("max-items", 1000, "Maximum number of items to keep in merged feed")
	flag.Parse()

	if *feedURL == "" {
		fmt.Fprintf(os.Stderr, "Error: --feed parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate format
	if *format != "rss" && *format != "markdown" {
		fmt.Fprintf(os.Stderr, "Error: --format must be 'rss' or 'markdown'\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load allowed authors if --authors flag is present
	var allowedAuthors map[string]bool
	if *enableAuthors {
		allowedAuthorList := os.Getenv("ALLOWED_AUTHOR_LIST")
		if allowedAuthorList == "" {
			fmt.Fprintf(os.Stderr, "Error: --authors flag requires ALLOWED_AUTHOR_LIST environment variable to be set\n")
			os.Exit(1)
		}
		allowedAuthors = loadAllowedAuthorsFromEnv(allowedAuthorList)
	}

	// Fetch the feed
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

	// Parse the RSS feed
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

	// Calculate cutoff date if since is specified
	var cutoffDate time.Time
	if *sinceDays > 0 {
		cutoffDate = time.Now().AddDate(0, 0, -*sinceDays)
	}

	// Filter items
	var filteredItems []Item
	for _, item := range rss.Channel.Items {
		// Check if URL contains post_type=news or post_type=article
		if shouldFilter(item.Link) {
			continue
		}

		// Check if author should be filtered (email or has business title)
		if shouldFilterAuthor(item.Creator) {
			continue
		}

		// Check if author is in allowed list (if authors filtering is enabled)
		if allowedAuthors != nil && !allowedAuthors[item.Creator] {
			continue
		}

		// Check date if since parameter is specified
		if *sinceDays > 0 {
			pubDate, err := parseRSSDate(item.PubDate)
			if err != nil {
				// Skip items with unparseable dates
				continue
			}
			if pubDate.Before(cutoffDate) {
				continue
			}
		}

		filteredItems = append(filteredItems, item)
	}

	// Merge with existing feed if specified
	if *mergeExisting != "" {
		existingItems, err := fetchExistingFeed(*mergeExisting)
		if err != nil {
			// Log warning but continue with just the new items
			fmt.Fprintf(os.Stderr, "Warning: could not fetch existing feed: %v\n", err)
		} else {
			filteredItems = mergeAndDeduplicateItems(filteredItems, existingItems, *maxItems)
		}
	}

	// Output in the requested format
	if *format == "markdown" {
		outputMarkdown(filteredItems)
	} else {
		outputRSS(filteredItems, *feedURL)
	}
}

// shouldFilter returns true if the URL contains /news/ or /articles/ in the path
// or has post_type=news or post_type=article in query parameters
func shouldFilter(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	// Check URL path for /news/ or /articles/
	path := parsedURL.Path
	if strings.Contains(path, "/news/") || strings.Contains(path, "/articles/") {
		return true
	}

	// Also check query parameters as fallback
	query := parsedURL.Query()
	postType := query.Get("post_type")

	return postType == "news" || postType == "article" || postType == "articles"
}

// shouldFilterAuthor returns true if the author should be filtered out
// Filters: email addresses (contains @) and business titles (contains comma)
func shouldFilterAuthor(author string) bool {
	if author == "" {
		return false
	}

	// Filter out email addresses
	if strings.Contains(author, "@") {
		return true
	}

	// Filter out authors with business titles (indicated by comma)
	if strings.Contains(author, ",") {
		return true
	}

	return false
}

// parseRSSDate parses common RSS date formats
func parseRSSDate(dateStr string) (time.Time, error) {
	// RSS typically uses RFC1123Z format: "Mon, 02 Jan 2006 15:04:05 -0700"
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"2006-01-02T15:04:05Z07:00", // ISO 8601
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

// loadAllowedAuthorsFromEnv parses the ALLOWED_AUTHOR_LIST environment variable
// which contains author names separated by newlines
func loadAllowedAuthorsFromEnv(authorList string) map[string]bool {
	authors := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(authorList))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" && !strings.HasPrefix(name, "#") { // Skip empty lines and comments
			authors[name] = true
		}
	}
	return authors
}

// outputMarkdown prints filtered items as a Markdown list
func outputMarkdown(items []Item) {
	for _, item := range items {
		author := item.Creator
		if author == "" {
			author = "Unknown"
		}
		fmt.Printf("- [%s](%s) - %s\n", item.Title, item.Link, author)
	}
}

// outputRSS generates and prints an RSS feed with the filtered items
func outputRSS(items []Item, originalFeedURL string) {
	fmt.Println(`<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Println(`<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/">`)
	fmt.Println(`  <channel>`)
	fmt.Println(`    <title>Filtered Technical Blog Posts</title>`)
	fmt.Printf("    <link>%s</link>\n", escapeXML(originalFeedURL))
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

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// fetchExistingFeed downloads and parses an existing RSS feed from a URL
func fetchExistingFeed(feedURL string) ([]Item, error) {
	resp, err := http.Get(feedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// If the feed doesn't exist yet (404), return empty list
	if resp.StatusCode == http.StatusNotFound {
		return []Item{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, err
	}

	return rss.Channel.Items, nil
}

// mergeAndDeduplicateItems combines new and existing items, removes duplicates,
// sorts by date (newest first), and limits to maxItems
func mergeAndDeduplicateItems(newItems, existingItems []Item, maxItems int) []Item {
	// Use a map to deduplicate by GUID (or Link if GUID is empty)
	itemMap := make(map[string]Item)

	// Add existing items first
	for _, item := range existingItems {
		key := item.GUID
		if key == "" {
			key = item.Link
		}
		itemMap[key] = item
	}

	// Add new items (will overwrite existing if GUID/Link matches)
	for _, item := range newItems {
		key := item.GUID
		if key == "" {
			key = item.Link
		}
		itemMap[key] = item
	}

	// Convert map back to slice
	merged := make([]Item, 0, len(itemMap))
	for _, item := range itemMap {
		merged = append(merged, item)
	}

	// Sort by publication date (newest first)
	sortItemsByDate(merged)

	// Limit to maxItems
	if len(merged) > maxItems {
		merged = merged[:maxItems]
	}

	return merged
}

// sortItemsByDate sorts items by publication date in descending order (newest first)
func sortItemsByDate(items []Item) {
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			date1, err1 := parseRSSDate(items[i].PubDate)
			date2, err2 := parseRSSDate(items[j].PubDate)

			// If both dates are valid and date2 is newer, swap
			if err1 == nil && err2 == nil && date2.After(date1) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
