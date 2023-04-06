package lbtactor

import (
	"sync/atomic"
	"time"

	"github.com/agility323/liberty/lbtutil"
)

type taskWithNoReturn func()

type Worker struct {
	name string
	state int32
	activet int32
	qlen int
	taskq chan taskWithNoReturn
	stopq chan struct{}
}

func NewWorker(name string, qlen int) *Worker {
	return &Worker{
		name: name,
		state: 0,
		activet: int32(time.Now().Unix()),
		qlen: qlen,
		taskq: nil,
		stopq: nil,
	}
}

func (w *Worker) Start() {
	if !atomic.CompareAndSwapInt32(&w.state, 0, 1) { return }
	w.taskq = make(chan taskWithNoReturn, w.qlen)
	w.stopq = make(chan struct{}, 1)
	go w.workLoop()
}

func (w *Worker) workLoop() {
	defer lbtutil.Recover("Worker.workLoop " + w.name, func() { go w.workLoop() })

	for {
		select {
		case <-w.stopq:
			return
		case task := <-w.taskq:
			task()
		}
	}
}

func (w *Worker) Stop() {
	if !atomic.CompareAndSwapInt32(&w.state, 1, 2) { return }
	select {
	case w.stopq<- struct{}{}:
	default:
	}
}

func (w *Worker) PushTask(task taskWithNoReturn) bool {
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
}
