package main

import (
	"sync"
	"time"
)

const dataPointsLen = 50

type model struct {
	TargetIP string
	Interval time.Duration

	lock       sync.RWMutex
	dataPoints [dataPointsLen]dataPoint
	next       int
	wrapped    bool
}

type dataPoint struct {
	Time     time.Time
	Duration time.Duration
	Result   result
}

type result string

const (
	indeterminate result = "indeterminate"
	connErr              = "connect error"
	sendErr              = "send error"
	readErr              = "read error"
	parseErr             = "parse error"
	ok                   = "OK"
)

func (m *model) Add(t time.Time, d time.Duration, r result) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.dataPoints[m.next] = dataPoint{t, d, r}
	m.next = (m.next + 1) % dataPointsLen
	if m.next == 0 {
		m.wrapped = true
	}
}

func (m *model) Get() []dataPoint {
	m.lock.RLock()
	defer m.lock.RUnlock()

	res := make([]dataPoint, 0, dataPointsLen)
	if m.wrapped {
		res = append(res, m.dataPoints[m.next:dataPointsLen]...)
	}
	return append(res, m.dataPoints[0:m.next]...)
}
