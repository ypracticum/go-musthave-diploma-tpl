package main

import (
	"context"
	"log"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/database"
	router "github.com/Renal37/go-musthave-diploma-tpl/internal/http"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/logger"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/services"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/utils"
)

func main() {
	ctx := context.Background()
	config := NewConfig()

	if err := logger.Initialize(config.logLevel, config.env); err != nil {
		log.Fatalf("Logger wasn't initialized due to %s", err)
	}

	db, err := database.New(ctx, config.dsn)

	if err != nil {
		log.Fatalf("Database wasn't initialized due to %s", err)
	}

	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Migrations weren't run due to %s", err)
	}

	log.Printf("Running server on %s\n", config.endpoint)

	jobQueueService := services.NewJobQueueService(ctx, 100, 2)
	accrualService := services.NewAccrualService(db, jobQueueService, config.accrualEndpoint)

	if err := accrualService.StartCalculationAccruals(ctx); err != nil {
		log.Fatalf("Starting calculation accruals was failed due to %s", err)
	}

	utils.HandleTerminationProcess(func() {
		jobQueueService.Shutdown()
	})

	router.New(
		router.Config{Endpoint: config.endpoint},
		services.NewAuthService(db),
		services.NewJWTService(config.authSecretKey),
		services.NewOrderService(db),
		accrualService,
		services.NewBalanceService(db),
	).Run()
}
