package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"uniback/controller"
	"uniback/repository/postgres"
	"uniback/utils"
)

func main() {
	logger := utils.GlobalLogger().SetLevel(utils.Debug)

	logger.Info("Start APP!!!")
	defer logger.Info("APP Done!!!")

	cfg := utils.CfgLoad("UniBack")
	logger.Info("Set app name: %s", cfg.AppName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	DataBase := postgres.New(ctx, postgres.PgConfigFromConfig(*cfg))

	if DataBase == nil {
		logger.Critical("Database error")
		return
	}
	defer DataBase.Close()

	authController := controller.NewAuthController(DataBase, cfg.JwtKey)
	http.HandleFunc("/register", authController.RegistrationHandler)
	http.HandleFunc("/login", authController.LoginHandler)
	//
	http.HandleFunc("/accounts", authController.AuthMiddleware(authController.AccountsHandler))

	server := &http.Server{Addr: cfg.HostAddress}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-quit:
			server.Shutdown(ctx)
			logger.Info("Signal to escape! Shutdown")
		case <-ctx.Done():
			logger.Error("Context DONE! It wouldn't be happened!!!")
		}
	}()

	logger.Info("Try to start server...")
	err := server.ListenAndServe()
	if err != nil {
		logger.Critical("Server can't run: %w", err)
	}
}
