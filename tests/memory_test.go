package tests

import (
	"testing"

	"github.com/Diverstt/Calculator_Yandex/internal/models"
	"github.com/Diverstt/Calculator_Yandex/internal/storage"
)

func TestMemoryStore(t *testing.T) {
	store := storage.NewMemoryStore()
	expr := &models.Expression{
		ID:     "expr1",
		Status: "pending",
	}
	store.SaveExpression(expr.ID, expr)

	retrieved, exists := store.GetExpression(expr.ID)
	if !exists {
		t.Fatalf("Выражение с ID %s не найдено в хранилище", expr.ID)
	}
	if retrieved.ID != expr.ID || retrieved.Status != expr.Status {
		t.Errorf("Сохранённое и полученное выражения не совпадают")
	}
}
