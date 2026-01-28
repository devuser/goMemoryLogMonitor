package logstore

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type LogEntry struct {
	Time    time.Time `json:"time"`
	Content string    `json:"content"`
}

type Store struct {
	mu            sync.RWMutex
	entries       []LogEntry
	maxBytes      int64
	currentBytes  int64
	approxOverhead int64
}

type QueryOptions struct {
	Page      int
	PageSize  int
	StartTime *time.Time
	EndTime   *time.Time
	Text      string
	TopN      int
	SortBy    string // "time" or "content"
	SortOrder string // "asc" or "desc"
}

type QueryResult struct {
	Items          []LogEntry
	Total          int
	CacheCount     int
	CacheSizeBytes int64
}

func NewStore(maxBytes int64) *Store {
	if maxBytes <= 0 {
		maxBytes = 100 * 1024 * 1024 // default 100MB
	}
	return &Store{
		entries:       make([]LogEntry, 0, 1024),
		maxBytes:      maxBytes,
		approxOverhead: 64,
	}
}

func (s *Store) addBytes(e LogEntry) int64 {
	return int64(len(e.Content)) + s.approxOverhead
}

func (s *Store) Append(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := LogEntry{
		Time:    time.Now(),
		Content: content,
	}
	size := s.addBytes(entry)

	s.entries = append(s.entries, entry)
	s.currentBytes += size

	for s.currentBytes > s.maxBytes && len(s.entries) > 0 {
		removed := s.entries[0]
		s.entries = s.entries[1:]
		s.currentBytes -= s.addBytes(removed)
		if s.currentBytes < 0 {
			s.currentBytes = 0
		}
	}
}

func (s *Store) Snapshot() ([]LogEntry, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]LogEntry, len(s.entries))
	copy(out, s.entries)
	return out, s.currentBytes
}

func (s *Store) Query(opt QueryOptions) QueryResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cacheCount := len(s.entries)
	cacheBytes := s.currentBytes

	var filtered []LogEntry
	for i := len(s.entries) - 1; i >= 0; i-- {
		e := s.entries[i]

		if opt.StartTime != nil && e.Time.Before(*opt.StartTime) {
			continue
		}
		if opt.EndTime != nil && e.Time.After(*opt.EndTime) {
			continue
		}
		if opt.Text != "" && !strings.Contains(e.Content, opt.Text) {
			continue
		}
		filtered = append(filtered, e)

		if opt.TopN > 0 && len(filtered) >= opt.TopN {
			break
		}
	}

	total := len(filtered)

	// Apply sorting
	if opt.SortBy != "" {
		sort.Slice(filtered, func(i, j int) bool {
			var less bool
			switch opt.SortBy {
			case "time":
				less = filtered[i].Time.Before(filtered[j].Time)
			case "content":
				less = filtered[i].Content < filtered[j].Content
			default:
				// Default to time descending if invalid sortBy
				less = filtered[i].Time.Before(filtered[j].Time)
			}
			// Reverse if descending order
			if opt.SortOrder == "desc" {
				return !less
			}
			return less
		})
	} else {
		// Default: time descending (newest first)
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Time.After(filtered[j].Time)
		})
	}

	if opt.Page <= 0 {
		opt.Page = 1
	}
	if opt.PageSize <= 0 {
		opt.PageSize = 50
	}

	start := (opt.Page - 1) * opt.PageSize
	if start >= total {
		return QueryResult{
			Items:          []LogEntry{},
			Total:          total,
			CacheCount:     cacheCount,
			CacheSizeBytes: cacheBytes,
		}
	}
	end := start + opt.PageSize
	if end > total {
		end = total
	}

	return QueryResult{
		Items:          filtered[start:end],
		Total:          total,
		CacheCount:     cacheCount,
		CacheSizeBytes: cacheBytes,
	}
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = make([]LogEntry, 0, 1024)
	s.currentBytes = 0
}

