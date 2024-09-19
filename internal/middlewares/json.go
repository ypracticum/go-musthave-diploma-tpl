package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type parsedJSONDataFieldType string

const parsedJSONDataField parsedJSONDataFieldType = "parsedJSONDataField"

type ModelParameter interface {
	interface{} | []interface{}
}

func JSONMiddleware[Model ModelParameter](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type is not application/json", http.StatusUnsupportedMediaType)
			return
		}

		var parsedData Model
		var buf bytes.Buffer

		if _, err := buf.ReadFrom(r.Body); err != nil {
			http.Error(w, fmt.Sprintf("Error occurred during reading from the body: %s", err.Error()), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &parsedData); err != nil {
			http.Error(w, fmt.Sprintf("Error occurred during unmarshaling data %s", err.Error()), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), parsedJSONDataField, parsedData)))
	})
}

func GetParsedJSONData[Model ModelParameter](w http.ResponseWriter, r *http.Request) Model {
	data, ok := r.Context().Value(parsedJSONDataField).(Model)

	if !ok {
		http.Error(w, "Could not retrieve data from context", http.StatusInternalServerError)
		var empty Model
		return empty
	}

	return data
}

func EncodeJSONResponse[Model interface{}](w http.ResponseWriter, data Model) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(data)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during encoding json response: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during writing response: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
