# RSS Feed Filter

A Go command-line tool that filters RSS feeds and outputs them in RSS or Markdown format.

## Features

- Fetches and parses RSS feeds
- Optional whitelist of allowed authors from environment variable
- Optional date filtering to show only recent posts
- Outputs in RSS (default) or Markdown format
- GitHub Actions integration for automated RSS feed generation

## Installation

```bash
go build -o feed-filter
```

## Usage

### Basic usage (RSS output):
```bash
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed"
```

### Output as Markdown:
```bash
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --format markdown
```

### Filter by date (last N days):
```bash
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --since 7
```

### Filter by allowed authors:
```bash
export ALLOWED_AUTHOR_LIST="Giovanni Lanzani
XiaoHan Li
Katarzyna Kusznierczuk"
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --authors
```

### Combine all filters:
```bash
export ALLOWED_AUTHOR_LIST="Giovanni Lanzani
XiaoHan Li"
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --since 90 --authors --format rss
```

### Using the compiled binary:
```bash
./feed-filter --feed "https://xebia.com/blog/category/domains/data-ai/feed" --since 30 --format markdown
```

## Parameters

- `--feed` (required): RSS feed URL to fetch and filter
- `--since` (optional): Number of days to look back (0 = no limit, default: 0)
- `--authors` (optional): Enable author filtering using the `ALLOWED_AUTHOR_LIST` environment variable
- `--format` (optional): Output format: `rss` or `markdown` (default: `rss`)
- `--merge-existing` (optional): URL to existing RSS feed to merge with (useful for accumulating entries over time)
- `--max-items` (optional): Maximum number of items to keep in merged feed (default: 1000)

## Environment Variables

- `ALLOWED_AUTHOR_LIST`: Newline-separated list of allowed author names. Required when using `--authors` flag.

## Output Formats

### RSS Format (default)
Outputs a valid RSS 2.0 XML feed:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
  <channel>
    <title>Filtered Technical Blog Posts</title>
    ...
  </channel>
</rss>
```

### Markdown Format
Outputs a Markdown list to stdout:
```markdown
- [Post Title](https://example.com/post-url) - Author Name
- [Another Post](https://example.com/another-post) - Another Author
```

## Filtering Logic

The tool filters OUT posts that:
- Have authors not in the allowed authors list (if `--authors` flag is enabled)

## GitHub Actions Integration

The repository includes a GitHub Action workflow (`.github/workflows/generate-rss.yml`) that automatically generates a filtered RSS feed every 24 hours.

### Setup

1. Go to your repository Settings → Secrets and variables → Actions → Variables
2. Add the following repository variables:
   - `FEED_URL`: The RSS feed URL to filter (e.g., `https://xebia.com/blog/category/domains/data-ai/feed`)
   - `ALLOWED_AUTHOR_LIST`: Newline-separated list of allowed authors (e.g., `Giovanni Lanzani\nXiaoHan Li`)

The workflow will:
- Run automatically every 24 hours at midnight UTC
- Can be triggered manually from the Actions tab
- Fetch the existing filtered feed from releases (if it exists)
- Merge new filtered entries with existing ones, removing duplicates
- Keep up to 1000 most recent entries (configurable via `--max-items`)
- Upload the updated `filtered-feed.xml` to GitHub Releases (tagged as "latest")

This approach ensures that even though the original RSS feed only retains the last 10 entries, your filtered feed accumulates posts over time up to the configured maximum.

The RSS feed will be available at: `https://github.com/<your-username>/<your-repo>/releases/download/latest/filtered-feed.xml`

### Author List Format

The `ALLOWED_AUTHOR_LIST` environment variable should contain one author name per line, exactly as it appears in the RSS feed:

```
Giovanni Lanzani
XiaoHan Li
Katarzyna Kusznierczuk
```

Names are case-sensitive and must match exactly. Lines starting with `#` are treated as comments and ignored.

## Examples

### Markdown output
```bash
$ go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --since 7 --format markdown
- [Realist's Guide to Hybrid Mesh Architecture (1): Single Source of Truth vs Democratisation](https://xebia.com/blog/realists-guide-to-hybrid-mesh-architecture-1-single-source-of-truth-vs-democratisation/) - XiaoHan Li
- [How to Reap the Benefits of LLM-Powered Coding Assistants, While Avoiding Their Pitfalls](https://xebia.com/blog/how-reap-benefits-llm-powered-coding-assistants-avoiding-pitfalls/) - Giovanni Lanzani
```

### RSS output with author filtering
```bash
$ export ALLOWED_AUTHOR_LIST="Giovanni Lanzani
XiaoHan Li"
$ go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --authors --since 30 > filtered-feed.xml
```

### Merging with existing feed (accumulate entries over time)
```bash
$ go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" \
  --merge-existing "https://example.com/existing-feed.xml" \
  --max-items 1000 > updated-feed.xml
```
This fetches the existing feed, merges it with newly filtered entries, removes duplicates (by GUID or link), sorts by date (newest first), and limits to 1000 items.
