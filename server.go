package slidown

import (
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/api/speakerdeck/download", httpWithContext(DownloadFromSpeakerDeck)).Methods("POST")
	router.HandleFunc("/api/slideshare/download", httpWithContext(DownloadFromSlideShare)).Methods("POST")
	http.Handle("/", router)
}

type HandlerWithContext func(context.Context, http.ResponseWriter, *http.Request)

func httpWithContext(h HandlerWithContext) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h(NewContext(r), w, r)
	})
}

var NewContext func(*http.Request) context.Context = func(r *http.Request) context.Context {
	return appengine.NewContext(r)
}
