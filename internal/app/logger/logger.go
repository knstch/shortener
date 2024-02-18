// Модуль logger отвечает за логирование ошибок и debug.
package logger

import "go.uber.org/zap"

// ErrorLogger принимает комментарий в виде строки и ошибку,
// далее выводит сообщение об ошибки в терминал.
func ErrorLogger(msg string, serverErr error) {
	var logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	var sugar = *logger.Sugar()
	defer logger.Sync()
	sugar.Errorf("Error: %v\nDetails: %v\n", msg, serverErr)
}

// InfoLogger принимает сообщение в виде строки и выводит его в консоль.
func InfoLogger(msg string) {
	var logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	var sugar = *logger.Sugar()
	defer logger.Sync()
	sugar.Infof(msg)
}
