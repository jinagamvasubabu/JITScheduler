package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jinagamvasubabu/JITScheduler-svc/adapters/logger"
	db "github.com/jinagamvasubabu/JITScheduler-svc/adapters/persistence"
	"github.com/jinagamvasubabu/JITScheduler-svc/config"
	handler "github.com/jinagamvasubabu/JITScheduler-svc/handler"
	"github.com/jinagamvasubabu/JITScheduler-svc/repository"
	"github.com/jinagamvasubabu/JITScheduler-svc/service"

	"github.com/google/uuid"

	cronFunc "github.com/jinagamvasubabu/JITScheduler-svc/service/kafka/cron"
	"github.com/jinagamvasubabu/JITScheduler-svc/service/kafka/delayer"
	"github.com/robfig/cron"
	"go.uber.org/zap"
)

func main() {
	defer recoverPanic()
	ctx := context.Background()
	//Initialize the config
	//Logger
	logger.InitLogger()
	if err := config.InitConfig(); err != nil {
		logger.Error("error while loading the config", zap.Error(err))
	}

	DB, err := db.InitPostgresDatabase()
	if err != nil {
		logger.Error("Error", zap.Error(err))
	}

	//Initialise the repositories
	tenantRepository := repository.NewTenantRepository(ctx, DB)
	eventRepository := repository.NewEventRepository(ctx, DB)

	//Initiliaze the  services
	tenantService := service.NewTenantService(ctx, tenantRepository)
	eventService := service.NewEventService(ctx, eventRepository)

	//Initiliaze the handler
	tenantHandler := handler.NewTenantHandler(tenantService)
	eventHandler := handler.NewEventHandler(eventService)

	//router
	router := handler.InitRouter(tenantHandler, eventHandler)
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.GetConfig().HOST, config.GetConfig().PORT),
		Handler: router,
	}
	//redis
	db.NewRedisClient()

	//Scheduler
	scheduler := delayer.NewScheduler(context.Background(), eventRepository, tenantRepository)
	scheduler.Start(context.Background())
	go scheduler.Wait()

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)

	//Graceful shutdown on OS signals (CTRL+C, etc)
	go func() {
		<-termChan // Blocks here until interrupted
		logger.Info("SIGTERM received. Shutdown process initiated\n")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		//delete the cron locks
		db.GetRedisClient().Del("scheduler_cron_lock", "rescheduler_cron_lock", "monitor_cron_lock")
		srv.Shutdown(ctx)
	}()
	logger.Info("Server started", zap.String("address", srv.Addr))

	//Run Cron for scheduling events
	go runCronJobForSchedulingEvents(uuid.New().String(), eventRepository)

	//Start the server
	if err := srv.ListenAndServe(); err != nil {
		logger.Info("HTTP server interupted, Error", zap.Error(err))
	}
	logger.Info("Server Stopped")
}

// function to recover a panic
func recoverPanic() {
	if r := recover(); r != nil {
		logger.Info("Recovered from panic!!!")
	}
}

func runCronJobForSchedulingEvents(instanceId string, eventRepository repository.EventRepository) {
	c := cron.New()
	conf := config.GetConfig()
	scheduleEvents := cronFunc.NewScheduleEventsCron(instanceId, eventRepository)
	rescheduleFailedEvents := cronFunc.NewRescheduleFailedEvents(instanceId, eventRepository)
	monitorCron := cronFunc.NewMonitorCron(eventRepository)
	scheduleEventsCronSpec := fmt.Sprintf("@every %s", conf.ReschedulerCronSpec)
	_ = c.AddFunc(scheduleEventsCronSpec, scheduleEvents.Run)
	rescheduleCronSpec := fmt.Sprintf("@every %s", conf.ReschedulerCronSpec)
	_ = c.AddFunc(rescheduleCronSpec, rescheduleFailedEvents.Run)
	monitorCronSpec := fmt.Sprintf("@every %s", conf.MonitorCronSpec)
	_ = c.AddFunc(monitorCronSpec, monitorCron.MonitorQueuedEvents)
	go c.Start()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}
