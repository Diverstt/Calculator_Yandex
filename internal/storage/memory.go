// internal/storage/memory.go
package storage

import (
	"sync"

	"github.com/Diverstt/Calculator_Yandex/internal/models"
)

// MemoryStore реализует простое in-memory хранилище для выражений.
type MemoryStore struct {
	Expressions map[string]*models.Expression
	Mutex       sync.Mutex
}

// NewMemoryStore возвращает новый экземпляр хранилища.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Expressions: make(map[string]*models.Expression),
	}
}

// SaveExpression сохраняет выражение в хранилище.
func (s *MemoryStore) SaveExpression(id string, expr *models.Expression) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Expressions[id] = expr
}

// GetExpression возвращает выражение по его ID.
func (s *MemoryStore) GetExpression(id string) (*models.Expression, bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	expr, exists := s.Expressions[id]
	return expr, exists
}
