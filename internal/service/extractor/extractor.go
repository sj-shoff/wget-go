package extractor

import (
	"regexp"
	"strings"
	"wget-go/internal/service"
)

// LinkExtractor извлекает ссылки из различного контента
type LinkExtractor struct {
	htmlParser  service.HTMLParser
	cssURLRegex *regexp.Regexp
}

// NewLinkExtractor создает новый экземпляр LinkExtractor
func NewLinkExtractor(htmlParser service.HTMLParser) *LinkExtractor {
	return &LinkExtractor{
		htmlParser:  htmlParser,
		cssURLRegex: regexp.MustCompile(`url\(['"]?([^'")]+)['"]?\)`),
	}
}

// ExtractLinks извлекает ссылки из контента
func (e *LinkExtractor) ExtractLinks(content []byte, baseURL string, contentType string) ([]string, error) {
	if strings.Contains(contentType, "text/html") {
		return e.htmlParser.Parse(content, baseURL)
	}

	if strings.Contains(contentType, "text/css") {
		return e.extractFromCSS(content), nil
	}

	return nil, nil
}

// extractFromCSS извлекает ссылки из CSS
func (e *LinkExtractor) extractFromCSS(cssContent []byte) []string {
	var links []string
	content := string(cssContent)

	// ищем url в css
	matches := e.cssURLRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			links = append(links, match[1])
		}
	}

	// ищем @import в css
	importRegex := regexp.MustCompile(`@import\s+["']([^"']+)["']`)
	importMatches := importRegex.FindAllStringSubmatch(content, -1)
	for _, match := range importMatches {
		if len(match) > 1 && match[1] != "" {
			links = append(links, match[1])
		}
	}

	return links
}
