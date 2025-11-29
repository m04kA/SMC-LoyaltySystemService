package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/api/handlers/configure_loyalty"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/api/handlers/create_loyalty_card"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/api/handlers/get_loyalty_card"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/api/middleware"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/config"
	loyaltyCardRepo "github.com/m04kA/SMC-LoyaltySystemService/internal/infra/storage/loyalty_card"
	loyaltyConfigRepo "github.com/m04kA/SMC-LoyaltySystemService/internal/infra/storage/loyalty_config"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/integrations/sellerservice"
	loyaltyService "github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty"
	"github.com/m04kA/SMC-LoyaltySystemService/pkg/dbmetrics"
	"github.com/m04kA/SMC-LoyaltySystemService/pkg/logger"
	"github.com/m04kA/SMC-LoyaltySystemService/pkg/metrics"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load("config.toml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем логгер
	log, err := logger.New(cfg.Logs.File)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Starting SMC-LoyaltySystemService...")
	log.Info("Configuration loaded from config.toml")

	// Инициализируем метрики (если включены)
	var metricsCollector *metrics.Metrics
	var wrappedDB *dbmetrics.DB
	stopMetricsCh := make(chan struct{})

	if cfg.Metrics.Enabled {
		metricsCollector = metrics.New(cfg.Metrics.ServiceName)
		log.Info("Metrics enabled at %s", cfg.Metrics.Path)
	}

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Настраиваем connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database: %v", err)
	}
	log.Info("Successfully connected to database (host=%s, port=%d, db=%s)",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// Инициализируем клиент SellerService
	sellerClient := sellerservice.NewClient(
		cfg.SellerService.BaseURL,
		time.Duration(cfg.SellerService.Timeout)*time.Second,
		log,
	)
	log.Info("SellerService client initialized (base_url=%s)", cfg.SellerService.BaseURL)

	// Инициализируем репозитории и сервисы (с метриками или без)
	var loyaltySvc *loyaltyService.Service

	if cfg.Metrics.Enabled {
		wrappedDB = dbmetrics.WrapWithDefault(db, metricsCollector, cfg.Metrics.ServiceName, stopMetricsCh)
		log.Info("Database metrics collection started")

		// Инициализируем репозитории с обёрткой метрик
		cardRepository := loyaltyCardRepo.NewRepository(wrappedDB)
		configRepository := loyaltyConfigRepo.NewRepository(wrappedDB)

		loyaltySvc = loyaltyService.NewService(cardRepository, configRepository, sellerClient)
	} else {
		// Инициализируем репозитории без метрик
		cardRepository := loyaltyCardRepo.NewRepository(db)
		configRepository := loyaltyConfigRepo.NewRepository(db)

		loyaltySvc = loyaltyService.NewService(cardRepository, configRepository, sellerClient)
	}

	// Инициализируем handlers
	getLoyaltyCardHandler := get_loyalty_card.NewHandler(loyaltySvc, log)
	createLoyaltyCardHandler := create_loyalty_card.NewHandler(loyaltySvc, log)
	configureLoyaltyHandler := configure_loyalty.NewHandler(loyaltySvc, log)

	// Настраиваем роутер
	r := mux.NewRouter()

	// Добавляем metrics middleware (если метрики включены)
	if cfg.Metrics.Enabled {
		r.Use(middleware.MetricsMiddleware(metricsCollector, cfg.Metrics.ServiceName))
		log.Info("HTTP metrics middleware enabled")
	}

	// Metrics endpoint (публичный, без аутентификации)
	if cfg.Metrics.Enabled {
		r.Handle(cfg.Metrics.Path, promhttp.Handler()).Methods(http.MethodGet)
		log.Info("Prometheus metrics endpoint exposed at %s", cfg.Metrics.Path)
	}

	// API prefix
	api := r.PathPrefix("/api/v1").Subrouter()

	// Public routes (не требуют аутентификации)
	api.HandleFunc("/loyalty-cards", getLoyaltyCardHandler.Handle).Methods(http.MethodGet)
	api.HandleFunc("/loyalty-cards", createLoyaltyCardHandler.Handle).Methods(http.MethodPost)

	// Protected routes (требуют X-User-ID)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth)

	// Protected routes для конфигурации лояльности
	protected.HandleFunc("/companies/{companyId}/loyalty-config", configureLoyaltyHandler.Handle).Methods(http.MethodPost)

	// Создаем HTTP сервер
	addr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Останавливаем сбор метрик connection pool
	if cfg.Metrics.Enabled {
		close(stopMetricsCh)
		log.Info("Metrics collection stopped")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.Server.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped gracefully")
}
