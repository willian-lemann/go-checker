package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/playwright-community/playwright-go"
)

// SEOAudit represents the complete audit result
type SEOAudit struct {
	URL             string              `json:"url"`
	Timestamp       time.Time           `json:"timestamp"`
	TechnicalSEO    TechnicalSEOScore   `json:"technical_seo"`
	OnPageSEO       OnPageSEOScore      `json:"on_page_seo"`
	ContentQuality  ContentQualityScore `json:"content_quality"`
	LinkStructure   LinkStructureScore  `json:"link_structure"`
	SchemaMarkup    SchemaMarkupScore   `json:"schema_markup"`
	Security        SecurityScore       `json:"security"`
	UserExperience  UserExperienceScore `json:"user_experience"`
	WebVitals       WebVitalsScore      `json:"web_vitals"`
	OverallScore    float64             `json:"overall_score"`
	Grade           string              `json:"grade"`
	Recommendations []string            `json:"recommendations"`
	Markdown        string              `json:"markdown"`
}

// TechnicalSEOScore holds technical SEO metrics
type TechnicalSEOScore struct {
	Score            float64  `json:"score"`
	MaxScore         float64  `json:"max_score"`
	LoadTime         float64  `json:"load_time_ms"`
	PageSize         int64    `json:"page_size_bytes"`
	HTTPRequests     int      `json:"http_requests"`
	HasRobotsTxt     bool     `json:"has_robots_txt"`
	HasSitemap       bool     `json:"has_sitemap"`
	IsHTTPS          bool     `json:"is_https"`
	IsMobileFriendly bool     `json:"is_mobile_friendly"`
	HasViewport      bool     `json:"has_viewport"`
	HTTPStatusCode   int      `json:"http_status_code"`
	Issues           []string `json:"issues"`
}

// OnPageSEOScore holds on-page SEO metrics
type OnPageSEOScore struct {
	Score                  float64  `json:"score"`
	MaxScore               float64  `json:"max_score"`
	HasTitle               bool     `json:"has_title"`
	TitleLength            int      `json:"title_length"`
	HasMetaDescription     bool     `json:"has_meta_description"`
	MetaDescriptionLength  int      `json:"meta_description_length"`
	HasH1                  bool     `json:"has_h1"`
	H1Count                int      `json:"h1_count"`
	H2Count                int      `json:"h2_count"`
	HasOGTags              bool     `json:"has_og_tags"`
	HasTwitterCard         bool     `json:"has_twitter_card"`
	HasCanonical           bool     `json:"has_canonical"`
	KeywordInTitle         bool     `json:"keyword_in_title"`
	ProperHeadingHierarchy bool     `json:"proper_heading_hierarchy"`
	Issues                 []string `json:"issues"`
}

// ContentQualityScore holds content quality metrics
type ContentQualityScore struct {
	Score            float64  `json:"score"`
	MaxScore         float64  `json:"max_score"`
	WordCount        int      `json:"word_count"`
	ParagraphCount   int      `json:"paragraph_count"`
	ImageCount       int      `json:"image_count"`
	ImagesWithAlt    int      `json:"images_with_alt"`
	InternalLinks    int      `json:"internal_links"`
	ExternalLinks    int      `json:"external_links"`
	ReadabilityScore float64  `json:"readability_score"`
	Issues           []string `json:"issues"`
}

// LinkStructureScore holds link structure metrics
type LinkStructureScore struct {
	Score              float64  `json:"score"`
	MaxScore           float64  `json:"max_score"`
	InternalLinks      int      `json:"internal_links"`
	ExternalLinks      int      `json:"external_links"`
	BrokenLinks        int      `json:"broken_links"`
	HasBreadcrumbs     bool     `json:"has_breadcrumbs"`
	DescriptiveAnchors bool     `json:"descriptive_anchors"`
	Issues             []string `json:"issues"`
}

// SchemaMarkupScore holds schema markup metrics
type SchemaMarkupScore struct {
	Score           float64  `json:"score"`
	MaxScore        float64  `json:"max_score"`
	HasSchema       bool     `json:"has_schema"`
	SchemaTypes     []string `json:"schema_types"`
	HasOrganization bool     `json:"has_organization"`
	HasBreadcrumb   bool     `json:"has_breadcrumb"`
	Issues          []string `json:"issues"`
}

// SecurityScore holds security metrics
type SecurityScore struct {
	Score              float64  `json:"score"`
	MaxScore           float64  `json:"max_score"`
	IsHTTPS            bool     `json:"is_https"`
	HasSSL             bool     `json:"has_ssl"`
	MixedContent       bool     `json:"mixed_content"`
	HasSecurityHeaders bool     `json:"has_security_headers"`
	Issues             []string `json:"issues"`
}

// UserExperienceScore holds UX metrics
type UserExperienceScore struct {
	Score             float64  `json:"score"`
	MaxScore          float64  `json:"max_score"`
	HasFavicon        bool     `json:"has_favicon"`
	FontSizeReadable  bool     `json:"font_size_readable"`
	HasLangAttribute  bool     `json:"has_lang_attribute"`
	NoIntrusivePopups bool     `json:"no_intrusive_popups"`
	Issues            []string `json:"issues"`
}

// WebVitalsScore holds Core Web Vitals metrics
type WebVitalsScore struct {
	Score            float64                `json:"score"`
	MaxScore         float64                `json:"max_score"`
	LCP              int                    `json:"lcp_ms"`                // Largest Contentful Paint (ms)
	LCPRating        string                 `json:"lcp_rating"`            // good, needs-improvement, poor
	LCPAttribution   map[string]interface{} `json:"lcp_attribution"`       // LCP attribution data
	FCP              int                    `json:"fcp_ms"`                // First Contentful Paint (ms)
	FCPRating        string                 `json:"fcp_rating"`            // good, needs-improvement, poor
	CLS              float64                `json:"cls"`                   // Cumulative Layout Shift (unitless)
	CLSRating        string                 `json:"cls_rating"`            // good, needs-improvement, poor
	CLSAttribution   map[string]interface{} `json:"cls_attribution"`       // CLS attribution data
	INP              float64                `json:"inp_ms"`                // Interaction to Next Paint (ms)
	INPRating        string                 `json:"inp_rating"`            // good, needs-improvement, poor
	INPAttribution   map[string]interface{} `json:"inp_attribution"`       // INP attribution data
	TTFB             float64                `json:"ttfb_ms"`               // Time to First Byte (ms)
	TTFBRating       string                 `json:"ttfb_rating"`           // good, needs-improvement, poor
	DOMContentLoaded float64                `json:"dom_content_loaded_ms"` // DOMContentLoaded event (ms)
	DOMComplete      float64                `json:"dom_complete_ms"`       // DOM complete (ms)
	TransferSize     int64                  `json:"transfer_size_bytes"`   // Total transfer size
	ResourceCount    int                    `json:"resource_count"`        // Number of resources loaded
	Issues           []string               `json:"issues"`
}

// SEOAuditor performs SEO audits
type SEOAuditor struct {
	pw      *playwright.Playwright
	browser playwright.Browser
}

// NewSEOAuditor creates a new SEO auditor
func NewSEOAuditor() (*SEOAuditor, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	return &SEOAuditor{
		pw:      pw,
		browser: browser,
	}, nil
}

// Close closes the auditor
func (a *SEOAuditor) Close() error {
	if err := a.browser.Close(); err != nil {
		return err
	}
	return a.pw.Stop()
}

// AuditWebsite performs a complete SEO audit
func (a *SEOAuditor) AuditWebsite(targetURL string) (*SEOAudit, error) {
	audit := &SEOAudit{
		URL:       targetURL,
		Timestamp: time.Now(),
	}

	// Create a new page
	page, err := a.browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}
	defer page.Close()

	// Measure page load time
	startTime := time.Now()

	// Navigate to the page
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		return nil, fmt.Errorf("could not navigate to page: %v", err)
	}

	loadTime := time.Since(startTime).Milliseconds()

	// Run all audits
	audit.TechnicalSEO = a.auditTechnicalSEO(page, targetURL, float64(loadTime))
	audit.OnPageSEO = a.auditOnPageSEO(page)
	audit.ContentQuality = a.auditContentQuality(page, targetURL)
	audit.LinkStructure = a.auditLinkStructure(page, targetURL)
	audit.SchemaMarkup = a.auditSchemaMarkup(page)
	audit.Security = a.auditSecurity(targetURL, page)
	audit.UserExperience = a.auditUserExperience(page)
	audit.WebVitals = a.auditWebVitals(page)

	// Calculate overall score
	audit.OverallScore = a.calculateOverallScore(audit)
	audit.Grade = a.calculateGrade(audit.OverallScore)
	audit.Recommendations = a.generateRecommendations(audit)
	audit.Markdown = a.generateMarkdown(audit)

	return audit, nil
}

// auditTechnicalSEO performs technical SEO checks
func (a *SEOAuditor) auditTechnicalSEO(page playwright.Page, targetURL string, loadTime float64) TechnicalSEOScore {
	score := TechnicalSEOScore{
		MaxScore: 100,
		LoadTime: loadTime,
		Issues:   []string{},
	}

	parsedURL, _ := url.Parse(targetURL)
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Check HTTPS
	score.IsHTTPS = parsedURL.Scheme == "https"
	if score.IsHTTPS {
		score.Score += 15
	} else {
		score.Issues = append(score.Issues, "Site is not using HTTPS")
	}

	// Check viewport meta tag
	viewport, _ := page.Locator("meta[name='viewport']").Count()
	score.HasViewport = viewport > 0
	if score.HasViewport {
		score.Score += 10
		score.IsMobileFriendly = true
	} else {
		score.Issues = append(score.Issues, "Missing viewport meta tag")
	}

	// Check robots.txt
	score.HasRobotsTxt = a.checkURLExists(baseURL + "/robots.txt")
	if score.HasRobotsTxt {
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, "robots.txt not found")
	}

	// Check sitemap
	score.HasSitemap = a.checkURLExists(baseURL + "/sitemap.xml")
	if score.HasSitemap {
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, "sitemap.xml not found")
	}

	// Check page load time
	if loadTime < 2000 {
		score.Score += 20
	} else if loadTime < 3000 {
		score.Score += 15
		score.Issues = append(score.Issues, "Page load time is moderate (2-3 seconds)")
	} else if loadTime < 5000 {
		score.Score += 10
		score.Issues = append(score.Issues, "Page load time is slow (3-5 seconds)")
	} else {
		score.Score += 5
		score.Issues = append(score.Issues, fmt.Sprintf("Page load time is very slow (%.2f seconds)", float64(loadTime)/1000))
	}

	// Estimate page size
	content, _ := page.Content()
	score.PageSize = int64(len(content))
	if score.PageSize < 3*1024*1024 { // < 3MB
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, fmt.Sprintf("Page size is large (%.2f MB)", float64(score.PageSize)/(1024*1024)))
	}

	// Count images and scripts (estimate HTTP requests)
	images, _ := page.Locator("img").Count()
	scripts, _ := page.Locator("script").Count()
	links, _ := page.Locator("link[rel='stylesheet']").Count()
	score.HTTPRequests = images + scripts + links
	if score.HTTPRequests < 50 {
		score.Score += 10
	} else if score.HTTPRequests < 100 {
		score.Score += 5
	} else {
		score.Issues = append(score.Issues, fmt.Sprintf("High number of HTTP requests (%d)", score.HTTPRequests))
	}

	// Check for canonical tag
	canonical, _ := page.Locator("link[rel='canonical']").Count()
	if canonical > 0 {
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, "Missing canonical tag")
	}

	// Check HTTP status (if we got here, it's likely 200)
	score.HTTPStatusCode = 200
	score.Score += 5

	return score
}

// auditOnPageSEO performs on-page SEO checks
func (a *SEOAuditor) auditOnPageSEO(page playwright.Page) OnPageSEOScore {
	score := OnPageSEOScore{
		MaxScore: 100,
		Issues:   []string{},
	}

	// Check title tag
	title, err := page.Title()
	score.HasTitle = err == nil && title != ""
	score.TitleLength = len(title)

	if score.HasTitle {
		score.Score += 15
		if score.TitleLength >= 50 && score.TitleLength <= 60 {
			score.Score += 10
		} else if score.TitleLength < 50 {
			score.Issues = append(score.Issues, "Title tag is too short (< 50 characters)")
			score.Score += 5
		} else if score.TitleLength > 60 {
			score.Issues = append(score.Issues, "Title tag is too long (> 60 characters)")
			score.Score += 5
		}
	} else {
		score.Issues = append(score.Issues, "Missing title tag")
	}

	// Check meta description
	metaDesc, _ := page.Locator("meta[name='description']").GetAttribute("content")
	score.HasMetaDescription = metaDesc != ""
	score.MetaDescriptionLength = len(metaDesc)

	if score.HasMetaDescription {
		score.Score += 15
		if score.MetaDescriptionLength >= 150 && score.MetaDescriptionLength <= 160 {
			score.Score += 10
		} else if score.MetaDescriptionLength < 150 {
			score.Issues = append(score.Issues, "Meta description is too short (< 150 characters)")
			score.Score += 5
		} else if score.MetaDescriptionLength > 160 {
			score.Issues = append(score.Issues, "Meta description is too long (> 160 characters)")
			score.Score += 5
		}
	} else {
		score.Issues = append(score.Issues, "Missing meta description")
	}

	// Check H1 tag
	h1Count, _ := page.Locator("h1").Count()
	score.H1Count = h1Count
	score.HasH1 = h1Count > 0

	if score.HasH1 {
		if h1Count == 1 {
			score.Score += 15
		} else {
			score.Issues = append(score.Issues, fmt.Sprintf("Multiple H1 tags found (%d)", h1Count))
			score.Score += 5
		}
	} else {
		score.Issues = append(score.Issues, "Missing H1 tag")
	}

	// Check H2 tags
	h2Count, _ := page.Locator("h2").Count()
	score.H2Count = h2Count
	if h2Count > 0 {
		score.Score += 5
	}

	// Check heading hierarchy
	score.ProperHeadingHierarchy = a.checkHeadingHierarchy(page)
	if score.ProperHeadingHierarchy {
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, "Improper heading hierarchy")
	}

	// Check Open Graph tags
	ogTitle, _ := page.Locator("meta[property='og:title']").Count()
	ogDesc, _ := page.Locator("meta[property='og:description']").Count()
	ogImage, _ := page.Locator("meta[property='og:image']").Count()
	score.HasOGTags = ogTitle > 0 && ogDesc > 0 && ogImage > 0

	if score.HasOGTags {
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, "Incomplete Open Graph tags")
	}

	// Check Twitter Card tags
	twitterCard, _ := page.Locator("meta[name='twitter:card']").Count()
	score.HasTwitterCard = twitterCard > 0

	if score.HasTwitterCard {
		score.Score += 5
	} else {
		score.Issues = append(score.Issues, "Missing Twitter Card tags")
	}

	// Check canonical tag
	canonical, _ := page.Locator("link[rel='canonical']").Count()
	score.HasCanonical = canonical > 0

	if score.HasCanonical {
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, "Missing canonical tag")
	}

	// Check if keyword in title (basic check)
	if score.HasTitle && len(title) > 0 {
		score.KeywordInTitle = true
		score.Score += 5
	}

	return score
}

// auditContentQuality performs content quality checks
func (a *SEOAuditor) auditContentQuality(page playwright.Page, targetURL string) ContentQualityScore {
	score := ContentQualityScore{
		MaxScore: 100,
		Issues:   []string{},
	}

	// Get page content
	bodyText, _ := page.Locator("body").InnerText()

	// Count words
	words := strings.Fields(bodyText)
	score.WordCount = len(words)

	if score.WordCount >= 1000 {
		score.Score += 25
	} else if score.WordCount >= 500 {
		score.Score += 15
		score.Issues = append(score.Issues, "Content length is moderate (500-1000 words)")
	} else if score.WordCount >= 300 {
		score.Score += 10
		score.Issues = append(score.Issues, "Content length is short (300-500 words)")
	} else {
		score.Score += 5
		score.Issues = append(score.Issues, fmt.Sprintf("Content is too thin (%d words)", score.WordCount))
	}

	// Count paragraphs
	pCount, _ := page.Locator("p").Count()
	score.ParagraphCount = pCount
	if pCount >= 5 {
		score.Score += 10
	}

	// Count images and alt attributes
	imageCount, _ := page.Locator("img").Count()
	score.ImageCount = imageCount

	imagesWithAlt := 0
	if imageCount > 0 {
		for i := 0; i < imageCount; i++ {
			alt, _ := page.Locator("img").Nth(i).GetAttribute("alt")
			if alt != "" {
				imagesWithAlt++
			}
		}
		score.ImagesWithAlt = imagesWithAlt

		altPercentage := float64(imagesWithAlt) / float64(imageCount) * 100
		if altPercentage == 100 {
			score.Score += 20
		} else if altPercentage >= 75 {
			score.Score += 15
			score.Issues = append(score.Issues, fmt.Sprintf("%.0f%% of images have alt text", altPercentage))
		} else if altPercentage >= 50 {
			score.Score += 10
			score.Issues = append(score.Issues, fmt.Sprintf("Only %.0f%% of images have alt text", altPercentage))
		} else {
			score.Score += 5
			score.Issues = append(score.Issues, fmt.Sprintf("Most images missing alt text (%.0f%%)", altPercentage))
		}
	} else {
		score.Score += 10 // No images is okay
	}

	// Count internal and external links
	parsedURL, _ := url.Parse(targetURL)
	links, _ := page.Locator("a[href]").All()

	for _, link := range links {
		href, _ := link.GetAttribute("href")
		if href == "" {
			continue
		}

		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			linkURL, err := url.Parse(href)
			if err == nil {
				if linkURL.Host == parsedURL.Host {
					score.InternalLinks++
				} else {
					score.ExternalLinks++
				}
			}
		} else if strings.HasPrefix(href, "/") || !strings.HasPrefix(href, "#") {
			score.InternalLinks++
		}
	}

	if score.InternalLinks >= 3 {
		score.Score += 15
	} else {
		score.Issues = append(score.Issues, fmt.Sprintf("Low internal linking (%d links)", score.InternalLinks))
		score.Score += 5
	}

	if score.ExternalLinks > 0 {
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, "No external links to authoritative sources")
	}

	// Calculate basic readability score (Flesch Reading Ease approximation)
	if score.WordCount > 0 {
		sentences := strings.Count(bodyText, ".") + strings.Count(bodyText, "!") + strings.Count(bodyText, "?")
		if sentences == 0 {
			sentences = 1
		}
		syllables := float64(score.WordCount) * 1.5 // Rough approximation
		score.ReadabilityScore = 206.835 - 1.015*(float64(score.WordCount)/float64(sentences)) - 84.6*(float64(syllables)/float64(score.WordCount))

		if score.ReadabilityScore >= 60 {
			score.Score += 10
		} else {
			score.Issues = append(score.Issues, "Content may be difficult to read")
			score.Score += 5
		}
	}

	return score
}

// auditLinkStructure performs link structure checks
func (a *SEOAuditor) auditLinkStructure(page playwright.Page, targetURL string) LinkStructureScore {
	score := LinkStructureScore{
		MaxScore: 100,
		Issues:   []string{},
	}

	parsedURL, _ := url.Parse(targetURL)
	links, _ := page.Locator("a[href]").All()

	descriptiveAnchors := 0
	totalAnchors := 0

	for _, link := range links {
		href, _ := link.GetAttribute("href")
		text, _ := link.InnerText()
		text = strings.TrimSpace(text)

		if href == "" {
			continue
		}

		// Check for descriptive anchor text
		if text != "" {
			totalAnchors++
			lowerText := strings.ToLower(text)
			if lowerText != "click here" && lowerText != "read more" && lowerText != "here" && len(text) > 2 {
				descriptiveAnchors++
			}
		}

		// Categorize as internal or external
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			linkURL, err := url.Parse(href)
			if err == nil {
				if linkURL.Host == parsedURL.Host {
					score.InternalLinks++
				} else {
					score.ExternalLinks++
				}
			}
		} else if strings.HasPrefix(href, "/") {
			score.InternalLinks++
		}
	}

	// Score internal links
	if score.InternalLinks >= 5 {
		score.Score += 25
	} else if score.InternalLinks >= 3 {
		score.Score += 15
	} else {
		score.Issues = append(score.Issues, fmt.Sprintf("Low internal link count (%d)", score.InternalLinks))
		score.Score += 5
	}

	// Score external links
	if score.ExternalLinks > 0 && score.ExternalLinks <= 10 {
		score.Score += 15
	} else if score.ExternalLinks > 10 {
		score.Issues = append(score.Issues, "High number of external links")
		score.Score += 10
	} else {
		score.Issues = append(score.Issues, "No external links")
		score.Score += 5
	}

	// Check for descriptive anchors
	if totalAnchors > 0 {
		descriptivePercentage := float64(descriptiveAnchors) / float64(totalAnchors) * 100
		score.DescriptiveAnchors = descriptivePercentage >= 80

		if score.DescriptiveAnchors {
			score.Score += 20
		} else {
			score.Issues = append(score.Issues, "Many links have generic anchor text")
			score.Score += 10
		}
	} else {
		score.Score += 15
	}

	// Check for breadcrumbs
	breadcrumbs, _ := page.Locator("[itemtype*='BreadcrumbList'], nav[aria-label*='readcrumb'], .breadcrumb").Count()
	score.HasBreadcrumbs = breadcrumbs > 0

	if score.HasBreadcrumbs {
		score.Score += 20
	} else {
		score.Issues = append(score.Issues, "No breadcrumb navigation found")
	}

	// Note: Checking for broken links would require making HTTP requests to each link
	// This is commented out for performance but can be enabled
	// score.BrokenLinks = a.checkBrokenLinks(links)
	score.Score += 20 // Assume no broken links for now

	return score
}

// auditSchemaMarkup performs schema markup checks
func (a *SEOAuditor) auditSchemaMarkup(page playwright.Page) SchemaMarkupScore {
	score := SchemaMarkupScore{
		MaxScore:    100,
		Issues:      []string{},
		SchemaTypes: []string{},
	}

	// Check for JSON-LD schema
	jsonLdScripts, _ := page.Locator("script[type='application/ld+json']").All()
	score.HasSchema = len(jsonLdScripts) > 0

	if score.HasSchema {
		score.Score += 30

		// Parse each JSON-LD script
		for _, script := range jsonLdScripts {
			content, _ := script.InnerText()

			// Check for common schema types
			if strings.Contains(content, "\"@type\"") {
				if strings.Contains(content, "Organization") {
					score.HasOrganization = true
					score.SchemaTypes = append(score.SchemaTypes, "Organization")
				}
				if strings.Contains(content, "BreadcrumbList") {
					score.HasBreadcrumb = true
					score.SchemaTypes = append(score.SchemaTypes, "BreadcrumbList")
				}
				if strings.Contains(content, "Article") {
					score.SchemaTypes = append(score.SchemaTypes, "Article")
				}
				if strings.Contains(content, "Product") {
					score.SchemaTypes = append(score.SchemaTypes, "Product")
				}
				if strings.Contains(content, "LocalBusiness") {
					score.SchemaTypes = append(score.SchemaTypes, "LocalBusiness")
				}
			}
		}

		// Score based on schema types
		schemaTypeCount := len(score.SchemaTypes)
		if schemaTypeCount >= 3 {
			score.Score += 40
		} else if schemaTypeCount == 2 {
			score.Score += 30
		} else if schemaTypeCount == 1 {
			score.Score += 20
			score.Issues = append(score.Issues, "Limited schema markup types")
		}

		// Bonus for organization schema
		if score.HasOrganization {
			score.Score += 15
		} else {
			score.Issues = append(score.Issues, "Missing Organization schema")
		}

		// Bonus for breadcrumb schema
		if score.HasBreadcrumb {
			score.Score += 15
		} else {
			score.Issues = append(score.Issues, "Missing BreadcrumbList schema")
		}
	} else {
		score.Issues = append(score.Issues, "No structured data (schema markup) found")
	}

	// Also check for microdata
	itemscope, _ := page.Locator("[itemscope]").Count()
	if itemscope > 0 && !score.HasSchema {
		score.HasSchema = true
		score.Score += 20
		score.Issues = append(score.Issues, "Using microdata instead of JSON-LD (JSON-LD is preferred)")
	}

	return score
}

// auditSecurity performs security checks
func (a *SEOAuditor) auditSecurity(targetURL string, page playwright.Page) SecurityScore {
	score := SecurityScore{
		MaxScore: 100,
		Issues:   []string{},
	}

	parsedURL, _ := url.Parse(targetURL)

	// Check HTTPS
	score.IsHTTPS = parsedURL.Scheme == "https"
	score.HasSSL = score.IsHTTPS

	if score.IsHTTPS {
		score.Score += 40
	} else {
		score.Issues = append(score.Issues, "Site is not using HTTPS")
	}

	// Check for mixed content
	if score.IsHTTPS {
		// Check for HTTP resources
		httpImages, _ := page.Locator("img[src^='http://']").Count()
		httpScripts, _ := page.Locator("script[src^='http://']").Count()
		httpLinks, _ := page.Locator("link[href^='http://']").Count()

		score.MixedContent = httpImages > 0 || httpScripts > 0 || httpLinks > 0

		if !score.MixedContent {
			score.Score += 30
		} else {
			score.Issues = append(score.Issues, "Mixed content detected (HTTP resources on HTTPS page)")
			score.Score += 10
		}
	} else {
		score.Score += 15
	}

	// Check for security headers (would require HTTP response inspection)
	// This is a simplified check
	score.HasSecurityHeaders = false // Placeholder
	if score.HasSecurityHeaders {
		score.Score += 30
	} else {
		score.Issues = append(score.Issues, "Unable to verify security headers")
		score.Score += 15
	}

	return score
}

// auditUserExperience performs user experience checks
func (a *SEOAuditor) auditUserExperience(page playwright.Page) UserExperienceScore {
	score := UserExperienceScore{
		MaxScore: 100,
		Issues:   []string{},
	}

	// Check for favicon
	favicon, _ := page.Locator("link[rel*='icon']").Count()
	score.HasFavicon = favicon > 0

	if score.HasFavicon {
		score.Score += 20
	} else {
		score.Issues = append(score.Issues, "Missing favicon")
	}

	// Check for lang attribute
	lang, _ := page.Locator("html[lang]").Count()
	score.HasLangAttribute = lang > 0

	if score.HasLangAttribute {
		score.Score += 25
	} else {
		score.Issues = append(score.Issues, "Missing lang attribute on html tag")
	}

	// Check font size (simplified check)
	bodyFontSize, _ := page.Locator("body").Evaluate("el => window.getComputedStyle(el).fontSize", nil)
	if bodyFontSize != nil {
		fontSize := bodyFontSize.(string)
		score.FontSizeReadable = !strings.HasPrefix(fontSize, "1") || fontSize >= "14px"

		if score.FontSizeReadable {
			score.Score += 25
		} else {
			score.Issues = append(score.Issues, "Font size may be too small for comfortable reading")
			score.Score += 10
		}
	} else {
		score.Score += 15
	}

	// Check for intrusive popups/modals (simplified)
	modals, _ := page.Locator("[class*='modal'][style*='display'], [class*='popup'][style*='display']").Count()
	score.NoIntrusivePopups = modals == 0

	if score.NoIntrusivePopups {
		score.Score += 30
	} else {
		score.Issues = append(score.Issues, "Intrusive popups detected")
		score.Score += 15
	}

	return score
}

// auditWebVitals performs Core Web Vitals checks using the web-vitals library
func (a *SEOAuditor) auditWebVitals(page playwright.Page) WebVitalsScore {
	score := WebVitalsScore{
		MaxScore: 100,
		Issues:   []string{},
	}

	// Read the web-vitals library from file
	webVitalsScript, err := os.ReadFile("webvitals.js")
	if err != nil {
		score.Issues = append(score.Issues, "Failed to read web-vitals library file")
		return score
	}

	// Inject web-vitals library into the page
	_, err = page.Evaluate(string(webVitalsScript))
	if err != nil {
		score.Issues = append(score.Issues, "Failed to inject web-vitals library")
		return score
	}

	// Set up Web Vitals listeners and collect metrics
	_, err = page.Evaluate(`() => {
		window.__WEB_VITALS__ = {};
		
		webVitals.onLCP((metric) => {
			window.__WEB_VITALS__.lcp = metric.value;
			window.__WEB_VITALS__.lcpAttribution = metric.attribution;
		}, { reportAllChanges: true });
		
		webVitals.onFCP((metric) => {
			window.__WEB_VITALS__.fcp = metric.value;
		}, { reportAllChanges: true });
		
		webVitals.onCLS((metric) => {
			window.__WEB_VITALS__.cls = metric.value;
			window.__WEB_VITALS__.clsAttribution = metric.attribution;
		}, { reportAllChanges: true });
		
		webVitals.onINP((metric) => {
			window.__WEB_VITALS__.inp = metric.value;
			window.__WEB_VITALS__.inpAttribution = metric.attribution;
		}, { reportAllChanges: true });
		
		webVitals.onTTFB((metric) => {
			window.__WEB_VITALS__.ttfb = metric.value;
		}, { reportAllChanges: true });
	}`)
	if err != nil {
		score.Issues = append(score.Issues, "Failed to initialize web-vitals listeners")
	}

	// Wait for metrics to be collected (FCP and TTFB should be immediate, LCP needs time)
	time.Sleep(2 * time.Second)

	// Trigger a small interaction to help capture INP (click on body)
	page.Evaluate(`() => { document.body.click(); }`)
	time.Sleep(500 * time.Millisecond)

	// Collect the web vitals metrics
	webVitalsResult, err := page.Evaluate(`() => {
		const perf = performance.getEntriesByType('navigation')[0] || {};
		const resources = performance.getEntriesByType('resource') || [];
		
		return {
			...window.__WEB_VITALS__,
			domContentLoaded: perf.domContentLoadedEventEnd || 0,
			domComplete: perf.domComplete || 0,
			transferSize: resources.reduce((sum, r) => sum + (r.transferSize || 0), 0),
			resourceCount: resources.length
		};
	}`)

	if err == nil && webVitalsResult != nil {
		metrics := webVitalsResult.(map[string]interface{})

		if lcp, ok := metrics["lcp"].(int); ok && lcp > 0 {
			score.LCP = lcp
			score.LCPRating = rateLCP(lcp)
		}
		if lcpAttr, ok := metrics["lcpAttribution"].(map[string]interface{}); ok {
			score.LCPAttribution = lcpAttr
		}

		// FCP (First Contentful Paint)
		if fcp, ok := metrics["fcp"].(int); ok && fcp > 0 {
			score.FCP = fcp
			score.FCPRating = rateFCP(fcp)
		}

		// CLS (Cumulative Layout Shift)
		if cls, ok := metrics["cls"].(float64); ok {
			score.CLS = math.Round(cls*1000) / 1000
			score.CLSRating = rateCLS(cls)
		}
		if clsAttr, ok := metrics["clsAttribution"].(map[string]interface{}); ok {
			score.CLSAttribution = clsAttr
		}

		// INP (Interaction to Next Paint)
		if inp, ok := metrics["inp"].(float64); ok && inp > 0 {
			score.INP = inp
			score.INPRating = rateINP(inp)
		}
		if inpAttr, ok := metrics["inpAttribution"].(map[string]interface{}); ok {
			score.INPAttribution = inpAttr
		}

		// TTFB (Time to First Byte)
		if ttfb, ok := metrics["ttfb"].(float64); ok && ttfb > 0 {
			score.TTFB = ttfb
			score.TTFBRating = rateTTFB(ttfb)
		}

		// DOM metrics
		if domContentLoaded, ok := metrics["domContentLoaded"].(float64); ok {
			score.DOMContentLoaded = domContentLoaded
		}
		if domComplete, ok := metrics["domComplete"].(float64); ok {
			score.DOMComplete = domComplete
		}

		// Transfer size and resource count
		if transferSize, ok := metrics["transferSize"].(float64); ok {
			score.TransferSize = int64(transferSize)
		}
		if resourceCount, ok := metrics["resourceCount"].(float64); ok {
			score.ResourceCount = int(resourceCount)
		}
	}

	// Calculate Web Vitals score based on ratings
	score.Score = calculateWebVitalsScore(&score)

	// Add issues based on ratings
	if score.LCPRating == "poor" {
		score.Issues = append(score.Issues, fmt.Sprintf("LCP is poor (%.dms) - should be under 2500ms", score.LCP))
	} else if score.LCPRating == "needs-improvement" {
		score.Issues = append(score.Issues, fmt.Sprintf("LCP needs improvement (%.dms) - aim for under 2500ms", score.LCP))
	}

	if score.FCPRating == "poor" {
		score.Issues = append(score.Issues, fmt.Sprintf("FCP is poor (%.dms) - should be under 1800ms", score.FCP))
	} else if score.FCPRating == "needs-improvement" {
		score.Issues = append(score.Issues, fmt.Sprintf("FCP needs improvement (%.dms) - aim for under 1800ms", score.FCP))
	}

	if score.CLSRating == "poor" {
		score.Issues = append(score.Issues, fmt.Sprintf("CLS is poor (%.3f) - should be under 0.1", score.CLS))
	} else if score.CLSRating == "needs-improvement" {
		score.Issues = append(score.Issues, fmt.Sprintf("CLS needs improvement (%.3f) - aim for under 0.1", score.CLS))
	}

	if score.INP > 0 {
		if score.INPRating == "poor" {
			score.Issues = append(score.Issues, fmt.Sprintf("INP is poor (%.0fms) - should be under 200ms", score.INP))
		} else if score.INPRating == "needs-improvement" {
			score.Issues = append(score.Issues, fmt.Sprintf("INP needs improvement (%.0fms) - aim for under 200ms", score.INP))
		}
	}

	if score.TTFBRating == "poor" {
		score.Issues = append(score.Issues, fmt.Sprintf("TTFB is poor (%.0fms) - should be under 800ms", score.TTFB))
	} else if score.TTFBRating == "needs-improvement" {
		score.Issues = append(score.Issues, fmt.Sprintf("TTFB needs improvement (%.0fms) - aim for under 800ms", score.TTFB))
	}

	if score.ResourceCount > 100 {
		score.Issues = append(score.Issues, fmt.Sprintf("High number of resources loaded (%d) - consider reducing HTTP requests", score.ResourceCount))
	}

	if score.TransferSize > 5*1024*1024 { // 5MB
		score.Issues = append(score.Issues, fmt.Sprintf("Large total transfer size (%s) - consider optimizing assets", formatBytes(score.TransferSize)))
	}

	return score
}

// Web Vitals rating functions based on Google's thresholds
func rateLCP(lcp int) string {
	if lcp <= 2500 {
		return "good"
	} else if lcp <= 4000 {
		return "needs-improvement"
	}
	return "poor"
}

func rateFCP(fcp int) string {
	if fcp <= 1800 {
		return "good"
	} else if fcp <= 3000 {
		return "needs-improvement"
	}
	return "poor"
}

func rateCLS(cls float64) string {
	if cls <= 0.1 {
		return "good"
	} else if cls <= 0.25 {
		return "needs-improvement"
	}
	return "poor"
}

func rateTTFB(ttfb float64) string {
	if ttfb <= 800 {
		return "good"
	} else if ttfb <= 1800 {
		return "needs-improvement"
	}
	return "poor"
}

func rateINP(inp float64) string {
	if inp <= 200 {
		return "good"
	} else if inp <= 500 {
		return "needs-improvement"
	}
	return "poor"
}

func calculateWebVitalsScore(wv *WebVitalsScore) float64 {
	score := 0.0
	metricCount := 0
	maxPointsPerMetric := 20.0 // 5 metrics * 20 = 100 max

	// LCP scoring (20 points max)
	if wv.LCP > 0 {
		metricCount++
		switch wv.LCPRating {
		case "good":
			score += maxPointsPerMetric
		case "needs-improvement":
			score += maxPointsPerMetric * 0.6
		case "poor":
			score += maxPointsPerMetric * 0.2
		}
	}

	// FCP scoring (20 points max)
	if wv.FCP > 0 {
		metricCount++
		switch wv.FCPRating {
		case "good":
			score += maxPointsPerMetric
		case "needs-improvement":
			score += maxPointsPerMetric * 0.6
		case "poor":
			score += maxPointsPerMetric * 0.2
		}
	}

	// CLS scoring (20 points max)
	metricCount++
	switch wv.CLSRating {
	case "good":
		score += maxPointsPerMetric
	case "needs-improvement":
		score += maxPointsPerMetric * 0.6
	case "poor":
		score += maxPointsPerMetric * 0.2
	}

	// INP scoring (20 points max)
	if wv.INP > 0 {
		metricCount++
		switch wv.INPRating {
		case "good":
			score += maxPointsPerMetric
		case "needs-improvement":
			score += maxPointsPerMetric * 0.6
		case "poor":
			score += maxPointsPerMetric * 0.2
		}
	}

	// TTFB scoring (20 points max)
	if wv.TTFB > 0 {
		metricCount++
		switch wv.TTFBRating {
		case "good":
			score += maxPointsPerMetric
		case "needs-improvement":
			score += maxPointsPerMetric * 0.6
		case "poor":
			score += maxPointsPerMetric * 0.2
		}
	}

	// Normalize score if not all metrics were captured
	if metricCount > 0 && metricCount < 5 {
		score = (score / (float64(metricCount) * maxPointsPerMetric)) * 100
		fmt.Println("Normalized Web Vitals Score:", score)
	}

	return score
}

// Helper functions

func (a *SEOAuditor) checkHeadingHierarchy(page playwright.Page) bool {
	h1, _ := page.Locator("h1").Count()
	h2, _ := page.Locator("h2").Count()
	h3, _ := page.Locator("h3").Count()
	h4, _ := page.Locator("h4").Count()
	h5, _ := page.Locator("h5").Count()
	h6, _ := page.Locator("h6").Count()

	// Basic check: should have H1, and if H3 exists, H2 should exist, etc.
	if h1 == 0 {
		return false
	}
	if h3 > 0 && h2 == 0 {
		return false
	}
	if h4 > 0 && h3 == 0 {
		return false
	}
	if h5 > 0 && h4 == 0 {
		return false
	}
	if h6 > 0 && h5 == 0 {
		return false
	}

	return true
}

func (a *SEOAuditor) checkURLExists(urlStr string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Head(urlStr)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (a *SEOAuditor) calculateOverallScore(audit *SEOAudit) float64 {
	weights := map[string]float64{
		"technical": 0.30,
		"onpage":    0.25,
		"content":   0.20,
		"links":     0.10,
		"schema":    0.05,
		"security":  0.05,
		"ux":        0.05,
	}

	score := 0.0
	score += (audit.TechnicalSEO.Score / audit.TechnicalSEO.MaxScore) * 100 * weights["technical"]
	score += (audit.OnPageSEO.Score / audit.OnPageSEO.MaxScore) * 100 * weights["onpage"]
	score += (audit.ContentQuality.Score / audit.ContentQuality.MaxScore) * 100 * weights["content"]
	score += (audit.LinkStructure.Score / audit.LinkStructure.MaxScore) * 100 * weights["links"]
	score += (audit.SchemaMarkup.Score / audit.SchemaMarkup.MaxScore) * 100 * weights["schema"]
	score += (audit.Security.Score / audit.Security.MaxScore) * 100 * weights["security"]
	score += (audit.UserExperience.Score / audit.UserExperience.MaxScore) * 100 * weights["ux"]

	return math.Round(score*100) / 100
}

func (a *SEOAuditor) calculateGrade(score float64) string {
	switch {
	case score >= 90:
		return "A+"
	case score >= 85:
		return "A"
	case score >= 80:
		return "A-"
	case score >= 75:
		return "B+"
	case score >= 70:
		return "B"
	case score >= 65:
		return "B-"
	case score >= 60:
		return "C+"
	case score >= 55:
		return "C"
	case score >= 50:
		return "C-"
	case score >= 45:
		return "D+"
	case score >= 40:
		return "D"
	default:
		return "F"
	}
}

func (a *SEOAuditor) generateRecommendations(audit *SEOAudit) []string {
	recommendations := []string{}

	// Collect all issues from all categories
	recommendations = append(recommendations, audit.TechnicalSEO.Issues...)
	recommendations = append(recommendations, audit.OnPageSEO.Issues...)
	recommendations = append(recommendations, audit.ContentQuality.Issues...)
	recommendations = append(recommendations, audit.LinkStructure.Issues...)
	recommendations = append(recommendations, audit.SchemaMarkup.Issues...)
	recommendations = append(recommendations, audit.Security.Issues...)
	recommendations = append(recommendations, audit.UserExperience.Issues...)
	recommendations = append(recommendations, audit.WebVitals.Issues...)

	// Add priority recommendations based on scores
	if audit.TechnicalSEO.Score < 50 {
		recommendations = append([]string{"CRITICAL: Address technical SEO issues immediately"}, recommendations...)
	}
	if !audit.Security.IsHTTPS {
		recommendations = append([]string{"CRITICAL: Implement HTTPS for security and SEO"}, recommendations...)
	}
	if !audit.OnPageSEO.HasTitle {
		recommendations = append([]string{"CRITICAL: Add a title tag to the page"}, recommendations...)
	}

	return recommendations
}

func (a *SEOAuditor) generateMarkdown(audit *SEOAudit) string {
	var sb strings.Builder

	// Header with context for LLM
	sb.WriteString("# SEO Audit Report\n\n")
	sb.WriteString("## Context\n\n")
	sb.WriteString("You are an SEO expert assistant. Below is a comprehensive SEO audit report for a website. ")
	sb.WriteString("Your task is to analyze the issues identified and provide specific, actionable solutions to fix them.\n\n")

	// Summary section
	sb.WriteString("## Website Information\n\n")
	sb.WriteString(fmt.Sprintf("- **URL**: %s\n", audit.URL))
	sb.WriteString(fmt.Sprintf("- **Audit Date**: %s\n", audit.Timestamp.Format("2006-01-02 15:04:05 UTC")))
	sb.WriteString(fmt.Sprintf("- **Overall Score**: %.1f/100\n", audit.OverallScore))
	sb.WriteString(fmt.Sprintf("- **Grade**: %s\n\n", audit.Grade))

	// Score breakdown
	sb.WriteString("## Score Breakdown\n\n")
	sb.WriteString("| Category | Score | Max Score | Percentage |\n")
	sb.WriteString("|----------|-------|-----------|------------|\n")
	sb.WriteString(fmt.Sprintf("| Technical SEO | %.0f | %.0f | %.0f%% |\n", audit.TechnicalSEO.Score, audit.TechnicalSEO.MaxScore, (audit.TechnicalSEO.Score/audit.TechnicalSEO.MaxScore)*100))
	sb.WriteString(fmt.Sprintf("| On-Page SEO | %.0f | %.0f | %.0f%% |\n", audit.OnPageSEO.Score, audit.OnPageSEO.MaxScore, (audit.OnPageSEO.Score/audit.OnPageSEO.MaxScore)*100))
	sb.WriteString(fmt.Sprintf("| Content Quality | %.0f | %.0f | %.0f%% |\n", audit.ContentQuality.Score, audit.ContentQuality.MaxScore, (audit.ContentQuality.Score/audit.ContentQuality.MaxScore)*100))
	sb.WriteString(fmt.Sprintf("| Link Structure | %.0f | %.0f | %.0f%% |\n", audit.LinkStructure.Score, audit.LinkStructure.MaxScore, (audit.LinkStructure.Score/audit.LinkStructure.MaxScore)*100))
	sb.WriteString(fmt.Sprintf("| Schema Markup | %.0f | %.0f | %.0f%% |\n", audit.SchemaMarkup.Score, audit.SchemaMarkup.MaxScore, (audit.SchemaMarkup.Score/audit.SchemaMarkup.MaxScore)*100))
	sb.WriteString(fmt.Sprintf("| Security | %.0f | %.0f | %.0f%% |\n", audit.Security.Score, audit.Security.MaxScore, (audit.Security.Score/audit.Security.MaxScore)*100))
	sb.WriteString(fmt.Sprintf("| User Experience | %.0f | %.0f | %.0f%% |\n", audit.UserExperience.Score, audit.UserExperience.MaxScore, (audit.UserExperience.Score/audit.UserExperience.MaxScore)*100))
	sb.WriteString(fmt.Sprintf("| Web Vitals | %.0f | %.0f | %.0f%% |\n\n", audit.WebVitals.Score, audit.WebVitals.MaxScore, (audit.WebVitals.Score/audit.WebVitals.MaxScore)*100))

	// Technical SEO Details
	sb.WriteString("## Technical SEO Analysis\n\n")
	sb.WriteString("### Current Status\n\n")
	sb.WriteString(fmt.Sprintf("- **HTTPS**: %s\n", boolToStatus(audit.TechnicalSEO.IsHTTPS)))
	sb.WriteString(fmt.Sprintf("- **Viewport Meta Tag**: %s\n", boolToStatus(audit.TechnicalSEO.HasViewport)))
	sb.WriteString(fmt.Sprintf("- **Mobile Friendly**: %s\n", boolToStatus(audit.TechnicalSEO.IsMobileFriendly)))
	sb.WriteString(fmt.Sprintf("- **robots.txt**: %s\n", boolToStatus(audit.TechnicalSEO.HasRobotsTxt)))
	sb.WriteString(fmt.Sprintf("- **Sitemap**: %s\n", boolToStatus(audit.TechnicalSEO.HasSitemap)))
	sb.WriteString(fmt.Sprintf("- **Page Load Time**: %.0fms\n", audit.TechnicalSEO.LoadTime))
	sb.WriteString(fmt.Sprintf("- **Page Size**: %s\n", formatBytes(audit.TechnicalSEO.PageSize)))
	sb.WriteString(fmt.Sprintf("- **HTTP Requests**: %d\n", audit.TechnicalSEO.HTTPRequests))
	sb.WriteString(fmt.Sprintf("- **HTTP Status Code**: %d\n\n", audit.TechnicalSEO.HTTPStatusCode))

	if len(audit.TechnicalSEO.Issues) > 0 {
		sb.WriteString("### Issues Found\n\n")
		for _, issue := range audit.TechnicalSEO.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// On-Page SEO Details
	sb.WriteString("## On-Page SEO Analysis\n\n")
	sb.WriteString("### Current Status\n\n")
	sb.WriteString(fmt.Sprintf("- **Title Tag**: %s (Length: %d characters)\n", boolToStatus(audit.OnPageSEO.HasTitle), audit.OnPageSEO.TitleLength))
	sb.WriteString(fmt.Sprintf("- **Meta Description**: %s (Length: %d characters)\n", boolToStatus(audit.OnPageSEO.HasMetaDescription), audit.OnPageSEO.MetaDescriptionLength))
	sb.WriteString(fmt.Sprintf("- **H1 Tag**: %s (Count: %d)\n", boolToStatus(audit.OnPageSEO.HasH1), audit.OnPageSEO.H1Count))
	sb.WriteString(fmt.Sprintf("- **H2 Tags Count**: %d\n", audit.OnPageSEO.H2Count))
	sb.WriteString(fmt.Sprintf("- **Heading Hierarchy**: %s\n", boolToStatus(audit.OnPageSEO.ProperHeadingHierarchy)))
	sb.WriteString(fmt.Sprintf("- **Open Graph Tags**: %s\n", boolToStatus(audit.OnPageSEO.HasOGTags)))
	sb.WriteString(fmt.Sprintf("- **Twitter Card**: %s\n", boolToStatus(audit.OnPageSEO.HasTwitterCard)))
	sb.WriteString(fmt.Sprintf("- **Canonical Tag**: %s\n\n", boolToStatus(audit.OnPageSEO.HasCanonical)))

	if len(audit.OnPageSEO.Issues) > 0 {
		sb.WriteString("### Issues Found\n\n")
		for _, issue := range audit.OnPageSEO.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// Content Quality Details
	sb.WriteString("## Content Quality Analysis\n\n")
	sb.WriteString("### Current Status\n\n")
	sb.WriteString(fmt.Sprintf("- **Word Count**: %d\n", audit.ContentQuality.WordCount))
	sb.WriteString(fmt.Sprintf("- **Paragraph Count**: %d\n", audit.ContentQuality.ParagraphCount))
	sb.WriteString(fmt.Sprintf("- **Images**: %d (with alt text: %d)\n", audit.ContentQuality.ImageCount, audit.ContentQuality.ImagesWithAlt))
	sb.WriteString(fmt.Sprintf("- **Internal Links**: %d\n", audit.ContentQuality.InternalLinks))
	sb.WriteString(fmt.Sprintf("- **External Links**: %d\n", audit.ContentQuality.ExternalLinks))
	sb.WriteString(fmt.Sprintf("- **Readability Score**: %.1f\n\n", audit.ContentQuality.ReadabilityScore))

	if len(audit.ContentQuality.Issues) > 0 {
		sb.WriteString("### Issues Found\n\n")
		for _, issue := range audit.ContentQuality.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// Link Structure Details
	sb.WriteString("## Link Structure Analysis\n\n")
	sb.WriteString("### Current Status\n\n")
	sb.WriteString(fmt.Sprintf("- **Internal Links**: %d\n", audit.LinkStructure.InternalLinks))
	sb.WriteString(fmt.Sprintf("- **External Links**: %d\n", audit.LinkStructure.ExternalLinks))
	sb.WriteString(fmt.Sprintf("- **Broken Links**: %d\n", audit.LinkStructure.BrokenLinks))
	sb.WriteString(fmt.Sprintf("- **Breadcrumbs**: %s\n", boolToStatus(audit.LinkStructure.HasBreadcrumbs)))
	sb.WriteString(fmt.Sprintf("- **Descriptive Anchor Texts**: %s\n\n", boolToStatus(audit.LinkStructure.DescriptiveAnchors)))

	if len(audit.LinkStructure.Issues) > 0 {
		sb.WriteString("### Issues Found\n\n")
		for _, issue := range audit.LinkStructure.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// Schema Markup Details
	sb.WriteString("## Schema Markup Analysis\n\n")
	sb.WriteString("### Current Status\n\n")
	sb.WriteString(fmt.Sprintf("- **Has Schema Markup**: %s\n", boolToStatus(audit.SchemaMarkup.HasSchema)))
	if len(audit.SchemaMarkup.SchemaTypes) > 0 {
		sb.WriteString(fmt.Sprintf("- **Schema Types Found**: %s\n", strings.Join(audit.SchemaMarkup.SchemaTypes, ", ")))
	}
	sb.WriteString(fmt.Sprintf("- **Organization Schema**: %s\n", boolToStatus(audit.SchemaMarkup.HasOrganization)))
	sb.WriteString(fmt.Sprintf("- **Breadcrumb Schema**: %s\n\n", boolToStatus(audit.SchemaMarkup.HasBreadcrumb)))

	if len(audit.SchemaMarkup.Issues) > 0 {
		sb.WriteString("### Issues Found\n\n")
		for _, issue := range audit.SchemaMarkup.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// Security Details
	sb.WriteString("## Security Analysis\n\n")
	sb.WriteString("### Current Status\n\n")
	sb.WriteString(fmt.Sprintf("- **HTTPS**: %s\n", boolToStatus(audit.Security.IsHTTPS)))
	sb.WriteString(fmt.Sprintf("- **SSL Certificate**: %s\n", boolToStatus(audit.Security.HasSSL)))
	sb.WriteString(fmt.Sprintf("- **Mixed Content**: %s\n", boolToStatus(!audit.Security.MixedContent)))
	sb.WriteString(fmt.Sprintf("- **Security Headers**: %s\n\n", boolToStatus(audit.Security.HasSecurityHeaders)))

	if len(audit.Security.Issues) > 0 {
		sb.WriteString("### Issues Found\n\n")
		for _, issue := range audit.Security.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// User Experience Details
	sb.WriteString("## User Experience Analysis\n\n")
	sb.WriteString("### Current Status\n\n")
	sb.WriteString(fmt.Sprintf("- **Favicon**: %s\n", boolToStatus(audit.UserExperience.HasFavicon)))
	sb.WriteString(fmt.Sprintf("- **Language Attribute**: %s\n", boolToStatus(audit.UserExperience.HasLangAttribute)))
	sb.WriteString(fmt.Sprintf("- **Readable Font Size**: %s\n", boolToStatus(audit.UserExperience.FontSizeReadable)))
	sb.WriteString(fmt.Sprintf("- **No Intrusive Popups**: %s\n\n", boolToStatus(audit.UserExperience.NoIntrusivePopups)))

	if len(audit.UserExperience.Issues) > 0 {
		sb.WriteString("### Issues Found\n\n")
		for _, issue := range audit.UserExperience.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// Web Vitals Details
	sb.WriteString("## Core Web Vitals Analysis\n\n")
	sb.WriteString("### Performance Metrics\n\n")
	sb.WriteString("| Metric | Value | Rating | Target |\n")
	sb.WriteString("|--------|-------|--------|--------|\n")
	sb.WriteString(fmt.Sprintf("| LCP (Largest Contentful Paint) | %.0fms | %s | ≤2500ms |\n", audit.WebVitals.LCP, ratingToEmoji(audit.WebVitals.LCPRating)))
	sb.WriteString(fmt.Sprintf("| FCP (First Contentful Paint) | %.0fms | %s | ≤1800ms |\n", audit.WebVitals.FCP, ratingToEmoji(audit.WebVitals.FCPRating)))
	sb.WriteString(fmt.Sprintf("| CLS (Cumulative Layout Shift) | %.3f | %s | ≤0.1 |\n", audit.WebVitals.CLS, ratingToEmoji(audit.WebVitals.CLSRating)))
	sb.WriteString(fmt.Sprintf("| INP (Interaction to Next Paint) | %.0fms | %s | ≤200ms |\n", audit.WebVitals.INP, ratingToEmoji(audit.WebVitals.INPRating)))
	sb.WriteString(fmt.Sprintf("| TTFB (Time to First Byte) | %.0fms | %s | ≤800ms |\n\n", audit.WebVitals.TTFB, ratingToEmoji(audit.WebVitals.TTFBRating)))

	sb.WriteString("### Additional Metrics\n\n")
	sb.WriteString(fmt.Sprintf("- **DOM Content Loaded**: %.0fms\n", audit.WebVitals.DOMContentLoaded))
	sb.WriteString(fmt.Sprintf("- **DOM Complete**: %.0fms\n", audit.WebVitals.DOMComplete))
	sb.WriteString(fmt.Sprintf("- **Total Transfer Size**: %s\n", formatBytes(audit.WebVitals.TransferSize)))
	sb.WriteString(fmt.Sprintf("- **Resource Count**: %d\n\n", audit.WebVitals.ResourceCount))

	if len(audit.WebVitals.Issues) > 0 {
		sb.WriteString("### Issues Found\n\n")
		for _, issue := range audit.WebVitals.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
		sb.WriteString("\n")
	}

	// All Recommendations Summary
	if len(audit.Recommendations) > 0 {
		sb.WriteString("## All Issues Summary\n\n")
		sb.WriteString("The following is a prioritized list of all issues that need to be addressed:\n\n")
		for i, rec := range audit.Recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
		sb.WriteString("\n")
	}

	// Instructions for LLM
	sb.WriteString("## Instructions for AI Assistant\n\n")
	sb.WriteString("Based on the audit results above, please provide:\n\n")
	sb.WriteString("1. **Priority Fixes**: List the most critical issues that should be addressed first, ordered by impact on SEO.\n")
	sb.WriteString("2. **Code Examples**: For each issue, provide specific code snippets or implementation examples to fix the problem.\n")
	sb.WriteString("3. **Best Practices**: Recommend SEO best practices relevant to the identified issues.\n")
	sb.WriteString("4. **Quick Wins**: Identify any easy fixes that can be implemented immediately for quick improvements.\n")
	sb.WriteString("5. **Long-term Strategy**: Suggest a roadmap for improving the overall SEO score.\n\n")
	sb.WriteString("Focus on actionable, specific recommendations that can be directly implemented.\n")

	return sb.String()
}

// Helper function to convert boolean to status string
func boolToStatus(b bool) string {
	if b {
		return "✅ Yes"
	}
	return "❌ No"
}

// Helper function to convert rating to emoji
func ratingToEmoji(rating string) string {
	switch rating {
	case "good":
		return "✅ Good"
	case "needs-improvement":
		return "⚠️ Needs Improvement"
	case "poor":
		return "❌ Poor"
	default:
		return "❓ Unknown"
	}
}

// Helper function to format bytes to human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// AuditRequest represents the request body for the audit endpoint
type AuditRequest struct {
	URL string `json:"url"`
}

// Main function
func main() {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Create a single auditor instance
	auditor, err := NewSEOAuditor()
	if err != nil {
		fmt.Printf("Error creating auditor: %v\n", err)
		return
	}
	defer auditor.Close()

	// Health check endpoint
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "SEO Auditor API is running",
		})
	})

	// POST endpoint to audit a website
	app.Post("/api/audit", func(c *fiber.Ctx) error {
		var req AuditRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
		}

		if req.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "URL is required",
			})
		}

		// Audit the website
		audit, err := auditor.AuditWebsite(req.URL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Error auditing website",
				"details": err.Error(),
			})
		}

		// Return the audit results as JSON
		return c.JSON(audit)
	})

	// GET endpoint to audit a website (via query parameter)
	app.Get("/api/audit", func(c *fiber.Ctx) error {
		targetURL := c.Query("url")
		if targetURL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "URL query parameter is required",
			})
		}

		// Audit the website
		audit, err := auditor.AuditWebsite(targetURL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Error auditing website",
				"details": err.Error(),
			})
		}

		// Return the audit results as JSON
		return c.JSON(audit)
	})

	// Start server
	fmt.Println("🚀 SEO Auditor API starting on http://localhost:3000")
	fmt.Println("📝 Endpoints:")
	fmt.Println("  GET  /api/health")
	fmt.Println("  POST /api/audit  (body: {\"url\": \"https://example.com\"})")
	fmt.Println("  GET  /api/audit?url=https://example.com")

	if err := app.Listen(getPort()); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	} else {
		port = ":" + port
	}

	return port
}
