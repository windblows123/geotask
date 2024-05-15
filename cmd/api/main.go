package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"gitlab.com/iaroslavtsevaleksandr/geotask/run"
	"log"
	"os"
)

func main() {
	godotenv.Load()
	// инициализация приложения
	app := run.NewApp()
	// запуск приложения
	err := app.Run()
	// в случае ошибки выводим ее в лог и завершаем работу с кодом 2
	if err != nil {
		log.Println(fmt.Sprintf("error: %s", err))
		os.Exit(2)
	}
}
