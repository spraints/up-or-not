package main

import (
	"testing"
	"time"
)

func TestModelCircular(t *testing.T) {
	var m model

	vals := m.Get()
	if len(vals) != 0 {
		t.Errorf("expected empty, got %#v", vals)
	}

	start := time.Now()
	m.Add(start, time.Second, ok)

	vals = m.Get()
	if len(vals) != 1 || vals[0].Time != start {
		t.Errorf("expected {%v}, got %v", start, vals)
	}

	nextT := start
	toAdd := dataPointsLen * 3 / 2
	for i := 1; i < toAdd; i++ {
		nextT = nextT.Add(time.Second)
		m.Add(nextT, time.Second, ok)
	}

	vals = m.Get()
	if len(vals) != dataPointsLen || vals[len(vals)-1].Time != nextT {
		t.Errorf("expected %d elements {..., %v}, got %d %v", dataPointsLen, nextT, len(vals), vals)
	}
}
