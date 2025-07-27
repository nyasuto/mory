package memory

import "time"

// Memory represents a stored memory item
type Memory struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoryStore defines the interface for memory storage operations
type MemoryStore interface {
	Save(memory *Memory) error
	Get(key string) (*Memory, error)
	List(category string) ([]*Memory, error)
	Delete(key string) error
}