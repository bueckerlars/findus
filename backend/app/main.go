package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"findus/backend/internal/config"
	"findus/backend/internal/platform/logger"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/secrets"
	"findus/backend/internal/service"
	"findus/backend/internal/transport/http/handler"
	"findus/backend/internal/transport/http/render"
	"findus/frontend"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(logger.Options{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
	slog.SetDefault(log)

	ctx := context.Background()
	jwtSecret, err := secrets.LoadOrCreateJWTSecret(cfg.DataDir, cfg.JWTSecret)
	if err != nil {
		panic(err)
	}

	db, err := sqlite.OpenDB(ctx, cfg.DataDir)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	log.Info("application starting",
		slog.String("event", "startup"),
		slog.Int("http.port", cfg.Port),
		slog.String("data_dir", cfg.DataDir),
		slog.String("base_url", cfg.BaseURL),
		slog.Bool("cookie_secure", cfg.CookieSecure),
		slog.String("log_level", cfg.LogLevel),
		slog.String("log_format", cfg.LogFormat),
	)

	users := sqlite.NewUserRepo(db)
	locs := sqlite.NewLocationRepo(db)
	items := sqlite.NewItemRepo(db)
	templates := sqlite.NewItemTemplateRepo(db)
	invites := sqlite.NewInviteRepo(db)
	settings := sqlite.NewSettingsRepo(db)

	authSvc := &service.Auth{Users: users, Settings: settings, Invites: invites}
	adminSvc := &service.Admin{Users: users, Settings: settings, Invites: invites}
	labels := sqlite.NewLabelRepo(db)
	invSvc := &service.Inventory{Locations: locs, Items: items, Labels: labels, Templates: templates}
	qrSvc := &service.QR{BaseURL: cfg.BaseURL}

	tplFS, err := fs.Sub(frontend.Assets, "templates")
	if err != nil {
		panic(err)
	}
	tpl, err := render.Parse(tplFS)
	if err != nil {
		panic(err)
	}

	srv := &handler.Server{
		Log:       log,
		Config:    cfg,
		DB:        db,
		Users:     users,
		Locs:      locs,
		Items:     items,
		Labels:    labels,
		Templates: templates,
		Invites:   invites,
		Settings:  settings,
		Auth:      authSvc,
		Admin:     adminSvc,
		Inventory: invSvc,
		QR:        qrSvc,
		Backup:    &service.Backup{DataDir: cfg.DataDir},
		JWTSecret: jwtSecret,
		Tpl:       tpl,
	}

	h, err := srv.Handler()
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	log.Info("http server listening",
		slog.String("event", "server_listen"),
		slog.String("addr", addr),
	)

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Info("shutdown signal received",
			slog.String("event", "shutdown_signal"),
			slog.String("signal", sig.String()),
		)
	case err := <-errCh:
		signal.Stop(stop)
		if err != nil {
			log.Error("http server failed", slog.String("event", "server_error"), slog.Any("err", err))
			_ = db.Close()
			os.Exit(1)
		}
		// Server exited cleanly before any shutdown signal (unexpected in normal operation).
		return
	}

	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- httpSrv.Shutdown(shCtx)
	}()

	select {
	case err := <-shutdownDone:
		signal.Stop(stop)
		if err != nil {
			log.Error("graceful shutdown error", slog.String("event", "shutdown_error"), slog.Any("err", err))
		} else {
			log.Info("http server stopped", slog.String("event", "server_stopped"))
		}
	case sig := <-stop:
		log.Warn("shutdown signal during graceful stop; forcing close",
			slog.String("event", "shutdown_force"),
			slog.String("signal", sig.String()),
		)
		_ = httpSrv.Close()
		err := <-shutdownDone
		signal.Stop(stop)
		if err != nil {
			log.Warn("http server shutdown after force close",
				slog.String("event", "shutdown_force_done"),
				slog.Any("err", err),
			)
		}
		log.Info("http server stopped", slog.String("event", "server_stopped"))
	}
}
