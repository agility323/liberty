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
}

func NewWorkerActor() *WorkerActor {
	return &WorkerActor{
		state: 0,
		taskq: nil,
	}
}

func (w *WorkerActor) Start(qlen int) bool {
	if !atomic.CompareAndSwapInt32(&w.state, 0, 1) { return false }
	w.taskq = make(chan taskWithNoReturn, qlen)
	go func() {
		for task := range w.taskq {
			task()
		}
	} ()
	return true
}

func (w *WorkerActor) Stop() bool {
	if !atomic.CompareAndSwapInt32(&w.state, 1, 2) { return false }
	close(w.taskq)
	return true
}

func (w *WorkerActor) PushTask(task taskWithNoReturn) bool {
	select {
	case w.taskq <- task:
		atomic.StoreInt32(&w.activet, int32(time.Now().Unix()))
		return true
	default:
		return false
	}
	return false
}
