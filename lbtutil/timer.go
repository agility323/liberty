package lbtutil

import (
	"time"
)

func StartTickJob(name string, tickTime int, stopCh chan bool, f func()) {
	stopped := false
	ticker := time.NewTicker(time.Duration(tickTime) * time.Second)
	defer func() {
		ticker.Stop()
		if !stopped {
			go StartTickJob(name, tickTime, stopCh, f)
		}
	}()
	log.Info("tick job start %s", name)
	for {
		select {
		case <- stopCh:
			stopped = true
			log.Info("tick job stopped %s", name)
			return
		case <- ticker.C:
			f()
		}
	}
}
