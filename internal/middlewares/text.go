package middlewares

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// parsedTextDataFieldType определяет тип для ключа, используемого для хранения текстовых данных в контексте.
type parsedTextDataFieldType string

// parsedTextDataField используется для хранения текстовых данных в контексте запроса.
const parsedTextDataField parsedTextDataFieldType = "parsedTextDataField"

// TextMiddleware представляет middleware для обработки текстовых данных.
// Middleware проверяет, что Content-Type запроса - "text/plain" и сохраняет текстовые данные в контексте.
func TextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что заголовок Content-Type соответствует "text/plain".
		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Заголовок Content-Type не соответствует text/plain", http.StatusUnsupportedMediaType)
			return
		}

		var buf bytes.Buffer

		// Читаем данные из тела запроса.
		if _, err := buf.ReadFrom(r.Body); err != nil {
			http.Error(w, fmt.Sprintf("Произошла ошибка при чтении данных из тела запроса: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Добавляем прочитанные данные в контекст и передаем управление следующему обработчику.
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), parsedTextDataField, buf.String())))
	})
}

// GetParsedTextData извлекает текстовые данные из контекста запроса.
// В случае ошибки возвращает HTTP 500 и пустую строку.
func GetParsedTextData(w http.ResponseWriter, r *http.Request) string {
	data, ok := r.Context().Value(parsedTextDataField).(string)

	if !ok {
		http.Error(w, "Не удалось извлечь данные из контекста", http.StatusInternalServerError)
		return ""
	}

	return data
}
