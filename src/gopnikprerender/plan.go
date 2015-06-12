package main

import (
	"fmt"
	"sync"

	"gopnik"
)

const (
	PENDING    = 0
	INPROGRESS = 1
	DONE       = 2
)

type plan struct {
	bboxes []gopnik.TileCoord
	status []uint8
	cursor int
	mu     sync.Mutex
	cond   *sync.Cond
	condMu sync.Mutex
}

func newPlan(bboxes []gopnik.TileCoord) *plan {
	self := &plan{
		bboxes: bboxes,
		status: make([]uint8, len(bboxes)),
	}
	self.cond = sync.NewCond(&self.condMu)
	return self
}

func (self *plan) TotalTasks() int {
	return len(self.bboxes)
}

func (self *plan) DoneTasks() int {
	self.mu.Lock()
	defer self.mu.Unlock()

	return self.countTasks(DONE)
}

func (self *plan) countTasks(status uint8) int {
	result := 0
	for _, s := range self.status {
		if s == status {
			result++
		}
	}
	return result
}

func (self *plan) GetTask() *gopnik.TileCoord {
	self.mu.Lock()
	defer self.mu.Unlock()

	oldCursor := self.cursor
	allDone := true
	for {
		self.cursor++
		if self.cursor == len(self.bboxes) {
			self.cursor = 0
		}

		if self.status[self.cursor] == PENDING {
			bb := self.bboxes[self.cursor]
			self.status[self.cursor] = INPROGRESS
			return &bb
		}
		if self.status[self.cursor] != DONE {
			allDone = false
		}

		// On start point again
		if self.cursor == oldCursor {
			// Is plan complete?
			if allDone {
				return nil
			}
			// Waiting...
			log.Debug("Waiting for %v tasks...", self.countTasks(INPROGRESS))
			self.condMu.Lock()
			self.mu.Unlock()
			self.cond.Wait()
			self.condMu.Unlock()
			self.mu.Lock()
			oldCursor = self.cursor
			allDone = true
		}
	}
}

func (self *plan) setStatus(coord gopnik.TileCoord, status uint8) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	for i, c := range self.bboxes {
		if coord.Equals(&c) {
			self.status[i] = status
			self.cond.Broadcast()
			return nil
		}
	}
	return fmt.Errorf("Can't find task %v", coord)
}

func (self *plan) FailTask(coord gopnik.TileCoord) error {
	return self.setStatus(coord, PENDING)
}

func (self *plan) DoneTask(coord gopnik.TileCoord) error {
	return self.setStatus(coord, DONE)
}
