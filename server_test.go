package go_cron

import "testing"

func TestRun(t *testing.T) {
	cs := NewCronServer()
	err := cs.Start()
	if err != nil {
		return
	}
}
