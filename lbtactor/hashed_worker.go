package lbtactor

import (
	"fmt"
)

type HashedWorker struct {
	ws []*Worker
	size int
	qlen int
}

var MaxHashedWorkerCap int = 10000

func NewHashedWorker(size int, name string, qlen int) *HashedWorker {
	if qlen * size > MaxHashedWorkerCap {
		panic(fmt.Sprintf("HashedWorker size too big %d * %d", size, qlen))
	}
	ws := make([]*Worker, size, size)
	for i := 0; i < size; i++ {
		ws[i] = NewWorker(fmt.Sprintf("%s-%d", name, i), qlen)
	}
	return &HashedWorker{
		ws: ws,
		size: size,
		qlen: qlen,
	}
}

func (hw *HashedWorker) Start() {
	for _, w := range hw.ws {
		w.Start()
	}
}

func (hw *HashedWorker) Stop() {
	for _, w := range hw.ws {
		w.Stop()
	}

}

func (hw *HashedWorker) PushTask(task taskWithNoReturn, hval int) bool {
	return hw.ws[hval % len(hw.ws)].PushTask(task)
}
