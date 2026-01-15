# RSS Feed Filter

A Go command-line tool that filters RSS feeds to show only technical blog posts (excluding marketing/news content) and outputs them as a Markdown list.

## Features

- Fetches and parses RSS feeds
- Filters out marketing content (posts with `/news/` or `/articles/` in the URL path)
- Filters out posts by authors with email addresses or business titles
- Optional whitelist of allowed authors from a text file
- Optional date filtering to show only recent posts
- Outputs clean Markdown list with linked titles and authors

## Installation

```bash
go build -o feed-filter
```

## Usage

### Basic usage (all technical posts):
```bash
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed"
```

### Filter by date (last N days):
```bash
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --since 7
```

### Filter by allowed authors:
```bash
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --authors allowed_authors.txt
```

### Combine all filters:
```bash
go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --since 90 --authors allowed_authors.txt
```

### Using the compiled binary:
```bash
./feed-filter --feed "https://xebia.com/blog/category/domains/data-ai/feed" --since 30
```

## Parameters

- `--feed` (required): RSS feed URL to fetch and filter
- `--since` (optional): Number of days to look back (0 = no limit, default: 0)
- `--authors` (optional): Path to text file with allowed author names (one per line)

## Output Format

The tool outputs a Markdown list to stdout:

```markdown
- [Post Title](https://example.com/post-url) - Author Name
- [Another Post](https://example.com/another-post) - Another Author
```

## Filtering Logic

The tool filters OUT posts that:
- Contain `/news/` in the URL path
- Contain `/articles/` in the URL path
- Have `post_type=news` in query parameters
- Have `post_type=article` or `post_type=articles` in query parameters
- Have authors with email addresses (containing `@`)
- Have authors with business titles (containing `,`)
- Have authors not in the allowed authors file (if `--authors` is specified)

This keeps only technical blog posts from real consultant names.

## Authors File Format

The authors file should contain one author name per line, exactly as it appears in the RSS feed:

```
# This is a comment - lines starting with # are ignored
Giovanni Lanzani
XiaoHan Li
Katarzyna Kusznierczuk
```

Empty lines are also ignored. Names are case-sensitive and must match exactly.

## Example

```bash
$ go run main.go --feed "https://xebia.com/blog/category/domains/data-ai/feed" --since 7
- [Realist's Guide to Hybrid Mesh Architecture (1): Single Source of Truth vs Democratisation](https://xebia.com/blog/realists-guide-to-hybrid-mesh-architecture-1-single-source-of-truth-vs-democratisation/) - XiaoHan Li
```
