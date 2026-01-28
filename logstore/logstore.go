package logstore

import (
	"container/list"
	"strings"
	"sync"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// LogStore manages log entries with LRU eviction
type LogStore struct {
	mu      sync.RWMutex
	entries *list.List
	maxSize int
}

// NewLogStore creates a new LogStore with the specified maximum size
func NewLogStore(maxSize int) *LogStore {
	return &LogStore{
		entries: list.New(),
		maxSize: maxSize,
	}
}

// Add adds a new log entry to the store
func (ls *LogStore) Add(entry LogEntry) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Add new entry to front
	ls.entries.PushFront(entry)

	// Evict oldest if over capacity
	if ls.entries.Len() > ls.maxSize {
		oldest := ls.entries.Back()
		if oldest != nil {
			ls.entries.Remove(oldest)
		}
	}
}

// GetAll returns all log entries
func (ls *LogStore) GetAll() []LogEntry {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	result := make([]LogEntry, 0, ls.entries.Len())
	for e := ls.entries.Front(); e != nil; e = e.Next() {
		result = append(result, e.Value.(LogEntry))
	}
	return result
}

// GetFiltered returns filtered log entries based on keyword and date range
func (ls *LogStore) GetFiltered(keyword string, startTime, endTime *time.Time) []LogEntry {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	result := make([]LogEntry, 0)
	for e := ls.entries.Front(); e != nil; e = e.Next() {
		entry := e.Value.(LogEntry)
		
		// Filter by date range
		if startTime != nil && entry.Timestamp.Before(*startTime) {
			continue
		}
		if endTime != nil && entry.Timestamp.After(*endTime) {
			continue
		}
		
		// Filter by keyword
		if keyword != "" && !strings.Contains(entry.Message, keyword) {
			continue
		}
		
		result = append(result, entry)
	}
	return result
}

// Count returns the number of entries in the store
func (ls *LogStore) Count() int {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.entries.Len()
}
