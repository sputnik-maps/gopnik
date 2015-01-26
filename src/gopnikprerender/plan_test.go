package main

import (
	"time"

	"gopnik"

	. "gopkg.in/check.v1"

	json "github.com/orofarne/strict-json"
)

type PlanSuite struct {
	bboxes []gopnik.TileCoord
}

func (s *PlanSuite) SetUpSuite(c *C) {
	err := json.Unmarshal([]byte(planJson), &s.bboxes)
	if err != nil {
		panic(err)
	}
}

var _ = Suite(&PlanSuite{})

func (s *PlanSuite) TestSimpleGet(c *C) {
	plan := newPlan(s.bboxes)
	c.Assert(plan, NotNil)

	for i := 0; i < len(s.bboxes); i++ {
		task := plan.GetTask()
		c.Check(task, NotNil)
		plan.DoneTask(*task)
	}
	lastTask := plan.GetTask()
	c.Check(lastTask, IsNil)
	c.Check(plan.DoneTasks(), Equals, plan.TotalTasks())
}

func (s *PlanSuite) Test50pFails(c *C) {
	plan := newPlan(s.bboxes)
	c.Assert(plan, NotNil)

	for i := 0; i < len(s.bboxes); i++ {
		task := plan.GetTask()
		c.Check(task, NotNil)
		if i%2 == 0 {
			plan.DoneTask(*task)
		} else {
			plan.FailTask(*task)
		}
	}
	for i := 0; i < len(s.bboxes)/2; i++ {
		task := plan.GetTask()
		c.Check(task, NotNil)
		plan.DoneTask(*task)
	}
	lastTask := plan.GetTask()
	c.Check(lastTask, IsNil)
	c.Check(plan.DoneTasks(), Equals, plan.TotalTasks())
}

func (s *PlanSuite) TestWait(c *C) {
	plan := newPlan(s.bboxes)
	c.Assert(plan, NotNil)

	for i := 0; i < len(s.bboxes)-1; i++ {
		task := plan.GetTask()
		c.Check(task, NotNil)
		plan.DoneTask(*task)
	}
	task := plan.GetTask()
	c.Check(task, NotNil)
	c.Check(plan.DoneTasks(), Equals, plan.TotalTasks()-1)
	flag := false
	go func() {
		time.Sleep(time.Millisecond)
		flag = true
		plan.DoneTask(*task)
	}()
	lastTask := plan.GetTask()
	c.Check(lastTask, IsNil)
	c.Check(flag, Equals, true)
	c.Check(plan.DoneTasks(), Equals, plan.TotalTasks())
}

const planJson = `[{"X":0,"Y":0,"Zoom":3,"Size":8,"Tags":null},{"X":8,"Y":0,"Zoom":3,"Size":8,"Tags":null},{"X":0,"Y":0,"Zoom":4,"Size":8,"Tags":null},{"X":0,"Y":8,"Zoom":4,"Size":8,"Tags":null},{"X":8,"Y":0,"Zoom":4,"Size":8,"Tags":null},{"X":8,"Y":8,"Zoom":4,"Size":8,"Tags":null},{"X":16,"Y":0,"Zoom":4,"Size":8,"Tags":null},{"X":16,"Y":8,"Zoom":4,"Size":8,"Tags":null},{"X":0,"Y":0,"Zoom":5,"Size":8,"Tags":null},{"X":0,"Y":8,"Zoom":5,"Size":8,"Tags":null},{"X":0,"Y":16,"Zoom":5,"Size":8,"Tags":null},{"X":0,"Y":24,"Zoom":5,"Size":8,"Tags":null},{"X":8,"Y":0,"Zoom":5,"Size":8,"Tags":null},{"X":8,"Y":8,"Zoom":5,"Size":8,"Tags":null},{"X":8,"Y":16,"Zoom":5,"Size":8,"Tags":null},{"X":8,"Y":24,"Zoom":5,"Size":8,"Tags":null},{"X":16,"Y":0,"Zoom":5,"Size":8,"Tags":null},{"X":16,"Y":8,"Zoom":5,"Size":8,"Tags":null},{"X":16,"Y":16,"Zoom":5,"Size":8,"Tags":null},{"X":16,"Y":24,"Zoom":5,"Size":8,"Tags":null},{"X":24,"Y":0,"Zoom":5,"Size":8,"Tags":null},{"X":24,"Y":8,"Zoom":5,"Size":8,"Tags":null},{"X":24,"Y":16,"Zoom":5,"Size":8,"Tags":null},{"X":24,"Y":24,"Zoom":5,"Size":8,"Tags":null},{"X":32,"Y":0,"Zoom":5,"Size":8,"Tags":null},{"X":32,"Y":8,"Zoom":5,"Size":8,"Tags":null},{"X":32,"Y":16,"Zoom":5,"Size":8,"Tags":null},{"X":32,"Y":24,"Zoom":5,"Size":8,"Tags":null},{"X":0,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":0,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":0,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":0,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":0,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":0,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":0,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":0,"Y":56,"Zoom":6,"Size":8,"Tags":null},{"X":8,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":8,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":8,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":8,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":8,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":8,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":8,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":8,"Y":56,"Zoom":6,"Size":8,"Tags":null},{"X":16,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":16,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":16,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":16,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":16,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":16,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":16,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":16,"Y":56,"Zoom":6,"Size":8,"Tags":null},{"X":24,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":24,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":24,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":24,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":24,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":24,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":24,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":24,"Y":56,"Zoom":6,"Size":8,"Tags":null},{"X":32,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":32,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":32,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":32,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":32,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":32,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":32,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":32,"Y":56,"Zoom":6,"Size":8,"Tags":null},{"X":40,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":40,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":40,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":40,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":40,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":40,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":40,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":40,"Y":56,"Zoom":6,"Size":8,"Tags":null},{"X":48,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":48,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":48,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":48,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":48,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":48,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":48,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":48,"Y":56,"Zoom":6,"Size":8,"Tags":null},{"X":56,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":56,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":56,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":56,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":56,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":56,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":56,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":56,"Y":56,"Zoom":6,"Size":8,"Tags":null},{"X":64,"Y":0,"Zoom":6,"Size":8,"Tags":null},{"X":64,"Y":8,"Zoom":6,"Size":8,"Tags":null},{"X":64,"Y":16,"Zoom":6,"Size":8,"Tags":null},{"X":64,"Y":24,"Zoom":6,"Size":8,"Tags":null},{"X":64,"Y":32,"Zoom":6,"Size":8,"Tags":null},{"X":64,"Y":40,"Zoom":6,"Size":8,"Tags":null},{"X":64,"Y":48,"Zoom":6,"Size":8,"Tags":null},{"X":64,"Y":56,"Zoom":6,"Size":8,"Tags":null}]`
