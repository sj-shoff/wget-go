package html_parser

import (
	"strings"

	"golang.org/x/net/html"
)

// Parser реализует парсинг HTML
type Parser struct{}

// New создает новый экземпляр Parser
func New() *Parser {
	return &Parser{}
}

// Parse парсит HTML и возвращает ссылки
func (p *Parser) Parse(htmlContent []byte, baseURL string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(string(htmlContent)))
	if err != nil {
		return nil, err
	}

	var links []string
	var extract func(*html.Node)

	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a", "link":
				if href := getAttribute(n, "href"); href != "" {
					links = append(links, href)
				}
			case "img", "script", "iframe", "embed":
				if src := getAttribute(n, "src"); src != "" {
					links = append(links, src)
				}
			case "meta":
				if getAttribute(n, "property") == "og:image" ||
					getAttribute(n, "name") == "twitter:image" {
					if content := getAttribute(n, "content"); content != "" {
						links = append(links, content)
					}
				}
			case "object":
				if data := getAttribute(n, "data"); data != "" {
					links = append(links, data)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)
	return p.filterLinks(links), nil
}

// filterLinks фильтрует ссылки
func (p *Parser) filterLinks(links []string) []string {
	var filtered []string
	seen := make(map[string]bool)

	for _, link := range links {
		// Пропускаем пустые ссылки, якоря и javascript
		if link == "" || strings.HasPrefix(link, "#") ||
			strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") {
			continue
		}

		// Убираем дубликаты
		if !seen[link] {
			seen[link] = true
			filtered = append(filtered, link)
		}
	}

	return filtered
}

// getAttribute возвращает значение атрибута по ключу
func getAttribute(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
