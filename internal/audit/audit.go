// Package audit writes tamper-evident, content-free audit records.
package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Event struct {
	Timestamp    time.Time `json:"timestamp"`
	Action       string    `json:"action"`
	Actor        string    `json:"actor,omitempty"`
	Resource     string    `json:"resource,omitempty"`
	PreviousHash string    `json:"previous_hash,omitempty"`
	Hash         string    `json:"hash"`
}
type Log struct {
	mu     sync.Mutex
	events []Event
}

func New() *Log { return &Log{} }
func (l *Log) Append(action, actor, resource string) Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Event{Timestamp: time.Now().UTC(), Action: action, Actor: actor, Resource: resource}
	if len(l.events) > 0 {
		e.PreviousHash = l.events[len(l.events)-1].Hash
	}
	b, _ := json.Marshal(e)
	sum := sha256.Sum256(b)
	e.Hash = hex.EncodeToString(sum[:])
	l.events = append(l.events, e)
	return e
}
func (l *Log) Verify() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	prev := ""
	for _, e := range l.events {
		if e.PreviousHash != prev {
			return fmt.Errorf("audit chain broken")
		}
		hash := e.Hash
		e.Hash = ""
		b, _ := json.Marshal(e)
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != hash {
			return fmt.Errorf("audit hash mismatch")
		}
		prev = hash
	}
	return nil
}
func (l *Log) Events() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Event(nil), l.events...)
}
