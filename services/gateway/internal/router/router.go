package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/config"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/handler"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/middleware"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/relay"
)

func Setup(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", cfg.AllowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "ai-token-gateway",
		})
	})

	// 初始化 stores
	var userStore *model.UserStore
	var apiKeyStore *model.APIKeyStore
	var logStore *model.LogStore
	var channelStore *model.ChannelStore
	var modelStore *model.ModelStore
	if db != nil {
		userStore = model.NewUserStore(db)
		apiKeyStore = model.NewAPIKeyStore(db)
		logStore = model.NewLogStore(db)
		channelStore = model.NewChannelStore(db)
		modelStore = model.NewModelStore(db)
	}

	// 初始化中转引擎
	var logWriter relay.LogWriter
	if db != nil {
		logWriter = relay.NewDBLogWriter(db)
	}
	engine := relay.NewRelayEngine(logWriter)
	relayHandler := handler.NewRelayHandler(engine)
	relayHandler.SetChannels(loadChannelsFromEnv(cfg))

	// OpenAI Compatible API（需要 API Key 认证）
	v1 := r.Group("/v1")
	if apiKeyStore != nil {
		v1.Use(middleware.APIKeyAuth(apiKeyStore))
		if rdb != nil {
			v1.Use(middleware.RateLimit(rdb, 60))
		}
	}
	{
		v1.GET("/models", relayHandler.ListModels)
		v1.POST("/chat/completions", relayHandler.ChatCompletions)
		v1.POST("/completions", relayHandler.Completions)
		v1.POST("/embeddings", placeholder("embeddings"))
		v1.POST("/images/generations", placeholder("images"))
	}

	// 内部管理 API
	api := r.Group("/api")
	{
		// 认证接口（无需登录）
		auth := api.Group("/auth")
		if userStore != nil {
			authHandler := handler.NewAuthHandler(userStore, cfg.JWTSecret)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		} else {
			auth.POST("/register", placeholder("register"))
			auth.POST("/login", placeholder("login"))
			auth.POST("/refresh", placeholder("refresh"))
		}

		// 用户接口（需要 JWT）
		user := api.Group("/user")
		user.Use(middleware.JWTAuth(cfg.JWTSecret))
		if userStore != nil && apiKeyStore != nil && logStore != nil {
			userHandler := handler.NewUserHandler(userStore, apiKeyStore, logStore)
			user.GET("/profile", userHandler.Profile)
			user.GET("/dashboard", userHandler.Dashboard)
			user.GET("/keys", userHandler.ListKeys)
			user.POST("/keys", userHandler.CreateKey)
			user.DELETE("/keys/:id", userHandler.DeleteKey)
			user.GET("/logs", userHandler.Logs)
			user.GET("/usage", userHandler.Usage)
		}

		// 管理员接口
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth(cfg.JWTSecret))
		admin.Use(middleware.AdminOnly())
		if channelStore != nil && modelStore != nil && userStore != nil && logStore != nil {
			adminHandler := handler.NewAdminHandler(channelStore, modelStore, userStore, logStore)
			admin.GET("/channels", adminHandler.ListChannels)
			admin.POST("/channels", adminHandler.CreateChannel)
			admin.PUT("/channels/:id", adminHandler.UpdateChannel)
			admin.DELETE("/channels/:id", adminHandler.DeleteChannel)
			admin.GET("/models", adminHandler.ListModels)
			admin.POST("/models", adminHandler.CreateModel)
			admin.PUT("/models/:id", adminHandler.UpdateModel)
			admin.DELETE("/models/:id", adminHandler.DeleteModel)
			admin.GET("/users", adminHandler.ListUsers)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.GET("/logs", adminHandler.GlobalLogs)
			admin.GET("/stats", adminHandler.Stats)
		}
	}

	return r
}

func loadChannelsFromEnv(cfg *config.Config) []relay.Channel {
	var channels []relay.Channel

	if cfg.OpenAIKey != "" {
		channels = append(channels, relay.Channel{
			ID: 1, Name: "OpenAI Default", Provider: "openai",
			BaseURL: cfg.OpenAIBaseURL, APIKey: cfg.OpenAIKey,
			Models: []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo"},
			Status: 1, Priority: 10, Weight: 1,
		})
	}

	if cfg.AnthropicKey != "" {
		channels = append(channels, relay.Channel{
			ID: 2, Name: "Anthropic Default", Provider: "anthropic",
			BaseURL: cfg.AnthropicBaseURL, APIKey: cfg.AnthropicKey,
			Models: []string{"claude-opus-4-7", "claude-sonnet-4-6", "claude-haiku-4-5"},
			Status: 1, Priority: 10, Weight: 1,
		})
	}

	if cfg.GoogleKey != "" {
		channels = append(channels, relay.Channel{
			ID: 3, Name: "Google Default", Provider: "google",
			BaseURL: cfg.GoogleBaseURL, APIKey: cfg.GoogleKey,
			Models: []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"},
			Status: 1, Priority: 10, Weight: 1,
		})
	}

	if cfg.DeepSeekKey != "" {
		channels = append(channels, relay.Channel{
			ID: 4, Name: "DeepSeek Default", Provider: "deepseek",
			BaseURL: cfg.DeepSeekBaseURL, APIKey: cfg.DeepSeekKey,
			Models: []string{"deepseek-chat", "deepseek-reasoner"},
			Status: 1, Priority: 10, Weight: 1,
		})
	}

	return channels
}

func placeholder(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":    "not_implemented",
			"endpoint": name,
		})
	}
}
