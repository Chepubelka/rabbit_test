package pool

import "log"

type Job struct {
	ID			int32
	Resources	string
}

type Pool struct {
	NumWorkers  int32
	JobChannels chan chan Job
	JobQueue    chan Job
	Stopped     chan bool
}

type Worker struct {
	ID          int
	JobChannel  chan Job
	JobChannels chan chan Job
	Quit        chan bool
}

func NewPool(numworkers int32) Pool {
	return Pool{
		NumWorkers:  numworkers,
		JobChannels: make(chan chan Job),
		JobQueue:    make(chan Job),
		Stopped:     make(chan bool),
	}
}

func (p *Pool) Run() {
	for i := 0; i < int(p.NumWorkers); i++ {
		worker := Worker{
			ID:          (i + 1),
			JobChannel:  make(chan Job),
			JobChannels: p.JobChannels,
			Quit:        make(chan bool),
		}
		worker.Start()
	}
	p.Allocate()
}

func (p *Pool) Allocate() {
	q := p.JobQueue
	s := p.Stopped
	go func(queue chan Job) {
		for {
			select {
			case job := <-q:
				// get from the JobChannels
				availChannel := <-p.JobChannels
				availChannel <- job

			case <-s:
				return
			}
		}
	}(q)
}

func (w *Worker) Start() {
	go func() {
		for {
			w.JobChannels <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				w.work(job)
			case <-w.Quit:
				return
			}

		}
	}()
}

func (w *Worker) work(job Job) {
	log.Printf("------")
	log.Printf("Processed by Worker [%d]", w.ID)
	log.Printf("Processed Job With ID [%d] & content: [%s]", job.ID, job.Resources)
	log.Printf("-------")
}
