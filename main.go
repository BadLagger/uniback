package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"uniback/controller"
	"uniback/repository/postgres"
	"uniback/service"
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

	Service := service.NewTransactionService(DataBase)

	CryptoService := service.NewPgpHmacService(service.PgpHmacConfgiFromGlobalConfig(cfg))

	if CryptoService == nil {
		logger.Critical("CryptoService init fail!!")
		return
	}

	authController := controller.NewAuthController(DataBase, CryptoService, Service, cfg.JwtKey)
	http.HandleFunc("/register", authController.RegistrationHandler)
	http.HandleFunc("/login", authController.LoginHandler)
	//
	http.HandleFunc("/accounts", authController.AuthMiddleware(authController.AccountsHandler))
	http.HandleFunc("/accounts/new", authController.AuthMiddleware(authController.AccountsCreateHandler))
	http.HandleFunc("/accounts/deposit", authController.AuthMiddleware(authController.DepositHandler))
	http.HandleFunc("/accounts/withdrawal", authController.AuthMiddleware(authController.WithdrawalHandler))
	http.HandleFunc("/accounts/transfer", authController.AuthMiddleware(authController.TransferHandler))
	//
	http.HandleFunc("/cards", authController.AuthMiddleware(authController.ShowCardsHandler))
	http.HandleFunc("/cards/new", authController.AuthMiddleware(authController.NewCardHandler))
	//
	http.HandleFunc("/credits", authController.AuthMiddleware(authController.ShowCreditsHanlder))
	http.HandleFunc("/credits/new", authController.AuthMiddleware(authController.NewCreditHandler))
	//
	http.HandleFunc("/analytics", authController.AuthMiddleware(authController.AnalyticsHanlder))

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
