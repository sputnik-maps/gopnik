package tilerender

import (
	"testing"
	"time"

	"gopnik"
)

func TestQueueSimple(t *testing.T) {
	rq := newRenderQueue(10)

	coord := gopnik.TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
	}
	resCh := make(chan *RenderPoolResponse, 1)

	err := rq.Push(coord, resCh)
	if err != nil {
		t.Errorf("Push error: %v", err)
	}
	coord2 := <-rq.TasksChan()
	if !coord.Equals(&coord2) {
		t.Error("Coordinates not equal")
	}
}

func TestQueueWait(t *testing.T) {
	rq := newRenderQueue(10)

	coord := gopnik.TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
	}
	resCh := make(chan *RenderPoolResponse, 1)

	go func() {
		err := rq.Push(coord, resCh)
		if err != nil {
			t.Errorf("Push error: %v", err)
		}
	}()

	coord2 := <-rq.TasksChan()
	if !coord.Equals(&coord2) {
		t.Error("Coordinates not equal")
	}
}

func TestQueueWaitMulti(t *testing.T) {
	rq := newRenderQueue(10)

	coordA := gopnik.TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
	}
	coordB := gopnik.TileCoord{
		X:    3,
		Y:    5,
		Zoom: 12,
	}
	coordC := gopnik.TileCoord{
		X:    3,
		Y:    5,
		Zoom: 12,
		Size: 4,
	}
	resCh := make(chan *RenderPoolResponse, 10)

	go func() {
		time.Sleep(1 * time.Millisecond)
		err := rq.Push(coordA, resCh)
		if err != nil {
			t.Errorf("Push error: %v", err)
		}
		time.Sleep(1 * time.Millisecond)
		err = rq.Push(coordB, resCh)
		if err != nil {
			t.Errorf("Push error: %v", err)
		}
		time.Sleep(1 * time.Millisecond)
		err = rq.Push(coordC, resCh)
		if err != nil {
			t.Errorf("Push error: %v", err)
		}
	}()

	coordA2 := <-rq.TasksChan()
	if !coordA.Equals(&coordA2) {
		t.Errorf("Coordinates not equal: %v != %v", coordA2, coordA)
	}

	coordB2 := <-rq.TasksChan()
	if !coordB.Equals(&coordB2) {
		t.Errorf("Coordinates not equal: %v != %v", coordB2, coordB)
	}

	coordC2 := <-rq.TasksChan()
	if !coordC.Equals(&coordC2) {
		t.Errorf("Coordinates not equal: %v != %v", coordC2, coordC)
	}
}

func TestQueueDub(t *testing.T) {
	rq := newRenderQueue(10)

	coordA := gopnik.TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
	}
	coordB := gopnik.TileCoord{
		X:    10,
		Y:    7,
		Zoom: 11,
	}
	resCh := make(chan *RenderPoolResponse, 1)

	go func() {
		err := rq.Push(coordA, resCh)
		if err != nil {
			t.Errorf("Push error: %v", err)
		}
		err = rq.Push(coordB, resCh)
		if err != nil {
			t.Errorf("Push error: %v", err)
		}
	}()

	coord2 := <-rq.TasksChan()
	if !coordA.Equals(&coord2) {
		t.Error("Coordinates not equal")
	}
}
