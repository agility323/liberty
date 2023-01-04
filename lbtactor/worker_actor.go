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
	done chan struct{}
	closed bool
}

func NewWorkerActor() *WorkerActor {
	return &WorkerActor{
		state: 0,
		taskq: nil,
		done: nil,
		closed: false,
	}
}

func (w *WorkerActor) Start(qlen int) bool {
	if !atomic.CompareAndSwapInt32(&w.state, 0, 1) { return false }
	w.taskq = make(chan taskWithNoReturn, qlen)
	w.done = make(chan struct{} , 1)
	go func() {
		for {
			select {
			case <-w.done:
				w.closed = true
				return
			case task:=<-w.taskq:
				task()
			}
		}
	} ()
	return true
}

func (w *WorkerActor) Stop() bool {
	if !atomic.CompareAndSwapInt32(&w.state, 1, 2) { return false }
	w.done <- struct{}{}
	return true
}

func (w *WorkerActor) PushTask(task taskWithNoReturn) bool {
	if w.closed{
		return false
	}
	select {
	case w.taskq <- task:
		atomic.StoreInt32(&w.activet, int32(time.Now().Unix()))
		return true
	default:
		return false
	}
	return false
}
