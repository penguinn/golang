package cron

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/robfig/cron"
)

var (
	cEnginer = cron.New()

	tasks sync.Map

	ErrTaskNotFound  = errors.New("task not found")
	ErrDuplicateTask = errors.New("duplicate task")
	ErrTaskInCron    = errors.New("task is in cron")

	config Config
)

type CronTaskItem struct {
	Name string
	RC   string
}

type Config []CronTaskItem

type Cpnt struct {
}

func (Cpnt) Init(options ...interface{}) error {
	cEnginer.Start()

	if len(options) == 0 {
		return nil
	}
	c, ok := options[0].(*Config)
	if !ok {
		return nil
	}

	config = *c

	return nil
}

func (Cpnt) CfgKey() string {
	return "cron"
}
func (Cpnt) CfgType() interface{} {
	return Config{}
}
func (Cpnt) CfgUpdate(interface{}) {

}

func Start() error {

	for _, tc := range config {

		val, ok := tasks.Load(tc.Name)
		if !ok {
			return ErrTaskNotFound
		}
		t := val.(*Task)
		if t.inCron {
			continue
		}

		if tc.RC != "" {
			if err := t.StartWithRC(tc.RC); err != nil {
				return err
			}
		} else {
			if err := t.Start(); err != nil {
				return err
			}
		}

	}
	return nil
}

func StopTask(name string) error {

	val, ok := tasks.Load(name)
	if !ok {
		return ErrTaskNotFound
	}
	t := val.(*Task)
	t.Stop()
	return nil
}

func StartTask(name string) error {

	val, ok := tasks.Load(name)
	if !ok {
		return ErrTaskNotFound
	}

	return val.(*Task).Start()
}

func StartTaskWithRC(name, rc string) error {
	val, ok := tasks.Load(name)
	if !ok {
		return ErrTaskNotFound
	}

	return val.(*Task).StartWithRC(rc)
}

func RegisterTask(m mode, f func()) string {

	t := newTask(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), m, f, cEnginer)

	tasks.Store(t.name, t)
	fmt.Println("[cron] Register Task : ", t.name)
	return t.name
}
