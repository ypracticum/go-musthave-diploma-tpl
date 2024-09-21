package router

import (
	"log"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/logger"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/middlewares"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	// Endpoint адрес и порт, на которых сервер будет слушать входящие запросы.
	Endpoint string
}

type Router struct {
	config         Config
	authService    models.AuthService
	jwtService     models.JWTService
	orderService   models.OrderService
	accrualService models.AccrualService
	balanceService models.BalanceService
}

// New создает новый экземпляр Router с заданными зависимостями.
func New(
	config Config,
	authService models.AuthService,
	jwtService models.JWTService,
	orderService models.OrderService,
	accrualService models.AccrualService,
	balanceService models.BalanceService,
) *Router {
	return &Router{
		config:         config,
		authService:    authService,
		jwtService:     jwtService,
		orderService:   orderService,
		accrualService: accrualService,
		balanceService: balanceService,
	}
}

// get возвращает настроенный роутер.
func (router *Router) get() chi.Router {
	r := chi.NewRouter()

	// Настройка промежуточного ПО (middleware) для роутера.
	r.Use(
		// Инжектор сервисов для предоставления сервисов в обработчиках.
		middlewares.ServiceInjectorMiddleware(
			router.authService,
			router.jwtService,
			router.orderService,
			router.accrualService,
			router.balanceService,
		),
		// Логгер для регистрации запросов.
		logger.RequestLogger,
		// Middleware для проверки аутентификации, исключая указанные пути.
		middlewares.AuthMiddleware().WithExcludedPaths(
			"/api/user/register",
			"/api/user/login",
		).Middleware,
	)

	// Настройка маршрутов для API пользователей.
	r.Route("/api/user", func(r chi.Router) {
		// Регистрация нового пользователя.
		r.With(middlewares.JSONMiddleware[models.UnknownUser]).Post("/register", Register)
		// Аутентификация пользователя.
		r.With(middlewares.JSONMiddleware[models.UnknownUser]).Post("/login", Login)

		// Создание заказа.
		r.With(middlewares.TextMiddleware).Post("/orders", CreateOrder)
		// Получение списка заказов.
		r.Get("/orders", GetOrders)

		// Получение баланса пользователя.
		r.Get("/balance", GetBalance)
		// Вывод средств.
		r.With(middlewares.JSONMiddleware[models.Withdrawal]).Post("/balance/withdraw", CreateWithdrawal)

		// Получение истории выводов.
		r.Get("/withdrawals", GetWithdrawals)
	})

	return r
}

// Run запускает HTTP сервер на заданном endpoint и начинает принимать запросы.
func (router *Router) Run() {
	log.Fatal(http.ListenAndServe(router.config.Endpoint, router.get()))
}
