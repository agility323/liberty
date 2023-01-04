package lbtactor

import (
	"sync/atomic"
	"time"

	"github.com/agility323/liberty/lbtutil"
)

type taskWithNoReturn func()

type WorkerActor struct {
	name string
	state int32
	activet int32
	taskq chan taskWithNoReturn
	stopq chan struct{}
}

func NewWorkerActor(name string) *WorkerActor {
	return &WorkerActor{
		name: name,
		state: 0,
		taskq: nil,
		stopq: nil,
	}
}

func (w *WorkerActor) Start(qlen int) bool {
	if !atomic.CompareAndSwapInt32(&w.state, 0, 1) { return false }
	w.taskq = make(chan taskWithNoReturn, qlen)
	w.stopq = make(chan struct{}, 1)
	go w.workLoop(qlen)
	return true
}

func (w *WorkerActor) workLoop(qlen int) {
	defer lbtutil.Recover("WorkerActor.workLoop " + w.name, func() { go w.workLoop(qlen) })

	for {
		select {
		case <-w.stopq:
			return
		case task := <-w.taskq:
			task()
		}
	}
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
