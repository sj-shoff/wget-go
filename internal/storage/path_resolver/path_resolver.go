package path_resolver

import (
	"net/url"
	"path/filepath"
	"strings"
)

type PathResolverImpl struct {
	baseDir string
}

func New(baseDir string) *PathResolverImpl {
	return &PathResolverImpl{baseDir: baseDir}
}

func (pr *PathResolverImpl) URLToLocalPath(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Получаем путь из URL
	urlPath := parsed.Path
	if urlPath == "" {
		urlPath = "/"
	}

	// Создаем базовый путь: basedir/host/path
	basePath := filepath.Join(pr.baseDir, parsed.Host)

	// Обрабатываем путь для создания правильной структуры файлов
	localPath := pr.buildLocalPath(basePath, urlPath)

	return localPath, nil
}

func (pr *PathResolverImpl) buildLocalPath(basePath, urlPath string) string {
	// Если путь заканчивается на /, это директория - создаем index.html
	if strings.HasSuffix(urlPath, "/") {
		return filepath.Join(basePath, urlPath, "index.html")
	}

	// Если в пути нет точки (скорее всего HTML страница без расширения)
	if !strings.Contains(filepath.Base(urlPath), ".") {
		// Проверяем, не является ли это корневым URL
		if urlPath == "/" {
			return filepath.Join(basePath, "index.html")
		}
		return filepath.Join(basePath, urlPath) + ".html"
	}

	// Обычный файл с расширением
	return filepath.Join(basePath, urlPath)
}

func (pr *PathResolverImpl) ResolveAbsoluteURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	relative, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(relative).String(), nil
}

func (pr *PathResolverImpl) IsSameDomain(url1, url2 string) bool {
	u1, err := url.Parse(url1)
	if err != nil {
		return false
	}

	u2, err := url.Parse(url2)
	if err != nil {
		return false
	}

	return u1.Host == u2.Host
}
