package fs

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type FileSystem struct {
	*os.File
	contentType string
	size        int64
}

func (fs *FileSystem) ContentType() string {
	return fs.contentType
}

func (fs *FileSystem) Size() int64 {
	return fs.size
}

func (fs *FileSystem) Path() string {
	return fs.File.Name()
}

func (fs *FileSystem) Destroy() error {
	_ = fs.File.Close()
	return os.Remove(fs.File.Name())
}

func New(path string) (*FileSystem, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	fileStat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	_, _ = file.Seek(0, io.SeekStart)

	return &FileSystem{
		File:        file,
		size:        fileStat.Size(),
		contentType: contentType,
	}, nil
}
