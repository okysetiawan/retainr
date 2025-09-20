package ports

import "io"

type Reader interface {
	io.Reader
	Size() int64
	ContentType() string
}

type ReadCloser interface {
	io.Closer
	Reader
}
