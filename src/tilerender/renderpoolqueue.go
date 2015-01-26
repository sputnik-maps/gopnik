package tilerender

import (
	"container/list"
	"errors"
	"sync"

	"gopnik"
)

type renderTask struct {
	ResultCh []chan<- *RenderPoolResponse
	Coord    gopnik.TileCoord
}

type renderQueue struct {
	mu      sync.Mutex
	tasksCh chan gopnik.TileCoord
	tasks   *list.List
}

func newRenderQueue(reqLimit uint) *renderQueue {
	rq := new(renderQueue)
	rq.tasksCh = make(chan gopnik.TileCoord, reqLimit)
	rq.tasks = list.New()
	return rq
}

func (rq *renderQueue) Size() int {
	return cap(rq.tasksCh)
}

func (rq *renderQueue) Push(coord gopnik.TileCoord, resCh chan<- *RenderPoolResponse) error {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	// Search task in progress
	for e := rq.tasks.Front(); e != nil; e = e.Next() {
		task := e.Value.(*renderTask)
		if task.Coord.Equals(&coord) {
			task.ResultCh = append(task.ResultCh, resCh)
			return nil
		}
	}

	// Enqueue rendering task
	select {
	case rq.tasksCh <- coord:
		// Ok, task now in queue
	default:
		return errors.New("Queue limit exceeded")
	}

	// Put task details in list
	rq.tasks.PushBack(&renderTask{
		Coord:    coord,
		ResultCh: []chan<- *RenderPoolResponse{resCh},
	})

	return nil
}

func (rq *renderQueue) TasksChan() <-chan gopnik.TileCoord {
	return rq.tasksCh
}

func (rq *renderQueue) Done(coord gopnik.TileCoord, resp *RenderPoolResponse) error {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	for e := rq.tasks.Front(); e != nil; e = e.Next() {
		task := e.Value.(*renderTask)
		if task.Coord.Equals(&coord) {
			// send response
			for _, ch := range task.ResultCh {
				ch <- resp
			}

			// remove task from list
			rq.tasks.Remove(e)
			return nil
		}
	}

	return errors.New("Done invalid task")
}
