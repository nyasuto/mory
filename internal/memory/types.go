package memory

import (
	"fmt"
	"time"
)

// Memory represents a stored memory item
type Memory struct {
	ID        string    `json:"id"` // 自動生成: memory_20250127123456
	Category  string    `json:"category"`
	Key       string    `json:"key"` // オプション: ユーザーフレンドリーなエイリアス
	Value     string    `json:"value"`
	Tags      []string  `json:"tags"` // 関連タグ（将来的な検索用）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OperationLog represents a log entry for memory operations
type OperationLog struct {
	Timestamp   time.Time `json:"timestamp"`
	OperationID string    `json:"operation_id"`
	Operation   string    `json:"operation"`
	Key         string    `json:"key,omitempty"`
	Before      *Memory   `json:"before,omitempty"`
	After       *Memory   `json:"after,omitempty"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
}

// MemoryStore defines the interface for memory storage operations
type MemoryStore interface {
	Save(memory *Memory) (string, error) // IDを返すように変更
	Get(key string) (*Memory, error)
	GetByID(id string) (*Memory, error) // ID検索メソッド追加
	List(category string) ([]*Memory, error)
	Search(query SearchQuery) ([]*SearchResult, error) // 検索メソッド追加
	Delete(key string) error
	DeleteByID(id string) error           // ID削除メソッド追加
	LogOperation(log *OperationLog) error // 操作ログ記録メソッド追加
}

// GenerateID generates a timestamp-based unique ID
func GenerateID() string {
	return fmt.Sprintf("memory_%d", time.Now().UnixNano())
}

// GenerateOperationID generates a unique operation ID
func GenerateOperationID() string {
	return fmt.Sprintf("op_%d", time.Now().UnixNano())
}

// MemoryNotFoundError represents an error when a memory is not found
type MemoryNotFoundError struct {
	Key string
}

func (e *MemoryNotFoundError) Error() string {
	return fmt.Sprintf("memory not found: %s", e.Key)
}
