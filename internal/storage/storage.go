package storage

// FileManager управляет файловой системой
type FileManager interface {
	Save(filePath string, content []byte) error
	Exists(filePath string) bool
	CreateDir(path string) error
}

// LinkRewriter перезаписывает ссылки в контенте
type LinkRewriter interface {
	RewriteHTML(htmlContent []byte, baseURL string) ([]byte, error)
	RewriteCSS(cssContent []byte, baseURL string) ([]byte, error)
}

// PathResolver преобразует URL в локальные пути
type PathResolver interface {
	URLToLocalPath(url string) (string, error)
	ResolveAbsoluteURL(baseURL, relativeURL string) (string, error)
	IsSameDomain(url1, url2 string) bool
}
