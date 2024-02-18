// Пакет logger используется для логирования взаимодействия с сервером.
package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponse struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write - это модифицированный метод интерфейса http.ResponseWriter.
func (r *loggingResponse) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader - это модифицированный метод интерфейса http.ResponseWriter.
func (r *loggingResponse) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// RequestsLogger - это middlware обработчик для запросов, записывает URI, method, duration.
func RequestsLogger(h http.Handler) http.Handler {
	var logger, err = zap.NewDevelopment()
	var sugar = *logger.Sugar()
	if err != nil {
		panic(err)
	}

	logFn := func(res http.ResponseWriter, req *http.Request) {

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		loggingRes := loggingResponse{
			ResponseWriter: res,
			responseData:   responseData,
		}

		start := time.Now()

		uri := req.RequestURI

		method := req.Method

		h.ServeHTTP(&loggingRes, req)

		duration := time.Since(start)

		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"status code", responseData.status,
			"size", responseData.size,
		)
	}
	defer logger.Sync()
	return http.HandlerFunc(logFn)
}
