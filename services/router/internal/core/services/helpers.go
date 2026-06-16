package services

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// logSourceIssueThrottled пишет предупреждение не чаще чем раз в sourceWaitLogIntervalSec.
func logSourceIssueThrottled(last *time.Time, lg zerolog.Logger, msg string, kv ...any) {
	interval := time.Duration(sourceWaitLogIntervalSec) * time.Second
	if time.Since(*last) < interval {
		return
	}
	*last = time.Now()
	e := lg.Warn()
	for i := 0; i+1 < len(kv); i += 2 {
		k, _ := kv[i].(string)
		e = e.Interface(k, kv[i+1])
	}
	e.Msg(msg)
}

// sleepBackoff ждёт d или отмену ctx.
func sleepBackoff(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
