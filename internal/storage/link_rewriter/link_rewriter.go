package link_rewriter

import (
	"path/filepath"
	"regexp"
	"strings"

	"wget-go/internal/storage"
)

type LinkRewriterImpl struct {
	pathResolver storage.PathResolver
}

func New(pathResolver storage.PathResolver) *LinkRewriterImpl {
	return &LinkRewriterImpl{pathResolver: pathResolver}
}

func (lr *LinkRewriterImpl) RewriteHTML(htmlContent []byte, baseURL string) ([]byte, error) {
	content := string(htmlContent)

	// Обрабатываем href атрибуты
	hrefRegex := regexp.MustCompile(`href=["']([^"']+)["']`)
	content = hrefRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := hrefRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		originalURL := parts[1]
		if lr.shouldRewrite(originalURL) {
			localPath, err := lr.urlToRelativePath(originalURL, baseURL)
			if err == nil {
				return `href="` + localPath + `"`
			}
		}

		return match
	})

	// Обрабатываем src атрибуты
	srcRegex := regexp.MustCompile(`src=["']([^"']+)["']`)
	content = srcRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := srcRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		originalURL := parts[1]
		if lr.shouldRewrite(originalURL) {
			localPath, err := lr.urlToRelativePath(originalURL, baseURL)
			if err == nil {
				return `src="` + localPath + `"`
			}
		}

		return match
	})

	return []byte(content), nil
}

func (lr *LinkRewriterImpl) RewriteCSS(cssContent []byte, baseURL string) ([]byte, error) {
	// Пока просто возвращаем оригинальный контент
	// Можно добавить обработку позже
	return cssContent, nil
}

func (lr *LinkRewriterImpl) urlToRelativePath(originalURL, baseURL string) (string, error) {
	// Преобразуем относительный URL в абсолютный
	absoluteURL, err := lr.pathResolver.ResolveAbsoluteURL(baseURL, originalURL)
	if err != nil {
		return "", err
	}

	// Получаем локальный путь для целевого URL
	targetPath, err := lr.pathResolver.URLToLocalPath(absoluteURL)
	if err != nil {
		return "", err
	}

	// Получаем локальный путь для базового URL
	basePath, err := lr.pathResolver.URLToLocalPath(baseURL)
	if err != nil {
		return "", err
	}

	// Вычисляем относительный путь
	relativePath, err := filepath.Rel(filepath.Dir(basePath), targetPath)
	if err != nil {
		return "", err
	}

	// Для веб-совместимости используем прямые слеши
	return filepath.ToSlash(relativePath), nil
}

func (lr *LinkRewriterImpl) shouldRewrite(urlStr string) bool {
	// Не перезаписываем абсолютные URL, data URI, якоря и т.д.
	if strings.HasPrefix(urlStr, "http://") ||
		strings.HasPrefix(urlStr, "https://") ||
		strings.HasPrefix(urlStr, "data:") ||
		strings.HasPrefix(urlStr, "#") ||
		strings.HasPrefix(urlStr, "javascript:") ||
		strings.HasPrefix(urlStr, "mailto:") {
		return false
	}
	return true
}
