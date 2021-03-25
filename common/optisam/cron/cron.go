// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package cron

import (
	"log"

	"github.com/robfig/cron/v3"
)

//Add other params that is required to use in cron JOB
type Config struct {
	Time string
}

var cfg Config

func CronConfigInit(c Config) {
	cfg = c
}

//AddCronJob initiate the cron job
func AddCronJob(fp func()) {
	cronOb := cron.New(cron.WithLogger(cron.DefaultLogger))
	cronOb.Start()
	log.Println("starting cron job per ", cfg.Time)
	cronOb.AddFunc(cfg.Time, fp)
}
