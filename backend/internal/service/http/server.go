package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/config"
	"github.com/durianpay/fullstack-boilerplate/internal/middleware"
	authUsecase "github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	oapinethttpmw "github.com/oapi-codegen/nethttp-middleware"
)

type Server struct {
	router http.Handler
}

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second
)

var publicAPIPaths = map[string]bool{
	"/dashboard/v1/auth/login": true,
	"/openapi.json":            true,
	"/swagger":                 true,
}

func NewServer(apiHandler openapigen.ServerInterface, authUC authUsecase.AuthUsecase) *Server {
	swagger, err := openapigen.GetSwagger()
	if err != nil {
		log.Fatalf("failed to load swagger: %v", err)
	}
	swagger.Servers = nil

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.CorsAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/openapi.json", openapiJSONHandler())
	r.Get("/swagger", swaggerUIHandler())

	r.Group(func(api chi.Router) {
		api.Use(oapinethttpmw.OapiRequestValidatorWithOptions(
			swagger,
			&oapinethttpmw.Options{
				DoNotValidateServers:  true,
				SilenceServersWarning: true,
				Options: openapi3filter.Options{
					AuthenticationFunc: func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
						return nil
					},
				},
			},
		))
		api.Use(conditionalAuth(authUC))
		openapigen.HandlerFromMux(apiHandler, api)
	})

	return &Server{router: r}
}

func conditionalAuth(uc authUsecase.AuthUsecase) func(http.Handler) http.Handler {
	auth := middleware.Auth(uc)
	return func(next http.Handler) http.Handler {
		protected := auth(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if publicAPIPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}

func (s *Server) Start(addr string) {
	service := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	go func() {
		log.Printf("listening on %s", addr)
		if err := service.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	const shutdownTimeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := service.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func (s *Server) Routes() http.Handler {
	return s.router
}
