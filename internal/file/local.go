package file

import (
	"os"
	"path/filepath"
)

type LocalFileProvider struct{}

func NewLocalFileProvider() *LocalFileProvider { return &LocalFileProvider{} }

func (lf *LocalFileProvider) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0777)
}

func (lf *LocalFileProvider) CreateFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0777)
}

func (lf *LocalFileProvider) CreateFiles(dir string, files map[string]string) error {
	for name, content := range files {
		fullPath := filepath.Join(dir, name)
		if err := lf.CreateFile(fullPath, content); err != nil {
			return err
		}
	}
	return nil
}

func (lf *LocalFileProvider) DeleteDirectory(path string) error {
	return os.RemoveAll(path)
}