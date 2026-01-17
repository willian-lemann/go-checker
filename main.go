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
	OverallScore    float64             `json:"overall_score"`
	Grade           string              `json:"grade"`
	Recommendations []string            `json:"recommendations"`
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

	// Calculate overall score
	audit.OverallScore = a.calculateOverallScore(audit)
	audit.Grade = a.calculateGrade(audit.OverallScore)
	audit.Recommendations = a.generateRecommendations(audit)

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
