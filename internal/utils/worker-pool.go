package utils

type WorkerPool interface {
	Work()
	Done()
}

// Implement buffered channel to only allow X workers to work at at time
type workerPool chan bool

func (p workerPool) Work() {
	p <- true
}

func (p workerPool) Done() {
	<-p
}

func NewWorkerPool(maxWorkers int) WorkerPool {
	return make(workerPool, maxWorkers)
}
