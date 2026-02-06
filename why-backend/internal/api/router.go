package api

import (
	"database/sql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"why-backend/internal/api/handlers"
	"why-backend/internal/api/middleware"
	"why-backend/internal/config"
)

func NewRouter(db *sql.DB, minio *minio.Client, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("why-backend")) // OpenTelemetry tracing

	// CORS middleware to allow browser requests
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://why.local:8000", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Prometheus metrics endpoint for Alloy to scrape
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	messageHandler := handlers.NewMessageHandler(db)
	mediaHandler := handlers.NewMediaHandler(minio, cfg)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/signup", authHandler.Signup)
		v1.POST("/login", authHandler.Login)

		// Public read-only routes
		v1.GET("/messages", messageHandler.ListMessages)
		v1.GET("/messages/:id", messageHandler.GetMessage)
		v1.GET("/messages/:id/replies", messageHandler.ListReplies)

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.POST("/messages", messageHandler.CreateMessage)
			protected.POST("/messages/:id/replies", messageHandler.CreateReply)
			protected.POST("/media", mediaHandler.UploadMedia)
		}
	}

	return r
}
