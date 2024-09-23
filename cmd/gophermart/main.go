package main

import (
	"context"
	"log"

	// Импорт внутренних пакетов приложения
	router "github.com/Renal37/go-musthave-diploma-tpl/internal/app"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/database"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/logger"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/services"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/utils"
)


func main() {
	// Создание контекста приложения
	ctx := context.Background()

	// Загрузка конфигурации
	config := NewConfig()

	// Инициализация логгера с уровнем логирования и окружением из конфигурации
	if err := logger.Initialize(config.logLevel, config.env); err != nil {
		log.Fatalf("Ошибка инициализации логгера: %s", err)
	}

	// Подключение к базе данных с использованием DSN из конфигурации
	db, err := database.New(ctx, config.dsn)
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %s", err)
	}

	// Запуск миграций базы данных
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Ошибка при запуске миграций базы данных: %s", err)
	}

	log.Printf("Сервер запущен на %s\n", config.endpoint)

	// Создание сервиса очереди задач с ограничением на 100 задач и 5 воркерами
	jobQueueService := services.NewJobQueueService(ctx, 100, 5)

	// Создание сервиса начислений с подключением к базе данных и сервису очереди задач
	accrualService := services.NewAccrualService(db, jobQueueService, config.accrualEndpoint)

	// Запуск процесса расчета начислений
	if err := accrualService.StartCalculationAccruals(ctx); err != nil {
		log.Fatalf("Ошибка при запуске расчета начислений: %s", err)
	}

	// Обработка сигналов завершения процесса (например, SIGINT, SIGTERM)
	utils.HandleTerminationProcess(func() {
		// Корректное завершение работы очереди задач
		jobQueueService.Shutdown()
	})

	// Создание и запуск роутера с необходимыми сервисами
	router.New(
		router.Config{Endpoint: config.endpoint},
		services.NewAuthService(db),
		services.NewJWTService(config.authSecretKey),
		services.NewOrderService(db),
		accrualService,
		services.NewBalanceService(db),
	).Run()
}
