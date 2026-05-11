package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"findus/internal/config"
	"findus/internal/platform/logger"
	"findus/internal/repository/sqlite"
	"findus/internal/secrets"
	"findus/internal/service"
	"findus/internal/transport/http/handler"
	"findus/internal/transport/http/render"
	"findus/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.LogLevel)

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

	tplFS, err := fs.Sub(web.Assets, "templates")
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

	go func() {
		log.Info("listening", "addr", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shCtx)
}
