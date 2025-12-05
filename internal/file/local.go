package file

import "os"

type LocalFileProvider struct{}

func NewLocalFileProvider() *LocalFileProvider { return &LocalFileProvider{} }
func (lf *LocalFileProvider) CreateDirectory(path string) error { return os.MkdirAll(path, 0777) }
func (lf *LocalFileProvider) CreateFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0777)
}
func (lf *LocalFileProvider) DeleteDirectory(path string) error { return os.RemoveAll(path) }
