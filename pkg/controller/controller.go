package controller

import "time"

type Controller interface {
	Run() error
	Stop() error
}
type Queue struct{}

type Timer interface {
	OnTimer(t <-chan struct{})
}

func (q *Queue) Schedule(t Timer, duration time.Duration) {
	after := time.After(duration)
	<-after
	sig := make(chan struct{})
	defer func() { close(sig) }()
	go t.OnTimer(sig)
	sig <- struct{}{}
}
