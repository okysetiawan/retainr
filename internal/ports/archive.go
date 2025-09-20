package ports

import (
	"context"
)

type ArchiveSource interface {
	Subscribe(topic string, callback func(msg []byte)) error
}

type ArchiveDestination interface {
	PutReader(ctx context.Context, key string, reader Reader) error
}
