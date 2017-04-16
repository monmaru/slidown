package slidown

import (
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/api/speakerdeck/download", HTTPWithContext(MustParams(DownloadFromSpeakerDeck))).Methods("POST")
	router.HandleFunc("/api/slideshare/download", HTTPWithContext(MustParams(DownloadFromSlideShare))).Methods("POST")
	http.Handle("/", router)
}
