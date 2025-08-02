package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteMemoryStore implements MemoryStore interface using SQLite database
type SQLiteMemoryStore struct {
	db             *sql.DB
	semanticEngine SemanticSearchEngine

	// Prepared statements for performance
	stmtInsert      *sql.Stmt
	stmtUpdate      *sql.Stmt
	stmtGetByKey    *sql.Stmt
	stmtGetByID     *sql.Stmt
	stmtList        *sql.Stmt
	stmtListByCat   *sql.Stmt
	stmtDeleteByKey *sql.Stmt
	stmtDeleteByID  *sql.Stmt
	stmtLogOp       *sql.Stmt
}

// NewSQLiteMemoryStore creates a new SQLite-based memory store
func NewSQLiteMemoryStore(dbPath string) (*SQLiteMemoryStore, error) {
	// Open SQLite database with appropriate settings
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=memory")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	store := &SQLiteMemoryStore{db: db}

	// Initialize schema
	if err := store.initializeSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Prepare statements
	if err := store.prepareStatements(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to prepare statements: %w", err)
	}

	// Initialize FTS5 if available (optional)
	if err := store.initializeFTS5(); err != nil {
		log.Printf("[SQLiteMemoryStore] Warning: FTS5 not available, using LIKE search fallback: %v", err)
	}

	log.Printf("[SQLiteMemoryStore] Successfully initialized at %s", dbPath)
	return store, nil
}

// Close closes the database connection and prepared statements
func (s *SQLiteMemoryStore) Close() error {
	if s.stmtInsert != nil {
		_ = s.stmtInsert.Close()
	}
	if s.stmtUpdate != nil {
		_ = s.stmtUpdate.Close()
	}
	if s.stmtGetByKey != nil {
		_ = s.stmtGetByKey.Close()
	}
	if s.stmtGetByID != nil {
		_ = s.stmtGetByID.Close()
	}
	if s.stmtList != nil {
		_ = s.stmtList.Close()
	}
	if s.stmtListByCat != nil {
		_ = s.stmtListByCat.Close()
	}
	if s.stmtDeleteByKey != nil {
		_ = s.stmtDeleteByKey.Close()
	}
	if s.stmtDeleteByID != nil {
		_ = s.stmtDeleteByID.Close()
	}
	if s.stmtLogOp != nil {
		_ = s.stmtLogOp.Close()
	}

	return s.db.Close()
}

// initializeSchema creates tables and indexes from schema.sql
func (s *SQLiteMemoryStore) initializeSchema() error {
	schema := `
-- メイン記憶テーブル
CREATE TABLE IF NOT EXISTS memories (
    id TEXT PRIMARY KEY,
    category TEXT NOT NULL,
    key TEXT,
    value TEXT NOT NULL,
    tags TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    embedding BLOB,
    embedding_hash TEXT
);

-- パフォーマンス最適化用インデックス
CREATE INDEX IF NOT EXISTS idx_memories_category ON memories(category);
CREATE INDEX IF NOT EXISTS idx_memories_key ON memories(key) WHERE key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_memories_created_at ON memories(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memories_updated_at ON memories(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_memories_embedding_hash ON memories(embedding_hash) WHERE embedding_hash IS NOT NULL;

-- Note: FTS5 setup is optional and will be created separately if available

-- 操作ログテーブル
CREATE TABLE IF NOT EXISTS operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    operation_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    memory_key TEXT,
    memory_id TEXT,
    before_data TEXT,
    after_data TEXT,
    success BOOLEAN NOT NULL,
    error TEXT
);

-- 操作ログのインデックス
CREATE INDEX IF NOT EXISTS idx_operation_logs_timestamp ON operation_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_operation ON operation_logs(operation);
CREATE INDEX IF NOT EXISTS idx_operation_logs_memory_id ON operation_logs(memory_id);

-- データベース設定
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;
PRAGMA temp_store = memory;
PRAGMA mmap_size = 268435456;
`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Printf("[SQLiteMemoryStore] Schema initialized successfully")
	return nil
}

// initializeFTS5 initializes FTS5 full-text search if available
func (s *SQLiteMemoryStore) initializeFTS5() error {
	ftsSchema := `
-- 全文検索用仮想テーブル（FTS5）
CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
    id UNINDEXED,
    category,
    key,
    value,
    tags,
    content='memories',
    content_rowid='rowid',
    tokenize='unicode61 remove_diacritics 2'
);

-- FTS5自動更新用トリガー
CREATE TRIGGER IF NOT EXISTS memories_fts_insert AFTER INSERT ON memories BEGIN
    INSERT INTO memories_fts(id, category, key, value, tags)
    VALUES (new.id, new.category, new.key, new.value, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS memories_fts_update AFTER UPDATE ON memories BEGIN
    UPDATE memories_fts SET
        category = new.category,
        key = new.key,
        value = new.value,
        tags = new.tags
    WHERE id = new.id;
END;

CREATE TRIGGER IF NOT EXISTS memories_fts_delete AFTER DELETE ON memories BEGIN
    DELETE FROM memories_fts WHERE id = old.id;
END;
`

	_, err := s.db.Exec(ftsSchema)
	if err != nil {
		return fmt.Errorf("failed to initialize FTS5: %w", err)
	}

	log.Printf("[SQLiteMemoryStore] FTS5 initialized successfully")
	return nil
}

// prepareStatements prepares SQL statements for better performance
func (s *SQLiteMemoryStore) prepareStatements() error {
	var err error

	// Insert statement
	s.stmtInsert, err = s.db.Prepare(`
		INSERT INTO memories (id, category, key, value, tags, created_at, updated_at, embedding, embedding_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	// Update statement
	s.stmtUpdate, err = s.db.Prepare(`
		UPDATE memories 
		SET category = ?, key = ?, value = ?, tags = ?, updated_at = ?, embedding = ?, embedding_hash = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}

	// Get by key statement
	s.stmtGetByKey, err = s.db.Prepare(`
		SELECT id, category, key, value, tags, created_at, updated_at, embedding, embedding_hash
		FROM memories WHERE key = ? LIMIT 1
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare get by key statement: %w", err)
	}

	// Get by ID statement
	s.stmtGetByID, err = s.db.Prepare(`
		SELECT id, category, key, value, tags, created_at, updated_at, embedding, embedding_hash
		FROM memories WHERE id = ? LIMIT 1
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare get by ID statement: %w", err)
	}

	// List all memories statement
	s.stmtList, err = s.db.Prepare(`
		SELECT id, category, key, value, tags, created_at, updated_at, embedding, embedding_hash
		FROM memories ORDER BY created_at DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare list statement: %w", err)
	}

	// List by category statement
	s.stmtListByCat, err = s.db.Prepare(`
		SELECT id, category, key, value, tags, created_at, updated_at, embedding, embedding_hash
		FROM memories WHERE category = ? ORDER BY created_at DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare list by category statement: %w", err)
	}

	// Delete by key statement
	s.stmtDeleteByKey, err = s.db.Prepare(`
		DELETE FROM memories WHERE key = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare delete by key statement: %w", err)
	}

	// Delete by ID statement
	s.stmtDeleteByID, err = s.db.Prepare(`
		DELETE FROM memories WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare delete by ID statement: %w", err)
	}

	// Log operation statement
	s.stmtLogOp, err = s.db.Prepare(`
		INSERT INTO operation_logs (timestamp, operation_id, operation, memory_key, memory_id, before_data, after_data, success, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare log operation statement: %w", err)
	}

	log.Printf("[SQLiteMemoryStore] Prepared statements initialized successfully")
	return nil
}

// Save stores a memory item and returns the generated ID
func (s *SQLiteMemoryStore) Save(memory *Memory) (string, error) {
	// Generate ID if not provided
	if memory.ID == "" {
		memory.ID = GenerateID()
	}

	// Set timestamps
	now := time.Now()
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = now
	}
	memory.UpdatedAt = now

	// Initialize Tags if nil
	if memory.Tags == nil {
		memory.Tags = []string{}
	}

	// Serialize tags to JSON
	tagsJSON, err := json.Marshal(memory.Tags)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tags: %w", err)
	}

	// Serialize embedding to binary format
	var embeddingBlob []byte
	if len(memory.Embedding) > 0 {
		embeddingBlob, err = serializeEmbedding(memory.Embedding)
		if err != nil {
			return "", fmt.Errorf("failed to serialize embedding: %w", err)
		}
	}

	// Check if memory exists (by key or ID)
	var existingMemory *Memory
	if memory.Key != "" {
		if existing, err := s.Get(memory.Key); err == nil {
			existingMemory = existing
			memory.ID = existing.ID               // Preserve existing ID
			memory.CreatedAt = existing.CreatedAt // Preserve creation time
		}
	} else if existing, err := s.GetByID(memory.ID); err == nil {
		existingMemory = existing
		memory.CreatedAt = existing.CreatedAt // Preserve creation time
	}

	// Convert timestamps to Unix nanoseconds
	createdAtUnix := memory.CreatedAt.UnixNano()
	updatedAtUnix := memory.UpdatedAt.UnixNano()

	var operation string
	if existingMemory != nil {
		// Update existing memory
		operation = "update"
		_, err = s.stmtUpdate.Exec(
			memory.Category,
			memory.Key,
			memory.Value,
			string(tagsJSON),
			updatedAtUnix,
			embeddingBlob,
			memory.EmbeddingHash,
			memory.ID,
		)
	} else {
		// Insert new memory
		operation = "save"
		_, err = s.stmtInsert.Exec(
			memory.ID,
			memory.Category,
			memory.Key,
			memory.Value,
			string(tagsJSON),
			createdAtUnix,
			updatedAtUnix,
			embeddingBlob,
			memory.EmbeddingHash,
		)
	}

	if err != nil {
		return "", fmt.Errorf("failed to %s memory: %w", operation, err)
	}

	// Log the operation
	if logErr := s.logOperation(operation, memory.Key, memory.ID, existingMemory, memory, true, ""); logErr != nil {
		log.Printf("[SQLiteMemoryStore] Warning: failed to log operation: %v", logErr)
	}

	log.Printf("[SQLiteMemoryStore] Successfully %sd memory: %s", operation, memory.ID)
	return memory.ID, nil
}

// serializeEmbedding converts float32 slice to binary format
func serializeEmbedding(embedding []float32) ([]byte, error) {
	data, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding: %w", err)
	}
	return data, nil
}

// deserializeEmbedding converts binary format back to float32 slice
func deserializeEmbedding(data []byte) ([]float32, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var embedding []float32
	if err := json.Unmarshal(data, &embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}
	return embedding, nil
}

// Get retrieves a memory by key
func (s *SQLiteMemoryStore) Get(key string) (*Memory, error) {
	row := s.stmtGetByKey.QueryRow(key)
	return s.scanMemory(row)
}

// GetByID retrieves a memory by ID
func (s *SQLiteMemoryStore) GetByID(id string) (*Memory, error) {
	row := s.stmtGetByID.QueryRow(id)
	return s.scanMemory(row)
}

// List retrieves all memories, optionally filtered by category
func (s *SQLiteMemoryStore) List(category string) ([]*Memory, error) {
	var rows *sql.Rows
	var err error

	if category == "" {
		rows, err = s.stmtList.Query()
	} else {
		rows, err = s.stmtListByCat.Query(category)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var memories []*Memory
	for rows.Next() {
		memory, err := s.scanMemory(rows)
		if err != nil {
			log.Printf("[SQLiteMemoryStore] Warning: failed to scan memory: %v", err)
			continue
		}
		memories = append(memories, memory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return memories, nil
}

// Search performs search across memories using semantic search if available
func (s *SQLiteMemoryStore) Search(query SearchQuery) ([]*SearchResult, error) {
	// Use semantic search if available and enabled
	if s.semanticEngine != nil {
		return s.semanticEngine.Search(query)
	}

	// Fall back to FTS5 search or keyword search
	return s.performFTSSearch(query)
}

// performFTSSearch performs full-text search using SQLite FTS5
func (s *SQLiteMemoryStore) performFTSSearch(query SearchQuery) ([]*SearchResult, error) {
	if query.Query == "" {
		// Return all memories with equal score
		memories, err := s.List(query.Category)
		if err != nil {
			return nil, err
		}

		results := make([]*SearchResult, len(memories))
		for i, memory := range memories {
			results[i] = &SearchResult{
				Memory: memory,
				Score:  1.0,
			}
		}
		return results, nil
	}

	// Construct FTS5 query
	ftsQuery := query.Query
	if query.Category != "" {
		ftsQuery = fmt.Sprintf("category:%s AND (%s)", query.Category, query.Query)
	}

	// Execute FTS5 search
	sqlQuery := `
		SELECT m.id, m.category, m.key, m.value, m.tags, m.created_at, m.updated_at, m.embedding, m.embedding_hash,
		       fts.rank
		FROM memories_fts fts
		JOIN memories m ON m.id = fts.id
		WHERE memories_fts MATCH ?
		ORDER BY fts.rank
		LIMIT 50
	`

	rows, err := s.db.Query(sqlQuery, ftsQuery)
	if err != nil {
		// Fall back to simple LIKE search if FTS5 fails
		log.Printf("[SQLiteMemoryStore] FTS5 search failed, falling back to LIKE search: %v", err)
		return s.performLikeSearch(query)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var results []*SearchResult
	for rows.Next() {
		var memory Memory
		var rank float64
		var tagsJSON string
		var embeddingBlob []byte
		var createdAtUnix, updatedAtUnix int64

		err := rows.Scan(
			&memory.ID, &memory.Category, &memory.Key, &memory.Value, &tagsJSON,
			&createdAtUnix, &updatedAtUnix, &embeddingBlob, &memory.EmbeddingHash, &rank,
		)
		if err != nil {
			log.Printf("[SQLiteMemoryStore] Warning: failed to scan FTS search result: %v", err)
			continue
		}

		// Deserialize data
		if err := json.Unmarshal([]byte(tagsJSON), &memory.Tags); err != nil {
			log.Printf("[SQLiteMemoryStore] Warning: failed to unmarshal tags: %v", err)
			memory.Tags = []string{}
		}

		memory.Embedding, err = deserializeEmbedding(embeddingBlob)
		if err != nil {
			log.Printf("[SQLiteMemoryStore] Warning: failed to deserialize embedding: %v", err)
		}

		memory.CreatedAt = time.Unix(0, createdAtUnix)
		memory.UpdatedAt = time.Unix(0, updatedAtUnix)

		// Convert FTS5 rank to relevance score (0.0 - 1.0)
		score := 1.0 / (1.0 + rank) // Higher rank = lower score, so invert
		if score > 1.0 {
			score = 1.0
		}

		results = append(results, &SearchResult{
			Memory: &memory,
			Score:  score,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("FTS search row iteration error: %w", err)
	}

	return results, nil
}

// performLikeSearch performs simple LIKE-based search as fallback
func (s *SQLiteMemoryStore) performLikeSearch(query SearchQuery) ([]*SearchResult, error) {
	queryPattern := "%" + query.Query + "%"

	sqlQuery := `
		SELECT id, category, key, value, tags, created_at, updated_at, embedding, embedding_hash
		FROM memories 
		WHERE (value LIKE ? OR key LIKE ? OR category LIKE ?)
	`
	args := []interface{}{queryPattern, queryPattern, queryPattern}

	if query.Category != "" {
		sqlQuery += " AND category = ?"
		args = append(args, query.Category)
	}

	sqlQuery += " ORDER BY updated_at DESC LIMIT 50"

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute LIKE search: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var results []*SearchResult
	for rows.Next() {
		memory, err := s.scanMemory(rows)
		if err != nil {
			log.Printf("[SQLiteMemoryStore] Warning: failed to scan LIKE search result: %v", err)
			continue
		}

		// Calculate simple relevance score based on matches
		score := s.calculateLikeScore(memory, query.Query)

		results = append(results, &SearchResult{
			Memory: memory,
			Score:  score,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("LIKE search row iteration error: %w", err)
	}

	return results, nil
}

// calculateLikeScore calculates relevance score for LIKE search
func (s *SQLiteMemoryStore) calculateLikeScore(memory *Memory, query string) float64 {
	score := 0.0
	queryLower := query

	// Check exact matches first
	if memory.Key == query {
		score += 1.0
	} else if memory.Value == query {
		score += 0.9
	} else if memory.Category == query {
		score += 0.7
	}

	// Check partial matches
	if score == 0 {
		if memory.Key != "" && memory.Key == queryLower {
			score += 0.8
		}
		if memory.Value != "" && memory.Value == queryLower {
			score += 0.6
		}
		if memory.Category != "" && memory.Category == queryLower {
			score += 0.5
		}
	}

	// Normalize
	if score > 1.0 {
		score = 1.0
	}
	if score == 0 {
		score = 0.1 // Minimum score for matches
	}

	return score
}

// Delete removes a memory by key
func (s *SQLiteMemoryStore) Delete(key string) error {
	// Get existing memory for logging
	existingMemory, err := s.Get(key)
	if err != nil {
		return fmt.Errorf("memory with key '%s' not found", key)
	}

	// Execute deletion
	result, err := s.stmtDeleteByKey.Exec(key)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("memory with key '%s' not found", key)
	}

	// Log the operation
	if logErr := s.logOperation("delete", key, existingMemory.ID, existingMemory, nil, true, ""); logErr != nil {
		log.Printf("[SQLiteMemoryStore] Warning: failed to log deletion operation: %v", logErr)
	}

	log.Printf("[SQLiteMemoryStore] Successfully deleted memory with key: %s", key)
	return nil
}

// DeleteByID removes a memory by ID
func (s *SQLiteMemoryStore) DeleteByID(id string) error {
	// Get existing memory for logging
	existingMemory, err := s.GetByID(id)
	if err != nil {
		return fmt.Errorf("memory with ID '%s' not found", id)
	}

	// Execute deletion
	result, err := s.stmtDeleteByID.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("memory with ID '%s' not found", id)
	}

	// Log the operation
	if logErr := s.logOperation("delete", existingMemory.Key, id, existingMemory, nil, true, ""); logErr != nil {
		log.Printf("[SQLiteMemoryStore] Warning: failed to log deletion operation: %v", logErr)
	}

	log.Printf("[SQLiteMemoryStore] Successfully deleted memory with ID: %s", id)
	return nil
}

// LogOperation records an operation log
func (s *SQLiteMemoryStore) LogOperation(log *OperationLog) error {
	return s.logOperation(log.Operation, log.Key, "", log.Before, log.After, log.Success, log.Error)
}

// logOperation internal helper for logging operations
func (s *SQLiteMemoryStore) logOperation(operation, key, memoryID string, before, after *Memory, success bool, errorMsg string) error {
	now := time.Now()
	operationID := GenerateOperationID()

	var beforeData, afterData *string

	if before != nil {
		data, err := json.Marshal(before)
		if err != nil {
			log.Printf("[SQLiteMemoryStore] Warning: failed to marshal before data: %v", err)
		} else {
			dataStr := string(data)
			beforeData = &dataStr
		}
	}

	if after != nil {
		data, err := json.Marshal(after)
		if err != nil {
			log.Printf("[SQLiteMemoryStore] Warning: failed to marshal after data: %v", err)
		} else {
			dataStr := string(data)
			afterData = &dataStr
		}
	}

	// If memoryID is empty, try to get it from after or before
	if memoryID == "" {
		if after != nil {
			memoryID = after.ID
		} else if before != nil {
			memoryID = before.ID
		}
	}

	_, err := s.stmtLogOp.Exec(
		now.UnixNano(),
		operationID,
		operation,
		key,
		memoryID,
		beforeData,
		afterData,
		success,
		errorMsg,
	)

	if err != nil {
		return fmt.Errorf("failed to log operation: %w", err)
	}

	return nil
}

// scanMemory scans a database row into a Memory struct
func (s *SQLiteMemoryStore) scanMemory(scanner interface{}) (*Memory, error) {
	var memory Memory
	var tagsJSON string
	var embeddingBlob []byte
	var createdAtUnix, updatedAtUnix int64

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&memory.ID, &memory.Category, &memory.Key, &memory.Value, &tagsJSON,
			&createdAtUnix, &updatedAtUnix, &embeddingBlob, &memory.EmbeddingHash,
		)
	case *sql.Rows:
		err = s.Scan(
			&memory.ID, &memory.Category, &memory.Key, &memory.Value, &tagsJSON,
			&createdAtUnix, &updatedAtUnix, &embeddingBlob, &memory.EmbeddingHash,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("memory not found")
		}
		return nil, fmt.Errorf("failed to scan memory: %w", err)
	}

	// Deserialize tags
	if err := json.Unmarshal([]byte(tagsJSON), &memory.Tags); err != nil {
		log.Printf("[SQLiteMemoryStore] Warning: failed to unmarshal tags: %v", err)
		memory.Tags = []string{}
	}

	// Deserialize embedding
	memory.Embedding, err = deserializeEmbedding(embeddingBlob)
	if err != nil {
		log.Printf("[SQLiteMemoryStore] Warning: failed to deserialize embedding: %v", err)
	}

	// Convert timestamps
	memory.CreatedAt = time.Unix(0, createdAtUnix)
	memory.UpdatedAt = time.Unix(0, updatedAtUnix)

	return &memory, nil
}

// SetSemanticEngine sets the semantic search engine for this store
func (s *SQLiteMemoryStore) SetSemanticEngine(engine SemanticSearchEngine) {
	s.semanticEngine = engine
}

// GenerateEmbeddings generates embeddings for all memories that don't have them
func (s *SQLiteMemoryStore) GenerateEmbeddings() error {
	if s.semanticEngine == nil {
		return fmt.Errorf("semantic engine not initialized")
	}

	// Get all memories
	memories, err := s.List("")
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	log.Printf("[SQLiteMemoryStore] Processing %d memories for embedding generation", len(memories))

	generatedCount := 0
	for _, memory := range memories {
		// Generate embedding if needed
		if err := s.semanticEngine.GenerateEmbedding(memory); err != nil {
			log.Printf("[SQLiteMemoryStore] Failed to generate embedding for memory %s: %v", memory.ID, err)
			continue
		}

		// Check if embedding was actually generated and save if updated
		if len(memory.Embedding) > 0 {
			if _, err := s.Save(memory); err != nil {
				log.Printf("[SQLiteMemoryStore] Failed to save memory with embedding %s: %v", memory.ID, err)
				continue
			}
			generatedCount++
		}
	}

	if generatedCount > 0 {
		log.Printf("[SQLiteMemoryStore] Successfully generated %d embeddings", generatedCount)
	} else {
		log.Printf("[SQLiteMemoryStore] No new embeddings generated")
	}

	return nil
}

// GetSemanticStats returns statistics about semantic search functionality
func (s *SQLiteMemoryStore) GetSemanticStats() map[string]interface{} {
	stats := map[string]interface{}{
		"semantic_engine_available": s.semanticEngine != nil,
		"storage_type":              "sqlite",
	}

	if s.semanticEngine != nil {
		engineStats := s.semanticEngine.GetStats()
		for k, v := range engineStats {
			stats[k] = v
		}
	}

	// Count memories and embeddings from database
	var totalMemories, memoriesWithEmbeddings int

	// Count total memories
	row := s.db.QueryRow("SELECT COUNT(*) FROM memories")
	if err := row.Scan(&totalMemories); err != nil {
		log.Printf("[SQLiteMemoryStore] Warning: failed to count total memories: %v", err)
	}

	// Count memories with embeddings
	row = s.db.QueryRow("SELECT COUNT(*) FROM memories WHERE embedding IS NOT NULL AND embedding != ''")
	if err := row.Scan(&memoriesWithEmbeddings); err != nil {
		log.Printf("[SQLiteMemoryStore] Warning: failed to count memories with embeddings: %v", err)
	}

	stats["total_memories"] = totalMemories
	stats["memories_with_embeddings"] = memoriesWithEmbeddings
	if totalMemories > 0 {
		stats["embedding_coverage"] = float64(memoriesWithEmbeddings) / float64(totalMemories)
	} else {
		stats["embedding_coverage"] = 0.0
	}

	return stats
}
