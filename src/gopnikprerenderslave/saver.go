package main

import (
	"fmt"
	"time"

	"gopnik"
	"tilerender"
)

type saver struct {
	cache gopnik.CachePluginInterface
	stop  chan int
}

func newSaver(cache gopnik.CachePluginInterface) *saver {
	return &saver{cache: cache, stop: make(chan int)}
}

func (s *saver) Worker(tiles <-chan *tilerender.RenderPoolResponse, done chan<- tileReport) {
L:
	for {
		τ0 := time.Now() // for Stats.SaverWaitT

		select {
		case <-s.stop:
			break L
		case t := <-tiles:
			Stats.SaverWaitT.AddPoint(time.Since(τ0).Seconds())

			// Statistic for queue length
			Stats.SaverQueue.AddPoint(float64(len(tiles)))

			if t.Error != nil {
				done <- tileReport{
					Coord:      t.Coord,
					RenderTime: t.RenderTime,
					Error:      t.Error,
				}
				continue
			}

			beginTime := time.Now()
			err := s.cache.Set(t.Coord, t.Tiles)
			if err != nil {
				done <- tileReport{
					Coord:      t.Coord,
					RenderTime: t.RenderTime,
					Error:      fmt.Errorf("Save error: %v", err),
				}
				continue
			}

			report := tileReport{
				Coord:      t.Coord,
				RenderTime: t.RenderTime,
				SaveTime:   time.Since(beginTime),
			}

			// Copy stats to hmetrics
			Stats.RenderT.AddPoint(report.RenderTime.Seconds())
			Stats.SaverT.AddPoint(report.SaveTime.Seconds())
			Stats.RenderWaitT.AddPoint(t.WaitTime.Seconds())

			done <- report
		}
	}
}

func (s *saver) Stop() {
	s.stop <- 1
	close(s.stop)
}
