package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/go-chi/chi"
	"github.com/monmaru/slidown/handler"
	mylog "github.com/monmaru/slidown/library/log"
	"github.com/monmaru/slidown/service"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

func main() {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	if err := mylog.Init(os.Getenv("IS_LOCAL") != ""); err != nil {
		log.Fatal(err)
	}
	defer mylog.Close()

	speakerdeck := service.NewSpeakerDeckService()

	slideshare := service.NewSlideShareService(
		os.Getenv("APIKEY"),
		os.Getenv("SHAREDSECRET"))
	storage := service.NewStorage(os.Getenv("BUCKETNAME"))

	router := route(slideshare, speakerdeck, storage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
		Handler: &ochttp.Handler{
			Handler:     router,
			Propagation: &propagation.HTTPFormat{},
		},
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func route(
	slideshare service.SlideShareService,
	speakerdeck service.SpeakerDeckService,
	storage service.Storage) http.Handler {

	router := chi.NewRouter()

	// middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := mylog.WithContext(r.Context(), r)
			defer mylog.Duration(ctx, time.Now(), r.URL.Path)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	router.Get("/_ah/start", func(w http.ResponseWriter, r *http.Request) {
		mylog.Infof(r.Context(), "on /_ah/start")
	})
	router.Get("/_ah/stop", func(w http.ResponseWriter, r *http.Request) {
		mylog.Infof(r.Context(), "on /_ah/stop")
	})

	router.Route("/api", func(r chi.Router) {
		r.Method("POST", "/speakerdeck/download", handler.SpeakerDeck(speakerdeck, storage))
		r.Method("POST", "/slideshare/download", handler.SlideShare(slideshare, storage))
	})
	return router
}
