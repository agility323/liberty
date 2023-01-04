package lbtactor

type actorTask func()

type WorkerActor struct {
	taskq chan actorTask
}

func NewWorkerActor(qlen int) *WorkerActor {
	return &WorkerActor{
		taskq: make(chan actorTask, qlen),
	}
}

func (w *WorkerActor) Start() {
	go func() {
		for task := range w.taskq {
			task()
		}
	}()
}

func (w *WorkerActor) Stop() {
	close(w.taskq)
}

func (w *WorkerActor) PushTask(task actorTask) bool {
	select {
	case w.taskq <- task:
		return true
	default:
		return false
	}
	return false
}
