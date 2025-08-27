package vbasedata

import (
	"time"

	"github.com/alitto/pond"
	"github.com/go-kratos/kratos/v2/log"
)

type PondConfig struct {
	MaxWorkers  int `yaml:"max_workers" json:"max_workers"`
	MinWorkers  int `yaml:"min_workers" json:"min_workers"`
	MaxCapacity int `yaml:"max_capacity" json:"max_capacity"`
	StopAndWait int `yaml:"stop_and_wait" json:"stop_and_wait"`
}

type Pond struct {
	pond        *pond.WorkerPool
	stopAndWait int
}

func NewPond(c *PondConfig, log *log.Helper) *Pond {
	if c.MinWorkers == 0 {
		c.MinWorkers = 2
	}

	if c.MaxWorkers == 0 {
		c.MaxWorkers = 10
	}

	if c.StopAndWait == 0 {
		c.StopAndWait = 5
	}

	return &Pond{
		pond: pond.New(c.MaxWorkers, c.MaxCapacity, pond.MinWorkers(c.MinWorkers), pond.PanicHandler(func(i interface{}) {
			log.Errorf("当前Pond运行的任务异常退出:%v", i)
		})),
		stopAndWait: c.StopAndWait,
	}

}

func (t *Pond) Submit(f func()) {
	t.pond.Submit(f)
}

func (t *Pond) GetPond() *pond.WorkerPool {
	return t.pond
}

func (t *Pond) Stop() {
	t.pond.StopAndWaitFor(time.Duration(t.stopAndWait) * time.Second)
}
