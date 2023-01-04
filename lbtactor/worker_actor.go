package lbtactor

import (
	"sync/atomic"
	"time"
)

type taskWithNoReturn func()

type WorkerActor struct {
	state int32
	activet int32
	taskq chan taskWithNoReturn
	stopq chan struct{}
}

func NewWorkerActor() *WorkerActor {
	return &WorkerActor{
		state: 0,
		taskq: nil,
		stopq: nil,
	}
}

func (w *WorkerActor) Start(qlen int) bool {
	if !atomic.CompareAndSwapInt32(&w.state, 0, 1) { return false }
	w.taskq = make(chan taskWithNoReturn, qlen)
	w.stopq = make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-w.stopq:
				return
			case task := <-w.taskq:
				task()
			}
		}
	} ()
	return true
}

func (w *WorkerActor) Stop() bool {
	if !atomic.CompareAndSwapInt32(&w.state, 1, 2) { return false }
	select {
	case w.stopq<- struct{}{}:
		return true
	default:
		return false
	}
}

func (w *WorkerActor) PushTask(task taskWithNoReturn) bool {
	if w.state == 2 {
		return false
	}
	select {
	case w.taskq<- task:
		atomic.StoreInt32(&w.activet, int32(time.Now().Unix()))
		return true
	default:
		return false
	}
	return false
}
