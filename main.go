package main

import (
	"ecommerce/app"
	"ecommerce/components/log"
	module "ecommerce/modules"
	"ecommerce/repository"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
	}

	logger := log.GetLogger()
	logger.Print("Checking DB connection using env variables...")

	db := app.NewDBConnection(*logger)
	if db == nil {
		logger.Print("DB connection failed.")
		return
	}
	defer func() {
		sqlDB := db.DB()
		sqlDB.Close()
		logger.Print("DB Closed.")
	}()

	repo := repository.NewGormRespository()

	myApp := app.NewApp("Ecommerce App", db, *logger, repo)
	myApp.Init()

	module.Configure(myApp)

	module.RegisterModuleRoutes(myApp, repo)

	go func() {
		if err := myApp.StartServer(); err != nil {
			stopApp(myApp)
		}
	}()

	ch := make(chan os.Signal, 1) //here we make a channel to receive os signal , buffer is 1 to prevent missing the first signal if nobody is reading yet.
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch //This is a blocking read from the channel, the program will pause here until a signal is received.
	stopApp(myApp)
}

func stopApp(myApp *app.App) {
	myApp.Stop()
	log.GetLogger().Print("App Stopped")
	os.Exit(0)
}
