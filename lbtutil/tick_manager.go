package lbtutil

import (
	"sync"
	"time"
)

const MinTickTime = 30

type TickManager struct {
	lock sync.RWMutex
	ticker *time.Ticker
	stopq chan struct{}
	interval int
	idcnt uint64
	jobs map[uint64]func()
}

func NewTickManager(interval int) *TickManager {
	return &TickManager{
		ticker: time.NewTicker(time.Duration(interval) * time.Second),
		stopq: make(chan struct{}, 1),
		interval: interval,
		idcnt: 0,
		jobs: make(map[uint64]func()),
	}
}

func (m *TickManager) AddTickJob(job func()) uint64 {
	m.lock.Lock()
	defer m.lock.Unlock()

	// add job
	m.idcnt += 1
	id := m.idcnt
	m.jobs[id] = job
	// reset ticker
	m.resetTickTime()
	return id
}

func (m *TickManager) DelTickJob(id uint64) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.jobs, id)
	m.resetTickTime()
}

func (m *TickManager) resetTickTime() {
	n := len(m.jobs)
	if n == 0 { n = 1 }
	interval := (m.interval + n - 1) / n
	if interval < MinTickTime { interval = MinTickTime }
	m.ticker.Reset(time.Duration(interval) * time.Second)
	log.Info("tick manager tick time set to %d", interval)
}

func (m *TickManager) ResetTickTime(interval int) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.interval = interval
	m.resetTickTime()
}

func (m *TickManager) Start() {
	defer Recover("TickManager.Start", m.Start)

	go func() {
		log.Info("tick manager start")
		for {
			for id, job := range m.jobs {
				select {
				case <-m.stopq:
					log.Info("tick manager stop")
					return
				case <-m.ticker.C:
					log.Info("tick job start %d", id)
					job()
				}
			}
		}
	}()
}

func (m *TickManager) Stop() {
	m.stopq<- struct{}{}
}
