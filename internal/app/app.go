package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saimonsiddique/blog-api/internal/config"
	"github.com/saimonsiddique/blog-api/internal/database"
	"github.com/saimonsiddique/blog-api/internal/handler"
	"github.com/saimonsiddique/blog-api/internal/queue"
	"github.com/saimonsiddique/blog-api/internal/repository"
	"github.com/saimonsiddique/blog-api/internal/service"
	"github.com/saimonsiddique/blog-api/internal/worker"
	"github.com/sirupsen/logrus"
)

const (
	readTimeout  = 15 * time.Second
	writeTimeout = 15 * time.Second
	idleTimeout  = 60 * time.Second
)

type App struct {
	config       *config.Config
	router       *gin.Engine
	logger       *logrus.Logger
	server       *http.Server
	db           *pgxpool.Pool
	queue        *queue.RabbitMQ
	worker       *worker.PostPublishWorker
	workerCtx    context.Context
	workerCancel context.CancelFunc
}

func New(cfg *config.Config) (*App, error) {
	// Initialize logger
	logger := initLogger(cfg.App.Environment)

	// Initialize database
	db, err := database.NewPostgresPool(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize RabbitMQ
	queueCfg := &queue.Config{
		Host:     cfg.RabbitMQ.Host,
		Port:     cfg.RabbitMQ.Port,
		User:     cfg.RabbitMQ.User,
		Password: cfg.RabbitMQ.Password,
		Vhost:    cfg.RabbitMQ.Vhost,
	}
	rabbitMQ, err := queue.NewRabbitMQ(queueCfg, logger)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize RabbitMQ: %w", err)
	}

	// Initialize worker
	postPublishWorker := worker.NewPostPublishWorker(rabbitMQ, db, logger)

	// Configure Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create worker context
	workerCtx, workerCancel := context.WithCancel(context.Background())

	app := &App{
		config:       cfg,
		router:       gin.New(),
		logger:       logger,
		db:           db,
		queue:        rabbitMQ,
		worker:       postPublishWorker,
		workerCtx:    workerCtx,
		workerCancel: workerCancel,
	}

	// Setup middleware
	app.setupMiddleware()

	// Setup routes
	app.setupRoutes()

	// Start worker
	if err := app.worker.Start(app.workerCtx); err != nil {
		app.cleanup()
		return nil, fmt.Errorf("failed to start worker: %w", err)
	}

	return app, nil
}

func initLogger(env string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	if env == "production" {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	return logger
}

func (a *App) setupMiddleware() {
	// Recovery middleware
	a.router.Use(gin.Recovery())

	// Logger middleware
	a.router.Use(gin.Logger())
}

func (a *App) setupRoutes() {
	// Initialize repositories
	userRepo := repository.NewUserRepository(a.db)
	authRepo := repository.NewAuthRepository(a.db)
	postRepo := repository.NewPostRepository(a.db)

	// Initialize queue publisher
	postPublisher := queue.NewPostPublisher(a.queue)

	// Initialize services
	authService := service.NewAuthService(userRepo, authRepo, &a.config.JWT)
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo, userRepo, postPublisher)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(a.db)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	postHandler := handler.NewPostHandler(postService)

	// Health check
	a.router.GET("/health", healthHandler.HealthCheck)

	// API v1 routes
	v1 := a.router.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Public post routes
		v1.GET("/posts", postHandler.ListPosts)
		v1.GET("/posts/:id", postHandler.GetPost)

		// Protected routes
		protected := v1.Group("")
		protected.Use(handler.AuthMiddleware(&a.config.JWT))
		{
			// User routes
			protected.GET("/me", userHandler.GetProfile)
			protected.PUT("/me", userHandler.UpdateProfile)

			// Post routes
			protected.POST("/posts", postHandler.CreatePost)
			protected.PUT("/posts/:id", postHandler.UpdatePost)
			protected.DELETE("/posts/:id", postHandler.DeletePost)
		}
	}
}

func (a *App) Run() error {
	addr := fmt.Sprintf("%s:%s", a.config.Server.Host, a.config.Server.Port)

	a.logger.WithFields(logrus.Fields{
		"address":     addr,
		"environment": a.config.App.Environment,
	}).Info("Starting server")

	// Create HTTP server
	a.server = &http.Server{
		Addr:         addr,
		Handler:      a.router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down server...")

	if a.server == nil {
		return nil
	}

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.WithError(err).Error("Server shutdown failed")
		return err
	}

	a.logger.Info("Server shutdown successful")
	return nil
}

func (a *App) Close() {
	a.cleanup()
}

func (a *App) cleanup() {
	a.logger.Info("Cleaning up resources...")

	// Stop worker
	if a.workerCancel != nil {
		a.workerCancel()
		a.logger.Info("Worker stopped")
	}

	// Close RabbitMQ
	if a.queue != nil {
		a.queue.Close()
		a.logger.Info("RabbitMQ connection closed")
	}

	// Close database
	if a.db != nil {
		a.db.Close()
		a.logger.Info("Database connection closed")
	}
}
