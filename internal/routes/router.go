package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

type CacheOperator interface {
	GetKey(w http.ResponseWriter, r *http.Request)
	SetKey(w http.ResponseWriter, r *http.Request)
	DeleteRate(w http.ResponseWriter, r *http.Request)
	CacheSize(w http.ResponseWriter, r *http.Request)
	CheckRate(w http.ResponseWriter, r *http.Request)
	ClearAndRefresh(w http.ResponseWriter, r *http.Request)
	TTLKey(w http.ResponseWriter, r *http.Request)
}

type CurrencyConverter interface {
	GetRates(w http.ResponseWriter, r *http.Request)
	Convert(w http.ResponseWriter, r *http.Request)
}

type CurrencyLister interface {
	List(w http.ResponseWriter, r *http.Request)
}

type HealthChecker interface {
	CheckHealth(w http.ResponseWriter, r *http.Request)
}

type RateLimiter interface {
	Limit(next http.Handler) http.Handler
	Start(interval time.Duration)
	Stop()
}

type RouterConfig struct {
	CacheOperator       CacheOperator
	CurrencyLister      CurrencyLister
	HealthChecker       HealthChecker
	CurrencyConverter   CurrencyConverter
	RateLimiter         RateLimiter
	APIKeyAuthenticator func(http.Handler) http.Handler
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	if cfg.RateLimiter != nil {
		r.Use(cfg.RateLimiter.Limit)
	}

	r.Get("/health", cfg.HealthChecker.CheckHealth)
	r.Get("/rates", cfg.CurrencyConverter.GetRates)
	r.Get("/currencies", cfg.CurrencyLister.List)
	r.Get("/convert", cfg.CurrencyConverter.Convert)

	r.Route("/admin/cache", func(r chi.Router) {

		r.Use(cfg.APIKeyAuthenticator)
		r.Get("/size", cfg.CacheOperator.CacheSize)
		r.Get("/get", cfg.CacheOperator.GetKey)
		r.Get("/set", cfg.CacheOperator.SetKey)
		r.Get("/delete", cfg.CacheOperator.DeleteRate)
		r.Post("/clear", cfg.CacheOperator.ClearAndRefresh)
		r.Get("/check", cfg.CacheOperator.CheckRate)
		r.Get("/ttl", cfg.CacheOperator.TTLKey)
	})

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Wrap with CORS
	handler := cors.Default().Handler(r)
	return handler
}
