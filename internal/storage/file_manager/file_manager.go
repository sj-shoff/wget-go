package file_manager

import (
	"os"
	"path/filepath"
)

type FileManagerImpl struct {
	// УБИРАЕМ baseDir - пути уже абсолютные
}

func New() *FileManagerImpl {
	return &FileManagerImpl{}
}

func (fm *FileManagerImpl) Save(filePath string, content []byte) error {
	// filePath уже содержит полный путь от PathResolver
	dir := filepath.Dir(filePath)

	// Создаем директории если нужно
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Просто записываем файл
	return os.WriteFile(filePath, content, 0644)
}

func (fm *FileManagerImpl) Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
