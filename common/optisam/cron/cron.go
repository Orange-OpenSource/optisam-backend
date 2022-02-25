package cron

import (
	"log"

	"github.com/robfig/cron/v3"
)

// Add other params that is required to use in cron JOB
type Config struct {
	Time string
}

var cfg Config

func ConfigInit(c Config) {
	cfg = c
}

// AddCronJob initiate the cron job
func AddCronJob(fp func()) {
	cronOb := cron.New(cron.WithLogger(cron.DefaultLogger))
	cronOb.Start()
	log.Println("starting cron job per ", cfg.Time)
	cronOb.AddFunc(cfg.Time, fp)
}
