package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ranking-service/docs" // swagger generated docs

	"ranking-service/config"
	"ranking-service/internal/handlers"
	"ranking-service/internal/repository"
)

var (
	runService = &cobra.Command{
		Use:   "server",
		Short: "Run ranking service",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}
)

//	@title			Ranking Service API
//	@version		1.0
//	@description	Swagger docs for Ranking Service API
//
// // @securityDefinitions.apikey	Bearer
// // @in							header
// // @name						Authorization
//
//	@BasePath		/
func runServer() {
	cfg := config.MustLoadServerConfigFromEnv()

	postgresDb, err := repository.NewPostgresDB(cfg.Postgres)
	if err != nil {
		slog.Error("Failed to connect PostgreSQL:", "error", err)
		os.Exit(1)
	}

	redisDb, err := repository.NewRedisDB(cfg.Redis)
	if err != nil {
		slog.Error("Failed to connect Redis:", "error", err)
		os.Exit(1)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Swagger docs
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rankingHandler := handlers.NewRankingHandler(postgresDb, redisDb)

	// API Endpoints
	router.POST("/videos/:video_id/interaction", rankingHandler.UpdateVideoScoreHandler())
	router.GET("/videos/top", rankingHandler.GetGlobalTopVideosHandler())
	router.GET("/users/:userID/videos/top", rankingHandler.GetUserTopVideosHandler())

	// Notify server start/stop
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		slog.Info("Starting ranking service server on ", "address", fmt.Sprintf("%s:%s", cfg.ListenAddr, cfg.Port))
		if err := router.Run(":" + cfg.Port); err != nil {
			slog.Error("Failed to run server", "error", err)
		}
	}()
	<-ctx.Done()
	slog.Info("Shutdown ranking service server")
}
