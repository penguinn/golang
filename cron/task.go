package cron

import (
	"errors"
	"sync"

	"github.com/robfig/cron"
)

/**

运行模式

normal     正常模式 如果上一次执行没有完，当前直接退出
waiting    排队模式 如果上一次执行没有完，当前排队等待上次执行
waitingone 排队模式 如果上一次执行没有完，当前排队等待上次执行，但只排队一个
parallel   并行模式 如果上一次执行没有完，当前不管上次执行情况

*/

const (
	ModeNormal mode = iota
	ModeWaiting
	ModeWaitingOne
	ModeParallel
)

const (
	runRoutine byte = iota
	runNoRoutine
)

const (
	DefaultRC = "* * * * * *"
)

type mode byte

type Task struct {
	name   string
	f      func()
	m      mode
	runCfg string
	inCron bool

	cronEnginer *cron.Cron

	run     bool
	runChan chan byte

	mu sync.Mutex

	successTimes uint64
	panicTimes   uint64
	runNum       uint64
}

func newTask(name string, m mode, f func(), ce *cron.Cron) *Task {
	t := Task{
		name:        name,
		f:           f,
		m:           m,
		runCfg:      DefaultRC,
		run:         false,
		cronEnginer: ce,
	}
	t.runChan = make(chan byte)
	return &t
}

func (t *Task) listen() {

	for {
		if !t.run {
			return
		}
		select {
		case rc := <-t.runChan:

			switch rc {
			case runRoutine:
				go t.runFP()
			case runNoRoutine:
				t.runFP()
			}
		}
	}

}

func (t *Task) Start() error {
	if t.run {
		return nil
	}

	if !t.inCron {
		if err := t.cronEnginer.AddJob(t.runCfg, t); err != nil {
			return err
		}
	}
	t.run = true
	t.inCron = true

	go t.listen()
	return nil
}

func (t *Task) StartWithRC(rc string) error {
	if t.run {
		return errors.New("task is running")
	}

	if t.inCron && rc != t.runCfg {
		return errors.New("task run config could not change")
	}
	if !t.inCron {
		t.runCfg = rc
		if err := t.cronEnginer.AddJob(t.runCfg, t); err != nil {
			return err
		}
	}

	t.inCron = true
	t.run = true
	go t.listen()
	return nil
}

func (t *Task) Stop() {
	if !t.run {
		return
	}
	t.run = false
}

func (t *Task) runFP() {
	defer func() {
		if rcv := recover(); rcv != nil {
			t.mu.Lock()
			t.panicTimes++
			t.mu.Unlock()
		}
	}()
	if !t.run {
		return
	}

	t.f()
	t.mu.Lock()
	t.successTimes++
	t.mu.Unlock()
}

func (t *Task) put(b byte) {

	go func() {
		t.mu.Lock()
		t.runNum++
		t.mu.Unlock()
		t.runChan <- b
	}()
}

func (t *Task) Run() {

	switch t.m {
	case ModeNormal:
		t.mu.Lock()
		if t.runNum == t.successTimes+t.panicTimes {
			t.put(runNoRoutine)
		}
		t.mu.Unlock()

	case ModeWaiting:
		t.put(runNoRoutine)
	case ModeWaitingOne:
		t.mu.Lock()
		if t.runNum < t.successTimes+t.panicTimes+2 {
			t.put(runNoRoutine)
		}
		t.mu.Unlock()
	case ModeParallel:
		t.put(runRoutine)

	}

}
