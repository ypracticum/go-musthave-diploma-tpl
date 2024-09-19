package middlewares

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

type parsedTextDataFieldType string

const parsedTextDataField parsedTextDataFieldType = "parsedTextDataField"

func TextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Content-Type is not text/plain", http.StatusUnsupportedMediaType)
			return
		}

		var buf bytes.Buffer

		if _, err := buf.ReadFrom(r.Body); err != nil {
			http.Error(w, fmt.Sprintf("Error occurred during reading from the body: %s", err.Error()), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), parsedTextDataField, buf.String())))
	})
}

func GetParsedTextData(w http.ResponseWriter, r *http.Request) string {
	data, ok := r.Context().Value(parsedTextDataField).(string)

	if !ok {
		http.Error(w, "Could not retrieve data from context", http.StatusInternalServerError)
		return ""
	}

	return data
}
