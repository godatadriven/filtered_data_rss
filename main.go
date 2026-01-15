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
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	Creator string `xml:"http://purl.org/dc/elements/1.1/ creator"`
}

func main() {
	feedURL := flag.String("feed", "", "RSS feed URL (required)")
	sinceDays := flag.Int("since", 0, "Number of days to look back (0 = no limit)")
	authorsFile := flag.String("authors", "", "Path to file with allowed author names (one per line)")
	flag.Parse()

	if *feedURL == "" {
		fmt.Fprintf(os.Stderr, "Error: --feed parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load allowed authors if file is specified
	var allowedAuthors map[string]bool
	if *authorsFile != "" {
		var err error
		allowedAuthors, err = loadAllowedAuthors(*authorsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading authors file: %v\n", err)
			os.Exit(1)
		}
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

	// Filter and output items
	for _, item := range rss.Channel.Items {
		// Check if URL contains post_type=news or post_type=article
		if shouldFilter(item.Link) {
			continue
		}

		// Check if author should be filtered (email or has business title)
		if shouldFilterAuthor(item.Creator) {
			continue
		}

		// Check if author is in allowed list (if authors file is specified)
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

		// Output as Markdown list item
		author := item.Creator
		if author == "" {
			author = "Unknown"
		}
		fmt.Printf("- [%s](%s) - %s\n", item.Title, item.Link, author)
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

// loadAllowedAuthors reads a file with author names (one per line) and returns a set
func loadAllowedAuthors(filePath string) (map[string]bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	authors := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" && !strings.HasPrefix(name, "#") { // Skip empty lines and comments
			authors[name] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return authors, nil
}
