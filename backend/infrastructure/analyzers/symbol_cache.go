package analyzers

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"shotgun_code/domain/analysis"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// CachedSymbolIndex wraps SymbolIndexImpl with SQLite persistence
type CachedSymbolIndex struct {
	*SymbolIndexImpl
	db          *sql.DB
	dbPath      string
	projectRoot string
	mu          sync.Mutex
}

// NewCachedSymbolIndex creates a symbol index with SQLite caching
func NewCachedSymbolIndex(registry analysis.AnalyzerRegistry, cacheDir string) (*CachedSymbolIndex, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(cacheDir, "symbols.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		return nil, err
	}

	idx := &CachedSymbolIndex{
		SymbolIndexImpl: NewSymbolIndex(registry),
		db:              db,
		dbPath:          dbPath,
	}

	if err := idx.initDB(); err != nil {
		db.Close()
		return nil, err
	}

	return idx, nil
}

func (idx *CachedSymbolIndex) initDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		path TEXT PRIMARY KEY,
		hash TEXT NOT NULL,
		indexed_at INTEGER NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS symbols (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		name TEXT NOT NULL,
		kind TEXT NOT NULL,
		language TEXT,
		start_line INTEGER,
		end_line INTEGER,
		signature TEXT,
		doc_comment TEXT,
		parent TEXT,
		extra TEXT,
		FOREIGN KEY (file_path) REFERENCES files(path) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name);
	CREATE INDEX IF NOT EXISTS idx_symbols_file ON symbols(file_path);
	CREATE INDEX IF NOT EXISTS idx_symbols_kind ON symbols(kind);
	CREATE INDEX IF NOT EXISTS idx_symbols_name_lower ON symbols(lower(name));
	`
	_, err := idx.db.Exec(schema)
	return err
}


// IndexProject indexes project with caching
func (idx *CachedSymbolIndex) IndexProject(ctx context.Context, projectRoot string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.projectRoot = projectRoot
	idx.SymbolIndexImpl.Clear()

	// Load cached symbols first
	cachedFiles := make(map[string]string) // path -> hash
	rows, err := idx.db.Query("SELECT path, hash FROM files")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var path, hash string
			rows.Scan(&path, &hash)
			cachedFiles[path] = hash
		}
	}

	var filesToIndex []string
	var filesToRemove []string

	// Walk project and check what needs indexing
	err = filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "build" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		analyzer := idx.registry.GetAnalyzer(path)
		if analyzer == nil {
			return nil
		}

		relPath, _ := filepath.Rel(projectRoot, path)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		currentHash := hashContent(content)
		if cachedHash, exists := cachedFiles[relPath]; exists {
			if cachedHash == currentHash {
				// Load from cache
				delete(cachedFiles, relPath)
				return nil
			}
		}
		filesToIndex = append(filesToIndex, relPath)
		delete(cachedFiles, relPath)
		return nil
	})

	// Files in cachedFiles that weren't visited should be removed
	for path := range cachedFiles {
		filesToRemove = append(filesToRemove, path)
	}

	// Remove stale files from cache
	for _, path := range filesToRemove {
		idx.removeFileFromCache(path)
	}

	// Index new/changed files
	for _, relPath := range filesToIndex {
		fullPath := filepath.Join(projectRoot, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		idx.indexFileWithCache(ctx, relPath, content)
	}

	// Load all symbols from cache into memory
	idx.loadFromCache()

	idx.indexed = true
	return err
}

func (idx *CachedSymbolIndex) indexFileWithCache(ctx context.Context, filePath string, content []byte) error {
	analyzer := idx.registry.GetAnalyzer(filePath)
	if analyzer == nil {
		return nil
	}

	symbols, err := analyzer.ExtractSymbols(ctx, filePath, content)
	if err != nil {
		return err
	}

	hash := hashContent(content)

	// Begin transaction
	tx, err := idx.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove old symbols for this file
	tx.Exec("DELETE FROM symbols WHERE file_path = ?", filePath)
	tx.Exec("DELETE FROM files WHERE path = ?", filePath)

	// Insert file record
	tx.Exec("INSERT INTO files (path, hash, indexed_at) VALUES (?, ?, ?)",
		filePath, hash, time.Now().Unix())

	// Insert symbols
	stmt, err := tx.Prepare(`
		INSERT INTO symbols (file_path, name, kind, language, start_line, end_line, signature, doc_comment, parent, extra)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, sym := range symbols {
		extra := ""
		if sym.Extra != nil {
			if b, err := json.Marshal(sym.Extra); err == nil {
				extra = string(b)
			}
		}
		stmt.Exec(filePath, sym.Name, string(sym.Kind), sym.Language,
			sym.StartLine, sym.EndLine, sym.Signature, sym.DocComment, sym.Parent, extra)
	}

	return tx.Commit()
}

func (idx *CachedSymbolIndex) removeFileFromCache(filePath string) {
	idx.db.Exec("DELETE FROM symbols WHERE file_path = ?", filePath)
	idx.db.Exec("DELETE FROM files WHERE path = ?", filePath)
}

func (idx *CachedSymbolIndex) loadFromCache() {
	rows, err := idx.db.Query(`
		SELECT file_path, name, kind, language, start_line, end_line, signature, doc_comment, parent, extra
		FROM symbols
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sym analysis.Symbol
		var extra string
		rows.Scan(&sym.FilePath, &sym.Name, &sym.Kind, &sym.Language,
			&sym.StartLine, &sym.EndLine, &sym.Signature, &sym.DocComment, &sym.Parent, &extra)
		sym.Line = sym.StartLine
		if extra != "" {
			json.Unmarshal([]byte(extra), &sym.Extra)
		}
		idx.addSymbolLocked(sym)
	}
}

// IndexFile indexes a single file with caching
func (idx *CachedSymbolIndex) IndexFile(ctx context.Context, filePath string, content []byte) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove old symbols from memory index
	idx.removeSymbolsForFile(filePath)

	// Index with cache
	if err := idx.indexFileWithCache(ctx, filePath, content); err != nil {
		return err
	}

	// Add to memory index
	analyzer := idx.registry.GetAnalyzer(filePath)
	if analyzer == nil {
		return nil
	}

	symbols, err := analyzer.ExtractSymbols(ctx, filePath, content)
	if err != nil {
		return err
	}

	for _, sym := range symbols {
		idx.addSymbolLocked(sym)
	}
	return nil
}

func (idx *CachedSymbolIndex) removeSymbolsForFile(filePath string) {
	// Get indices of symbols to remove
	indices := idx.byFile[filePath]
	if len(indices) == 0 {
		return
	}

	// Create a set of indices to remove for O(1) lookup
	toRemove := make(map[int]bool)
	for _, i := range indices {
		toRemove[i] = true
	}

	// Remove from byName index
	for name, nameIndices := range idx.byName {
		var filtered []int
		for _, i := range nameIndices {
			if !toRemove[i] {
				filtered = append(filtered, i)
			}
		}
		if len(filtered) == 0 {
			delete(idx.byName, name)
		} else {
			idx.byName[name] = filtered
		}
	}

	// Remove from byKind index
	for kind, kindIndices := range idx.byKind {
		var filtered []int
		for _, i := range kindIndices {
			if !toRemove[i] {
				filtered = append(filtered, i)
			}
		}
		if len(filtered) == 0 {
			delete(idx.byKind, kind)
		} else {
			idx.byKind[kind] = filtered
		}
	}

	// Remove from byFile index
	delete(idx.byFile, filePath)

	// Note: We don't remove from symbols slice to avoid index invalidation
	// The symbols will be orphaned but won't be returned in queries
}

// Close closes the database connection
func (idx *CachedSymbolIndex) Close() error {
	if idx.db != nil {
		return idx.db.Close()
	}
	return nil
}

// GetCacheStats returns cache statistics
func (idx *CachedSymbolIndex) GetCacheStats() map[string]int {
	stats := idx.Stats()

	var fileCount, symbolCount int
	idx.db.QueryRow("SELECT COUNT(*) FROM files").Scan(&fileCount)
	idx.db.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&symbolCount)

	stats["cached_files"] = fileCount
	stats["cached_symbols"] = symbolCount
	return stats
}

// InvalidateCache clears the entire cache
func (idx *CachedSymbolIndex) InvalidateCache() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.SymbolIndexImpl.Clear()
	_, err := idx.db.Exec("DELETE FROM symbols; DELETE FROM files;")
	return err
}

func hashContent(content []byte) string {
	h := md5.Sum(content)
	return hex.EncodeToString(h[:])
}
