package router

import (
	"log"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/logger"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/middlewares"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/go-musthave-diploma-tplchi/chi/v5"
)

type Config struct {
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

func New(
	config Config,
	authService models.AuthService,
	jwtService models.JWTService,
	orderService models.OrderService,
	accrualService models.AccrualService,
	balanceService models.BalanceService,
) *Router {
	return &Router{
		config,
		authService,
		jwtService,
		orderService,
		accrualService,
		balanceService,
	}
}

func (router *Router) get() chi.Router {
	r := chi.NewRouter()

	r.Use(
		middlewares.ServiceInjectorMiddleware(
			router.authService,
			router.jwtService,
			router.orderService,
			router.accrualService,
			router.balanceService,
		),
		logger.RequestLogger,
		middlewares.AuthMiddleware().WithExcludedPaths(
			"/api/user/register",
			"/api/user/login",
		).Middleware,
	)

	r.Route("/api/user", func(r chi.Router) {
		r.With(middlewares.JSONMiddleware[models.UnknownUser]).Post("/register", Register)
		r.With(middlewares.JSONMiddleware[models.UnknownUser]).Post("/login", Login)

		r.With(middlewares.TextMiddleware).Post("/orders", CreateOrder)
		r.Get("/orders", GetOrders)

		r.Get("/balance", GetBalance)
		r.With(middlewares.JSONMiddleware[models.Withdrawal]).Post("/balance/withdraw", CreateWithdrawal)

		r.Get("/withdrawals", GetWithdrawals)
	})

	return r
}

func (router *Router) Run() {
	log.Fatal(http.ListenAndServe(router.config.Endpoint, router.get()))
}
