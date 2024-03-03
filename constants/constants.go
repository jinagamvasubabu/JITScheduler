package constants

import "time"

const (
	SchedulerCronTTL                  = 10 * time.Minute
	SchedulerCron                     = "scheduler_cron_lock"
	RescheduleCron                    = "rescheduler_cron_lock"
	ReschedulerCronTTL                = 10 * time.Second
	MonitorCron                       = "monitor_cron_lock"
	MonitorCronTTL                    = 5 * time.Minute
	SchedulerEventsWithInNextXMinutes = 30
)
