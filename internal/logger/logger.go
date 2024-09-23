package logger

import (
	"net/http"
	"time"
	"fmt"
	"go.uber.org/zap"
)

// Log глобальный логгер, инициализируется функцией Initialize.
// По умолчанию используется заглушка zap.NewNop(), которая не выводит никаких логов.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует логгер с заданным уровнем логирования и средой выполнения.
// Параметры:
// - level: уровень логирования (например, "debug", "info", "warn", "error").
// - env: среда выполнения ("development" или "production").
func Initialize(level, env string) error {
	// Парсинг уровня логирования.
	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("ошибка парсинга уровня логирования: %w", err)
	}

	var config zap.Config

	// Выбор конфигурации логгера в зависимости от среды выполнения.
	if env == "development" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	// Установка уровня логирования.
	config.Level = logLevel

	// Построение логгера на основе конфигурации.
	logger, err := config.Build()
	if err != nil {
		return fmt.Errorf("ошибка построения логгера: %w", err)
	}

	// Присваиваем глобальной переменной Log инициализированный логгер.
	Log = logger

	return nil
}

// responseWriter оборачивает http.ResponseWriter и сохраняет код статуса ответа.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter создает новый экземпляр responseWriter с кодом статуса по умолчанию (200 OK).
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader сохраняет код статуса ответа и вызывает метод WriteHeader у встроенного ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger является middleware, которое логирует информацию о каждом HTTP-запросе.
// Логируются URI, метод запроса, длительность обработки и код статуса ответа.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		wrappedWriter := newResponseWriter(w)

		// Обработка запроса следующим обработчиком.
		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(startTime)

		// Логирование информации о запросе.
		Log.Info("Запрос обработан",
			zap.String("URI", r.RequestURI),
			zap.String("метод", r.Method),
			zap.Duration("длительность", duration),
			zap.Int("статус", wrappedWriter.statusCode),
		)
	})
}
