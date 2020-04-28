package main

import (
	"ddrp-relayer/config"
	"ddrp-relayer/export"
	"ddrp-relayer/healthcheck"
	"ddrp-relayer/log"
	"ddrp-relayer/protocol"
	"ddrp-relayer/social"
	"ddrp-relayer/store"
	"ddrp-relayer/tlds"
	"ddrp-relayer/user"
	"ddrp-relayer/version"
	"ddrp-relayer/web"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/configor"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) == 1 {
		exitErr(errors.New("must specify a config file path"))
	}

	cfgPath := os.Args[1]
	var cfg config.Config
	if err := configor.Load(&cfg, cfgPath); err != nil {
		exitErr(errors.Wrap(err, "invalid configuration"))
	}

	lgr := log.WithModule("main")
	db, err := store.Connect(&cfg.Database)
	if err != nil {
		exitErr(err)
	}

	ddrpClient, err := protocol.DialRPC(cfg.DDRP.Address)
	if err != nil {
		exitErr(err)
	}
	var kns []protocol.KeysNames
	for _, tldCfg := range cfg.TLDs {
		kns = append(kns, protocol.KeysNames{
			Name: tldCfg.Name,
			Key:  tldCfg.PrivateKey,
		})
	}
	signer, err := protocol.NewNameSigner(kns)
	if err != nil {
		exitErr(err)
	}

	if err := tlds.Upsert(db, kns); err != nil {
		exitErr(err)
	}

	if cfg.FeatureFlags.AllowSignup {
		lgr.Info("public signup enabled")
	} else {
		lgr.Info("public signup disabled, service key required")
	}

	hcService := &healthcheck.Service{}
	userService := &user.Service{
		DB:          db,
		AllowSignup: cfg.FeatureFlags.AllowSignup,
		ServiceKey:  cfg.Server.ServiceKey,
	}
	socialService := &social.Service{
		DB: db,
	}
	exportService := &export.Service{
		DB:         db,
		Client:     ddrpClient,
		Signer:     signer,
		ServiceKey: cfg.Server.ServiceKey,
	}
	serverLogger := log.WithModule("web")
	r := mux.NewRouter()
	r.Use(web.RequestIDMW(), web.LoggingMW())
	hcService.Mount(r)
	userService.Mount(r)
	socialService.Mount(r)
	exportService.Mount(r)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           nil,
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    1024,
	}
	server.Handler = r

	go func() {
		if err := server.ListenAndServe(); err != nil {
			serverLogger.Error("error starting server", "err", err)
		}
	}()

	lgr.Info("started", "git_commit", version.GitCommit, "listen_port", cfg.Server.Port)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	<-sigs
	server.Close()
	os.Exit(0)
}

func exitErr(err error) {
	fmt.Println(err)
	os.Exit(1)
}
