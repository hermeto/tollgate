// Package budget provides deterministic spend reservations and threshold decisions.
package budget

import "sync"

type Action string

const (
	Allow Action = "allow"
	Warn  Action = "warn"
	Block Action = "block"
)

type Limit struct{ SoftUSD, HardUSD float64 }
type Status struct {
	Spent, Reserved, Soft, Hard float64
	Action                      Action
}
type Ledger struct {
	mu              sync.Mutex
	limits          map[string]Limit
	spent, reserved map[string]float64
}

func New(limits map[string]Limit) *Ledger {
	return &Ledger{limits: limits, spent: map[string]float64{}, reserved: map[string]float64{}}
}
func (l *Ledger) Reserve(key string, amount float64) Status {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim := l.limits[key]
	total := l.spent[key] + l.reserved[key]
	action := Allow
	if lim.HardUSD > 0 && total+amount > lim.HardUSD {
		action = Block
	} else if lim.SoftUSD > 0 && total+amount >= lim.SoftUSD {
		action = Warn
	}
	if action != Block {
		l.reserved[key] += amount
	}
	return l.status(key, lim, action)
}
func (l *Ledger) Commit(key string, amount float64) Status {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.reserved[key] -= amount
	if l.reserved[key] < 0 {
		l.reserved[key] = 0
	}
	l.spent[key] += amount
	lim := l.limits[key]
	return l.status(key, lim, Allow)
}
func (l *Ledger) status(key string, lim Limit, action Action) Status {
	return Status{Spent: l.spent[key], Reserved: l.reserved[key], Soft: lim.SoftUSD, Hard: lim.HardUSD, Action: action}
}
