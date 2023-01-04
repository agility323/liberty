package service_framework

import (
	"sync"
	"time"

	"github.com/agility323/liberty/lbtutil"
)

const MinTickTime = 30

var tickmgr *TickManager

func init() {
	interval := 600
	tickmgr = &TickManager{
		ticker: time.NewTicker(time.Duration(interval) * time.Second),
		stopq: make(chan struct{}, 1),
		interval: interval,
		idcnt: 0,
		jobs: make(map[uint64]func()),
	}
}

type TickManager struct {
	lock sync.RWMutex
	ticker *time.Ticker
	stopq chan struct{}
	interval int
	idcnt uint64
	jobs map[uint64]func()
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
	logger.Info("tick manager tick time set to %d", interval)
}

func (m *TickManager) ResetTickTime(interval int) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.interval = interval
	m.resetTickTime()
}

func (m *TickManager) Start() {
	defer lbtutil.Recover("TickManager.Start", m.Start)

	go func() {
		logger.Info("tick manager start")
		for {
			for id, job := range m.jobs {
				select {
				case <-m.stopq:
					logger.Info("tick manager stop")
					return
				case <-m.ticker.C:
					logger.Info("tick job start %d", id)
					job()
				}
			}
		}
	}()
}

func (m *TickManager) Stop() {
	m.stopq<- struct{}{}
}
