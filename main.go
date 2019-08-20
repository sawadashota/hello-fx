package main

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"go.uber.org/fx"
)

func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})
	return logger
}

func NewRequestLogger(logger *logrus.Logger) *negronilogrus.Middleware {
	return negronilogrus.NewMiddlewareFromLogger(logger, "request-logger")
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.URL.Path))
}

func NewMiddleware(lc fx.Lifecycle, l *logrus.Logger, lmw *negronilogrus.Middleware) *negroni.Negroni {
	mw := negroni.New(lmw)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mw,
	}
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			l.Infoln("Starting HTTP server.")
			go server.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			l.Infoln("Stopping HTTP server.")
			return server.Shutdown(ctx)
		},
	})

	return mw
}

func NewMux() *mux.Router {
	return mux.NewRouter()
}

func Register(router *mux.Router, mw *negroni.Negroni) {
	router.HandleFunc("/", IndexHandler).Methods(http.MethodGet)
	router.HandleFunc("/hello", IndexHandler).Methods(http.MethodGet)

	mw.UseHandler(router)
}

func main() {
	app := fx.New(
		fx.Provide(
			NewLogger,
			NewRequestLogger,
			NewMux,
			NewMiddleware,
		),
		fx.Invoke(Register),
	)

	app.Run()
}
