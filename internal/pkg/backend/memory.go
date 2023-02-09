package backend

import (
	"context"
	"errors"
	"time"
)

type MemoryBackend struct {
	Config   map[string]map[int][]string
	Metadata map[string]Metadata
}

func (b *MemoryBackend) Get(ctx context.Context, path string, version int) (string, error) {
	res, err := b.GetMany(ctx, path, version)
	if err != nil {
		return "", nil
	}
	if len(res) < 1 {
		return "", &NotFoundErr{path}
	}
	return res[0], nil
}

func (b *MemoryBackend) GetCurrent(ctx context.Context, path string) (string, error) {
	res, err := b.GetManyCurrent(ctx, path)
	if err != nil {
		return "", nil
	}
	if len(res) < 1 {
		return "", &NotFoundErr{path}
	}
	return res[0], nil
}

func (b *MemoryBackend) GetManyCurrent(ctx context.Context, path string) ([]string, error) {
	metadata, err := b.GetMetadata(ctx, path)
	if err != nil {
		return nil, err
	}
	return b.Config[metadata.Path][metadata.CurrentVersion], nil
}

func (b *MemoryBackend) GetMany(ctx context.Context, path string, version int) ([]string, error) {
	pathedConfig, ok := b.Config[path]
	if !ok {
		return nil, &NotFoundErr{path}
	}
	versionedConfig, ok := pathedConfig[version]
	if !ok {
		return nil, &NotFoundErr{path}
	}
	return versionedConfig, nil
}

func (b *MemoryBackend) Set(ctx context.Context, path string, value string, options SetOptions) (*Metadata, error) {
	return b.SetMany(ctx, path, []string{value}, options)
}

func (b *MemoryBackend) SetMany(ctx context.Context, path string, values []string, options SetOptions) (*Metadata, error) {
	if b.Config[path] == nil {
		b.Config[path] = make(map[int][]string)
	}
	var notFoundErr *NotFoundErr
	metadata, err := b.GetMetadata(ctx, path)
	if err != nil && errors.As(err, &notFoundErr) {
		metadata = &Metadata{
			Path:      path,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		b.Metadata[path] = *metadata
	} else if err != nil {
		return nil, err
	}
	b.Config[path][metadata.LatestVersion+1] = values
	return metadata, nil
}

func (b *MemoryBackend) GetMetadata(ctx context.Context, path string) (*Metadata, error) {
	metadata, ok := b.Metadata[path]
	if !ok {
		return nil, &NotFoundErr{path}
	}
	return &metadata, nil
}

func (b *MemoryBackend) Close(ctx context.Context) error {
	return nil
}
