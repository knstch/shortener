// Модуль logger отвечает за логирование ошибок и debug.
package logger

import "go.uber.org/zap"

// ErrorLogger принимает комментарий в виде строки и ошибку,
// далее выводит сообщение об ошибки в терминал.
func ErrorLogger(msg string, serverErr error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	var sugar = *logger.Sugar()
	sugar.Errorf("Error: %v\nDetails: %v\n", msg, serverErr)
	err = logger.Sync()
	if err != nil {
		panic(err)
	}
}

// InfoLogger принимает сообщение в виде строки и выводит его в консоль.
func InfoLogger(msg string) {
	var logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	var sugar = *logger.Sugar()
	sugar.Infof(msg)
	err = logger.Sync()
	if err != nil {
		panic(err)
	}
}
