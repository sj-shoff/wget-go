package domain

// ResourceType определяет тип скачиваемого ресурса
type ResourceType int

const (
	ResourceHTML ResourceType = iota
	ResourceCSS
	ResourceJavaScript
	ResourceImage
	ResourceFont
	ResourceOther
)

// DownloadTask представляет задачу на скачивание
type DownloadTask struct {
	URL       string
	Depth     int
	Type      ResourceType
	ParentURL string
}

// DownloadResult представляет результат скачивания
type DownloadResult struct {
	Task     DownloadTask
	Content  []byte
	Links    []string
	FilePath string
	Error    error
}

// String реализует интерфейс fmt.Stringer для ResourceType
func (rt ResourceType) String() string {
	switch rt {
	case ResourceHTML:
		return "HTML"
	case ResourceCSS:
		return "CSS"
	case ResourceJavaScript:
		return "JavaScript"
	case ResourceImage:
		return "Image"
	case ResourceFont:
		return "Font"
	default:
		return "Other"
	}
}
