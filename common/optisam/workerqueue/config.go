package workerqueue

import "time"

type QueueConfig struct {
	Qsize        int
	PollingRate  time.Duration
	Retries      int
	BaseDelay    time.Duration
	IsMultiQueue bool
}
