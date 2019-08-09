package global

import "github.com/eluv-io/inject-go/example/hierarchical"

func newStore() hierarchical.Store {
	return &FileStore{}
}

type FileStore struct {
}

func (s *FileStore) StoreTransanction(tx string) {
	panic("implement me")
}
