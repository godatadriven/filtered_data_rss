# Xebia Blog Feed Metadata Analysis

## Overview
Analysis of https://xebia.com/blog/category/domains/data-ai/feed to identify metadata patterns distinguishing marketing posts from technical consultant posts.

## Key Metadata Differentiators

### 1. **Post Type (in permalink)**

**Marketing Posts:**
- Use `post_type=news` or `post_type=articles` in permalinks
- Examples:
  - `?post_type=news&p=128195` (Xebia Recognized as a Disruptor)
  - `?post_type=news&p=127926` (2026: The Year Software Engineering...)
  - `?post_type=articles&p=127246` (Momentum in Motion)

**Technical Consultant Posts:**
- Use simple `?p=` parameter (no post_type)
- Examples:
  - `?p=130305` (Realist's Guide to Hybrid Mesh Architecture)
  - `?p=127255` (How to Reap the Benefits of LLM-Powered Coding Assistants)
  - `?p=127402` (AI in Healthcare)

**Note:** This pattern has exceptions - some technical articles also use `post_type=articles`

### 2. **Author Metadata**

**Marketing Posts:**
- Include job titles: "Kiran Madhunapantula, COO"
- Use corporate email addresses: "klaudia.wachnio@xebia.com"
- Use username/handles: "mariaskorupa"

**Technical Consultant Posts:**
- Plain full names: "Giovanni Lanzani", "XiaoHan Li", "Katarzyna Kusznierczuk", "Nafiseh Nazemi"
- No job titles or corporate identifiers

**Note:** "klaudia.wachnio@xebia.com" appears in both marketing and technical posts, suggesting some authors write both types

### 3. **Title Patterns**

**Marketing Posts:**
- Promotional/recognition-focused:
  - "Xebia Recognized as a Disruptor..."
  - "Momentum in Motion: AI Innovation, Market Impact, and the Power of Partnership"
- Broad predictions/vision statements:
  - "2026: The Year Software Engineering Will Become AI Native"
  - "AI as a Force Multiplier: How Enterprises and ISVs Will Do More With Less"

**Technical Consultant Posts:**
- How-to guides:
  - "How to Reap the Benefits of LLM-Powered Coding Assistants, While Avoiding Their Pitfalls"
  - "Realist's Guide to Hybrid Mesh Architecture"
- Specific implementation topics:
  - "Reimagining Catastrophic Risk Underwriting with Agentic AI"
  - "From Customer Voice to Actionable Intelligence"
- Industry-specific deep dives:
  - "AI in Healthcare: Trends, Strategies, and Barriers Shaping Healthcare in 2026 and Beyond"

### 4. **Category/Tag Patterns**

**Marketing Posts:**
- "Partnership Recognition"
- "Awards"
- "GenAI Services"
- Generic "Enterprise AI"

**Technical Consultant Posts:**
- Specific technical areas: "Data Analytics", "Agentic AI", "Healthcare AI"
- Implementation-focused: "AI Strategy", "Risk Management"
- Domain-specific: "Software Engineering" (when technical)

### 5. **Content Descriptions/Subtitles**

**Marketing Posts:**
- Focus on company achievements and partnerships
- Vision and prediction-oriented
- "Companies treating AI as a feature will fall behind..."

**Technical Consultant Posts:**
- Focus on practical implementation
- "Hybrid data mesh architecture practical implementation guide"
- "Why AI ideas succeed or stall: a product-led look at AI feasibility and data readiness"
- "From Static Risk Snapshots to Continuous, Context-Aware Decisions on Databricks"

## Classification of Articles

### Likely Marketing Posts (6):
1. **Momentum in Motion** (Keith Landis) - Partnership focus, exec-level
2. **Xebia Recognized as a Disruptor** (klaudia.wachnio@xebia.com) - Award announcement, post_type=news
3. **2026: The Year Software Engineering...** (mariaskorupa) - Prediction/vision, post_type=news
4. **AI as a Force Multiplier** (Kiran Madhunapantula, COO) - Exec byline with title

### Likely Technical Consultant Posts (6):
1. **Realist's Guide to Hybrid Mesh Architecture** (XiaoHan Li) - Technical implementation guide
2. **How to Reap the Benefits of LLM-Powered Coding Assistants** (Giovanni Lanzani) - Practical how-to
3. **A Product-Led Approach** (Nafiseh Nazemi) - Technical strategy
4. **From Customer Voice to Actionable Intelligence** (klaudia.wachnio@xebia.com) - Technical implementation
5. **AI in Healthcare** (Katarzyna Kusznierczuk) - Domain deep dive
6. **Reimagining Catastrophic Risk Underwriting** (klaudia.wachnio@xebia.com) - Technical implementation on Databricks

## Most Reliable Differentiators (Ranked)

1. **Author format with job title** (e.g., "COO") → Almost certainly marketing
2. **post_type=news in permalink** → Strong indicator of marketing/PR
3. **Title contains "Recognized", "Award", "Partnership"** → Strong indicator of marketing
4. **Plain full name + technical how-to title** → Strong indicator of technical consultant
5. **Mention of specific technical platforms/implementations** (e.g., "on Databricks") → Technical consultant
6. **Username or email format in author field** → Ambiguous (could be either)

## Observations

- Some authors (like klaudia.wachnio@xebia.com) write both marketing and technical content
- The `post_type` field in WordPress is a strong but not perfect indicator
- Title patterns are highly predictive when combined with other metadata
- Technical posts often mention specific tools, platforms, or methodologies in titles/descriptions
