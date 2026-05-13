package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erc-pham/surus/api/internal/auth"
	"github.com/erc-pham/surus/api/internal/course"
	"github.com/erc-pham/surus/api/internal/fork"
	"github.com/erc-pham/surus/api/internal/jobs"
	"github.com/erc-pham/surus/api/internal/lesson"
	"github.com/erc-pham/surus/api/internal/middleware"
	"github.com/erc-pham/surus/api/internal/quiz"
	"github.com/erc-pham/surus/api/internal/report"
	"github.com/erc-pham/surus/api/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "err", err)
		os.Exit(1)
	}

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	authCfg := &middleware.AuthConfig{JWTSecret: jwtSecret}

	frontendURL := os.Getenv("NEXT_PUBLIC_APP_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	authService := auth.NewService(auth.Config{
		Pool:              pool,
		JWTSecret:         jwtSecret,
		GoogleClientID:    os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret:      os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL: os.Getenv("GOOGLE_REDIRECT_URL"),
		MagicLinkBaseURL:  os.Getenv("MAGIC_LINK_BASE_URL"),
		AppEnv:            os.Getenv("APP_ENV"),
	})

	authHandler := auth.NewHandler(authService, frontendURL)
	userHandler := user.NewHandler(user.NewService(pool))
	courseHandler := course.NewHandler(course.NewService(pool))
	lessonHandler := lesson.NewHandler(lesson.NewService(pool))
	quizHandler := quiz.NewHandler(pool)
	forkHandler := fork.NewHandler(pool)
	reportHandler := report.NewHandler(pool)

	r := chi.NewRouter()

	r.Use(middleware.CORS(frontendURL))
	r.Use(middleware.RateLimit(middleware.NewRateLimiter(300)))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// OAuth browser-redirect routes at root (Google redirect URL doesn't include /v1)
	r.Get("/auth/google/start", authHandler.GoogleStartRedirect)
	r.Get("/auth/google/callback", authHandler.GoogleCallbackRedirect)

	r.Route("/v1", func(r chi.Router) {
		authRL := middleware.NewRateLimiter(10)
		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.RateLimit(authRL))
			r.Mount("/", authHandler.Routes(authCfg))
		})

		r.Mount("/users", userHandler.Routes(authCfg))

		r.Mount("/courses", courseHandler.Routes(authCfg))

		r.Route("/courses/{courseId}/lessons", func(r chi.Router) {
			r.Mount("/", lessonHandler.Routes(authCfg))

			r.Route("/{lessonId}/attempts", func(r chi.Router) {
				r.Mount("/", quizHandler.Routes(authCfg))
			})

			r.Group(func(r chi.Router) {
				r.Use(authCfg.RequireAuth)
				r.Post("/{lessonId}/report", reportHandler.ReportLesson)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(authCfg.RequireAuth)
			r.Post("/courses/{courseId}/fork", forkHandler.ForkCourse)
			r.Post("/courses/{courseId}/report", reportHandler.ReportCourse)
		})

		r.Mount("/admin/reports", reportHandler.AdminRoutes(authCfg))
	})

	linkRotChecker := jobs.NewLinkRotChecker(pool)
	go linkRotChecker.Run(ctx)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
