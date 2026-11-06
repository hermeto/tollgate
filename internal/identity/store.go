// Package identity manages high-entropy virtual keys without persisting secrets.
package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type Record struct {
	ID        string     `json:"id"`
	Prefix    string     `json:"prefix"`
	Hash      string     `json:"hash"`
	Team      string     `json:"team"`
	Project   string     `json:"project"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}
type Identity struct{ Prefix, Team, Project string }
type disk struct {
	Records []Record `json:"records"`
}
type Store struct {
	mu      sync.RWMutex
	path    string
	records map[string]Record
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, records: map[string]Record{}}
	if path == "" {
		return s, nil
	}
	body, err := os.ReadFile(path) //nolint:gosec // path is an explicit operator-selected local store.
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var d disk
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, fmt.Errorf("identity store: %w", err)
	}
	for _, r := range d.Records {
		s.records[r.ID] = r
	}
	return s, nil
}
func (s *Store) Create(team, project string, expiresAt *time.Time) (string, Record, error) {
	if team == "" || project == "" {
		return "", Record{}, fmt.Errorf("team and project are required")
	}
	secret, err := randomString(32)
	if err != nil {
		return "", Record{}, err
	}
	prefix, err := randomString(8)
	if err != nil {
		return "", Record{}, err
	}
	raw := "tg_" + prefix + "_" + secret
	now := time.Now().UTC()
	r := Record{ID: prefix, Prefix: "tg_" + prefix, Hash: hash(raw), Team: team, Project: project, CreatedAt: now, ExpiresAt: expiresAt}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[r.ID] = r
	if err := s.persistLocked(); err != nil {
		delete(s.records, r.ID)
		return "", Record{}, err
	}
	return raw, r, nil
}
func (s *Store) Authenticate(raw string, now time.Time) (Identity, bool) {
	digest := hash(raw)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.records {
		if subtle.ConstantTimeCompare([]byte(digest), []byte(r.Hash)) != 1 || r.RevokedAt != nil || (r.ExpiresAt != nil && !now.Before(*r.ExpiresAt)) {
			continue
		}
		return Identity{Prefix: r.Prefix, Team: r.Team, Project: r.Project}, true
	}
	return Identity{}, false
}
func (s *Store) List() []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Record, 0, len(s.records))
	for _, r := range s.records {
		r.Hash = ""
		out = append(out, r)
	}
	return out
}
func (s *Store) Revoke(prefix string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, r := range s.records {
		if r.Prefix == prefix || id == strings.TrimPrefix(prefix, "tg_") {
			now := time.Now().UTC()
			r.RevokedAt = &now
			s.records[id] = r
			return s.persistLocked()
		}
	}
	return os.ErrNotExist
}
func (s *Store) Rotate(prefix string, expiresAt *time.Time) (string, Record, error) {
	s.mu.Lock()
	var oldID string
	for id, r := range s.records {
		if r.Prefix == prefix || id == strings.TrimPrefix(prefix, "tg_") {
			now := time.Now().UTC()
			r.RevokedAt = &now
			s.records[id] = r
			oldID = id
			break
		}
	}
	if oldID == "" {
		s.mu.Unlock()
		return "", Record{}, os.ErrNotExist
	}
	old := s.records[oldID]
	secret, err := randomString(32)
	if err != nil {
		s.mu.Unlock()
		return "", Record{}, err
	}
	suffix, err := randomString(8)
	if err != nil {
		s.mu.Unlock()
		return "", Record{}, err
	}
	raw := "tg_" + suffix + "_" + secret
	r := Record{ID: suffix, Prefix: "tg_" + suffix, Hash: hash(raw), Team: old.Team, Project: old.Project, CreatedAt: time.Now().UTC(), ExpiresAt: expiresAt}
	s.records[suffix] = r
	err = s.persistLocked()
	s.mu.Unlock()
	return raw, r, err
}
func (s *Store) persistLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	records := make([]Record, 0, len(s.records))
	for _, r := range s.records {
		records = append(records, r)
	}
	body, err := json.MarshalIndent(disk{Records: records}, "", "  ")
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(s.path), ".keys-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(body); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(name, s.path)
}
func hash(raw string) string { sum := sha256.Sum256([]byte(raw)); return hex.EncodeToString(sum[:]) }
func randomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, b := range bytes {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out), nil
}
