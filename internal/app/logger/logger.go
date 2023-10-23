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

// Модификация интерфейса Write, добавляем сохрание размера в переменную
func (r *loggingResponse) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// Модификация интерфейса WriteHeader, добавляем сохрание статус кода в переменную
func (r *loggingResponse) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func ServerShutDownLog(serverErr error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := *logger.Sugar()
	sugar.Errorf("HTTP server shut down: %v", serverErr)
}

func ServerRuns(serverErr error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := *logger.Sugar()
	sugar.Fatalf("HTTP server ListenAndServe: %v", serverErr)
}

// Middlware обработчик для запросов, записывает URI, method, duration
func RequestsLogger(h http.HandlerFunc) http.HandlerFunc {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := *logger.Sugar()

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
	return http.HandlerFunc(logFn)
}