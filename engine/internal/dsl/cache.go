package dsl

import (
	"github.com/timickb/sagaflow/engine/internal/domain"
)

type Cache struct {
	sagas map[domain.SagaHeader]*domain.SagaView
}

func NewCache(dirPath string) (*Cache, error) {
	// todo: implement
	panic("implement me")
}

func (r *Cache) GetSagaView(header domain.SagaHeader) (*domain.SagaView, bool) {
	view, ok := r.sagas[header]
	return view, ok
}
