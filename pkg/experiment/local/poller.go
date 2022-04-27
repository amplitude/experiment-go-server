package local

import "time"

type poller struct {
	shutdown chan bool
}

func newPoller() *poller {
	return &poller{
		shutdown: make(chan bool),
	}
}

func (p *poller) Poll(interval time.Duration, function func()) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-p.shutdown:
				ticker.Stop()
				return
			case _ = <-ticker.C:
				go function()
			}
		}
	}()
}
