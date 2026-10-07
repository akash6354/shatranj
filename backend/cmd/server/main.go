package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akash6354/shatranj/backend/internal/achievements"
	"github.com/akash6354/shatranj/backend/internal/admin"
	"github.com/akash6354/shatranj/backend/internal/analysis"
	"github.com/akash6354/shatranj/backend/internal/auth"
	"github.com/akash6354/shatranj/backend/internal/cache"
	"github.com/akash6354/shatranj/backend/internal/chat"
	"github.com/akash6354/shatranj/backend/internal/clubs"
	"github.com/akash6354/shatranj/backend/internal/coaches"
	"github.com/akash6354/shatranj/backend/internal/config"
	"github.com/akash6354/shatranj/backend/internal/database"
	"github.com/akash6354/shatranj/backend/internal/friendships"
	"github.com/akash6354/shatranj/backend/internal/games"
	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/leaderboard"
	"github.com/akash6354/shatranj/backend/internal/lessons"
	"github.com/akash6354/shatranj/backend/internal/matchmaking"
	"github.com/akash6354/shatranj/backend/internal/middleware"
	"github.com/akash6354/shatranj/backend/internal/news"
	"github.com/akash6354/shatranj/backend/internal/notifications"
	"github.com/akash6354/shatranj/backend/internal/observability"
	"github.com/akash6354/shatranj/backend/internal/payments"
	"github.com/akash6354/shatranj/backend/internal/profiles"
	"github.com/akash6354/shatranj/backend/internal/puzzles"
	"github.com/akash6354/shatranj/backend/internal/ratings"
	"github.com/akash6354/shatranj/backend/internal/realtime"
	"github.com/akash6354/shatranj/backend/internal/review"
	"github.com/akash6354/shatranj/backend/internal/subscriptions"
	"github.com/akash6354/shatranj/backend/internal/tournaments"
	"github.com/akash6354/shatranj/backend/internal/users"
)

const (
	readHeaderTimeout = 5 * time.Second
	startupTimeout    = 5 * time.Second
	gameClockPoll     = 250 * time.Millisecond
)

func main() {
	logger := observability.NewLogger()
	if err := run(logger); err != nil {
		logger.Error("server stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), startupTimeout)
	defer cancelStartup()

	db, err := database.OpenPostgres(startupCtx, cfg.PostgresDSN)
	if err != nil {
		return err
	}
	if db != nil {
		defer func() {
			if err := db.Close(); err != nil {
				logger.Error("close PostgreSQL connection pool", "error", err)
			}
		}()
		logger.Info("PostgreSQL connected")
	}

	cacheClient, err := cache.Open(startupCtx, cfg.RedisURL)
	if err != nil {
		logger.Warn("Redis unavailable, using in-memory cache fallback", "error", err)
	}
	defer func() {
		if err := cacheClient.Close(); err != nil {
			logger.Error("close cache client", "error", err)
		}
	}()
	logger.Info(cacheClient.Describe())

	var tokens *auth.TokenManager
	if db != nil {
		tokens, err = auth.NewTokenManager(cfg.JWTSecret, 15*time.Minute)
		if err != nil {
			return err
		}
	}

	hub := realtime.NewHubWithPresence(logger, cacheClient)
	defer hub.CloseAll()
	var clockService *games.Service
	if db != nil {
		clockService = games.NewServiceWithRatings(
			games.NewRepository(db), hub, ratings.NewService(ratings.NewRepository(db)),
		)
		clockService.SetAchievementTracker(achievements.NewService(achievements.NewRepository(db)))
	}
	var handler http.Handler = newHandlerWithConfig(db, tokens, hub, cfg.CORSOrigins, cfg, cacheClient)

	handler = middleware.CORS(cfg.CORSOrigins)(handler)
	handler = middleware.Recovery(logger)(handler)
	handler = middleware.Logging(logger)(handler)
	handler = middleware.RequestID(handler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	if cacheClient.RedisBacked() {
		if err := hub.EnableGameEventFanout(ctx); err != nil {
			logger.Warn("shared game event fanout unavailable", "error", err)
		}
	}
	var clockDone chan struct{}
	if clockService != nil {
		clockDone = make(chan struct{})
		go func() {
			defer close(clockDone)
			runGameClockExpiry(ctx, clockService, logger)
		}()
	}
	defer func() {
		stop()
		if clockDone != nil {
			<-clockDone
		}
	}()

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", "address", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-serverErrors
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func runGameClockExpiry(ctx context.Context, service *games.Service, logger *slog.Logger) {
	ticker := time.NewTicker(gameClockPoll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := service.ExpireDueGames(ctx, 100); err != nil && ctx.Err() == nil {
				logger.ErrorContext(ctx, "expire game clocks", "error", err)
			}
			if err := service.ProcessCompletionJobs(ctx, 100); err != nil && ctx.Err() == nil {
				logger.ErrorContext(ctx, "process game completion jobs", "error", err)
			}
		}
	}
}

func newHandler(db *sql.DB, tokens *auth.TokenManager) http.Handler {
	return newHandlerWithOrigins(db, tokens, realtime.NewHub(nil), nil)
}

func newHandlerWithOrigins(db *sql.DB, tokens *auth.TokenManager, hub *realtime.Hub, corsOrigins []string) http.Handler {
	return newHandlerWithConfig(db, tokens, hub, corsOrigins, config.Config{PremiumPricePaise: 49_900}, nil)
}

func newHandlerWithConfig(db *sql.DB, tokens *auth.TokenManager, hub *realtime.Hub, corsOrigins []string, cfg config.Config, store cache.Store) http.Handler {
	return httpapi.NewRouter(func(mux *http.ServeMux) {
		if db == nil || tokens == nil {
			unavailable := func(w http.ResponseWriter, _ *http.Request) {
				httpapi.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "feature APIs require configured PostgreSQL and JWT signing")
			}
			mux.HandleFunc("/api/v1/auth/", unavailable)
			mux.HandleFunc("/api/v1/users/", unavailable)
			mux.HandleFunc("/api/v1/profiles/", unavailable)
			mux.HandleFunc("/api/v1/games", unavailable)
			mux.HandleFunc("/api/v1/games/", unavailable)
			mux.HandleFunc("/api/v1/reviews/", unavailable)
			mux.HandleFunc("/api/v1/matchmaking/", unavailable)
			mux.HandleFunc("/api/v1/ratings/", unavailable)
			mux.HandleFunc("/api/v1/leaderboard/", unavailable)
			mux.HandleFunc("/api/v1/puzzles", unavailable)
			mux.HandleFunc("/api/v1/puzzles/", unavailable)
			mux.HandleFunc("/api/v1/lessons", unavailable)
			mux.HandleFunc("/api/v1/lessons/", unavailable)
			mux.HandleFunc("/api/v1/tournaments", unavailable)
			mux.HandleFunc("/api/v1/tournaments/", unavailable)
			mux.HandleFunc("/api/v1/clubs", unavailable)
			mux.HandleFunc("/api/v1/clubs/", unavailable)
			mux.HandleFunc("/api/v1/friends", unavailable)
			mux.HandleFunc("/api/v1/friends/", unavailable)
			mux.HandleFunc("/api/v1/chat", unavailable)
			mux.HandleFunc("/api/v1/chat/", unavailable)
			mux.HandleFunc("/api/v1/achievements", unavailable)
			mux.HandleFunc("/api/v1/achievements/", unavailable)
			mux.HandleFunc("/api/v1/notifications", unavailable)
			mux.HandleFunc("/api/v1/notifications/", unavailable)
			mux.HandleFunc("/api/v1/payments", unavailable)
			mux.HandleFunc("/api/v1/payments/", unavailable)
			mux.HandleFunc("/api/v1/subscriptions", unavailable)
			mux.HandleFunc("/api/v1/subscriptions/", unavailable)
			mux.HandleFunc("/api/v1/coaches", unavailable)
			mux.HandleFunc("/api/v1/coaches/", unavailable)
			mux.HandleFunc("/api/v1/news", unavailable)
			mux.HandleFunc("/api/v1/news/", unavailable)
			mux.HandleFunc("/api/v1/admin/", unavailable)
			return
		}
		auth.RegisterRoutesWithCache(mux, auth.NewHandler(auth.NewService(auth.NewRepository(db), tokens)), tokens, store)
		users.RegisterRoutes(mux, users.NewRepository(db), tokens)
		profiles.RegisterRoutes(mux, profiles.NewRepository(db), tokens)
		achievementService := achievements.NewService(achievements.NewRepository(db))
		gameRepository := games.NewRepository(db)
		ratingService := ratings.NewService(ratings.NewRepository(db))
		gameService := games.NewServiceWithRatings(gameRepository, hub, ratingService)
		gameService.SetAchievementTracker(achievementService)
		games.RegisterRoutes(mux, games.NewHandler(gameService), tokens)
		reviewRepository := review.NewRepository(db)
		reviewService := review.NewService(
			reviewRepository, gameRepository,
			analysis.NewStockfish(os.Getenv("STOCKFISH_PATH"), 15*time.Second, 14).Available(),
		)
		review.RegisterRoutes(mux, review.NewHandler(reviewService), tokens)
		ratings.RegisterRoutes(mux, ratings.NewHandler(ratingService), tokens)
		leaderboard.RegisterRoutes(mux, leaderboard.NewHandler(leaderboard.NewService(leaderboard.NewRepository(db))))
		matchmakingService := matchmaking.NewService(matchmaking.NewQueueRepository(db), ratingService, gameService)
		matchmaking.RegisterRoutes(mux, matchmaking.NewHandler(matchmakingService), tokens)
		puzzleService := puzzles.NewService(puzzles.NewRepository(db))
		puzzleService.SetAchievementTracker(achievementService)
		puzzles.RegisterRoutes(mux, puzzles.NewHandler(puzzleService), tokens)
		lessonService := lessons.NewService(lessons.NewRepository(db))
		lessons.RegisterRoutes(mux, lessons.NewHandler(lessonService), tokens)
		tournaments.RegisterRoutes(mux, tournaments.NewHandler(tournaments.NewService(tournaments.NewRepository(db))), tokens)
		clubs.RegisterRoutes(mux, clubs.NewHandler(clubs.NewService(clubs.NewRepository(db))), tokens)
		friendshipService := friendships.NewService(friendships.NewRepository(db))
		friendships.RegisterRoutes(mux, friendships.NewHandler(friendshipService), tokens)
		chatService := chat.NewService(chat.NewRepository(db), friendshipService, hub, nil)
		chat.RegisterRoutes(mux, chat.NewHandler(chatService), tokens)
		achievements.RegisterRoutes(mux, achievements.NewHandler(achievementService), tokens)
		notificationService := notifications.NewService(notifications.NewRepository(db), nil)
		notifications.RegisterRoutes(mux, notifications.NewHandler(notificationService), tokens)

		subscriptionService := subscriptions.NewService(subscriptions.NewRepository(db), cfg.PremiumPricePaise)
		subscriptions.RegisterRoutes(mux, subscriptions.NewHandler(subscriptionService), tokens)
		var razorpay payments.Provider
		if cfg.RazorpayKeyID != "" && cfg.RazorpayKeySecret != "" && cfg.RazorpayWebhookSecret != "" {
			razorpay = payments.NewRazorpayClient(cfg.RazorpayKeyID, cfg.RazorpayKeySecret, cfg.RazorpayWebhookSecret)
		}
		paymentService := payments.NewService(payments.NewRepository(db), razorpay, subscriptionService, cfg.PremiumPricePaise)
		payments.RegisterRoutes(mux, payments.NewHandler(paymentService), tokens)
		coachService := coaches.NewService(coaches.NewRepository(db))
		coaches.RegisterRoutes(mux, coaches.NewHandler(coachService), tokens)

		adminRepository := admin.NewRepository(db)
		adminOnly := admin.AdminOnly(tokens, adminRepository)
		admin.RegisterRoutes(mux, admin.NewHandler(admin.NewService(adminRepository)), adminOnly)
		news.RegisterRoutes(mux, news.NewHandler(news.NewService(news.NewRepository(db))), adminOnly)
		if hub != nil {
			realtime.RegisterGameRoute(mux, hub, tokens, corsOrigins, func(ctx context.Context, gameID, userID string) error {
				_, err := gameService.Get(ctx, gameID, userID)
				return err
			})
			realtime.RegisterChatRoute(mux, hub, tokens, corsOrigins, chatService)
		}
	})
}
