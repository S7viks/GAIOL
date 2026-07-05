package orchestration

import (
	"sync"
	"time"
)

// TrustRepository persists model trust posteriors.
type TrustRepository interface {
	GetTrust(modelID, domain string) (*TrustRecord, error)
	UpsertTrust(record TrustRecord) error
	ListByDomain(domain string) ([]TrustRecord, error)
	ListAll() ([]TrustRecord, error)
}

// TraceRepository stores orchestration traces.
type TraceRepository interface {
	Append(trace OrchestrationTrace) error
	Get(traceID string) (*OrchestrationTrace, error)
	ListTraceIDs(limit int) ([]string, error)
}

// MemoryTrustStore is an in-process trust store.
type MemoryTrustStore struct {
	mu   sync.RWMutex
	rows map[string]TrustRecord // key modelID|domain
}

func NewMemoryTrustStore() *MemoryTrustStore {
	return &MemoryTrustStore{rows: make(map[string]TrustRecord)}
}

func trustKey(modelID, domain string) string { return modelID + "|" + domain }

func (s *MemoryTrustStore) GetTrust(modelID, domain string) (*TrustRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if r, ok := s.rows[trustKey(modelID, domain)]; ok {
		cp := r
		return &cp, nil
	}
	return nil, nil
}

func (s *MemoryTrustStore) UpsertTrust(record TrustRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows[trustKey(record.ModelID, record.Domain)] = record
	return nil
}

func (s *MemoryTrustStore) ListByDomain(domain string) ([]TrustRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []TrustRecord
	for _, r := range s.rows {
		if r.Domain == domain {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *MemoryTrustStore) ListAll() ([]TrustRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]TrustRecord, 0, len(s.rows))
	for _, r := range s.rows {
		out = append(out, r)
	}
	return out, nil
}

// MemoryTraceStore is an in-process trace store.
type MemoryTraceStore struct {
	mu      sync.RWMutex
	traces  map[string]OrchestrationTrace
	order   []string
}

func NewMemoryTraceStore() *MemoryTraceStore {
	return &MemoryTraceStore{
		traces: make(map[string]OrchestrationTrace),
		order:  make([]string, 0, 64),
	}
}

func (s *MemoryTraceStore) Append(trace OrchestrationTrace) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.traces[trace.TraceID]; !exists {
		s.order = append(s.order, trace.TraceID)
	}
	s.traces[trace.TraceID] = trace
	return nil
}

func (s *MemoryTraceStore) Get(traceID string) (*OrchestrationTrace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.traces[traceID]; ok {
		cp := t
		return &cp, nil
	}
	return nil, nil
}

func (s *MemoryTraceStore) ListTraceIDs(limit int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		limit = 50
	}
	n := len(s.order)
	if n > limit {
		n = limit
	}
	start := len(s.order) - n
	if start < 0 {
		start = 0
	}
	out := make([]string, n)
	copy(out, s.order[start:])
	return out, nil
}

func trustNow() string { return time.Now().UTC().Format(time.RFC3339Nano) }
