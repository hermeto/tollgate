package budget

import "testing"

func TestReserveSoftAndHardLimits(t *testing.T) {
	l := New(map[string]Limit{"team": {SoftUSD: 1, HardUSD: 2}})
	if l.Reserve("team", .8).Action != Allow {
		t.Fatal("unexpected warning")
	}
	if l.Reserve("team", .3).Action != Warn {
		t.Fatal("soft limit not reported")
	}
	l.Commit("team", .8)
	if l.Reserve("team", 1).Action != Block {
		t.Fatal("hard limit not enforced")
	}
}
