package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// parsedJSONDataFieldType является типом для хранения данных JSON в контексте запроса.
type parsedJSONDataFieldType string

// parsedJSONDataField - ключ для хранения данных JSON в контексте запроса.
const parsedJSONDataField parsedJSONDataFieldType = "parsedJSONDataField"

// ModelParameter определяет интерфейс, который могут реализовывать модели, поддерживающие как одиночные значения, так и срезы значений.
type ModelParameter interface {
	interface{} | []interface{}
}

// JSONMiddleware обрабатывает JSON-запросы и извлекает данные JSON из тела запроса.
func JSONMiddleware[Model ModelParameter](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверка заголовка Content-Type.
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Тип контента не является application/json", http.StatusUnsupportedMediaType)
			return
		}

		var parsedData Model
		var buf bytes.Buffer

		// Чтение данных из тела запроса.
		if _, err := buf.ReadFrom(r.Body); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка чтения из тела запроса: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Распаковка данных JSON в структуру.
		if err := json.Unmarshal(buf.Bytes(), &parsedData); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при разборе данных JSON: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Передача данных JSON в контексте запроса.
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), parsedJSONDataField, parsedData)))
	})
}

// GetParsedJSONData извлекает данные JSON из контекста запроса.
func GetParsedJSONData[Model ModelParameter](w http.ResponseWriter, r *http.Request) Model {
	data, ok := r.Context().Value(parsedJSONDataField).(Model)

	if !ok {
		// Возврат ошибки, если данные не найдены в контексте.
		http.Error(w, "Не удалось извлечь данные из контекста", http.StatusInternalServerError)
		var empty Model
		return empty
	}

	return data
}

// EncodeJSONResponse кодирует данные в формат JSON и отправляет их в ответе.
func EncodeJSONResponse[Model any](w http.ResponseWriter, data Model) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(data)
	if err != nil {
		// Возврат ошибки, если не удалось закодировать данные.
		http.Error(w, fmt.Sprintf("Ошибка при кодировании JSON-ответа: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		// Возврат ошибки, если не удалось отправить ответ.
		http.Error(w, fmt.Sprintf("Ошибка при отправке ответа: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Установка статуса ответа в OK.
	w.WriteHeader(http.StatusOK)
}
