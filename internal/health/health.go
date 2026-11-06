// Package health tracks dependency reachability and readiness.
package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/hermeto/tollgate/internal/config"
)

type Status struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	FailMode  string `json:"fail_mode"`
	Reachable bool   `json:"reachable"`
	Error     string `json:"error,omitempty"`
}
type Manager struct {
	deps     []config.DependencyConfig
	mu       sync.RWMutex
	statuses map[string]Status
	logger   *slog.Logger
}

func New(deps []config.DependencyConfig, logger *slog.Logger) *Manager {
	return &Manager{deps: deps, statuses: map[string]Status{}, logger: logger}
}

func (m *Manager) Start(ctx context.Context) { go m.loop(ctx) }
func (m *Manager) loop(ctx context.Context) {
	m.check(ctx)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.check(ctx)
		}
	}
}
func (m *Manager) check(ctx context.Context) {
	for _, dep := range m.deps {
		checkCtx, cancel := context.WithTimeout(ctx, time.Second)
		conn, err := (&net.Dialer{}).DialContext(checkCtx, "tcp", dep.Address)
		cancel()
		if conn != nil {
			_ = conn.Close()
		}
		next := Status{Name: dep.Name, Address: dep.Address, FailMode: dep.FailMode, Reachable: err == nil}
		if err != nil {
			next.Error = err.Error()
		}
		m.mu.Lock()
		old, existed := m.statuses[dep.Name]
		m.statuses[dep.Name] = next
		m.mu.Unlock()
		if m.logger != nil && (!existed || old.Reachable != next.Reachable) {
			if next.Reachable {
				m.logger.Info("dependency reachable", "dependency", dep.Name)
			} else {
				m.logger.Warn("dependency unreachable", "dependency", dep.Name, "error", next.Error)
			}
		}
	}
}
func (m *Manager) Ready() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, d := range m.deps {
		s, ok := m.statuses[d.Name]
		if !ok && d.FailMode == "closed" {
			return false
		}
		if ok && !s.Reachable && d.FailMode == "closed" {
			return false
		}
	}
	return true
}
func (m *Manager) Report() []Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Status, 0, len(m.statuses))
	for _, s := range m.statuses {
		out = append(out, s)
	}
	return out
}
func (m *Manager) MarshalJSON() ([]byte, error) { return json.Marshal(m.Report()) }
