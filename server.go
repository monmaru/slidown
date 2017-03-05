package slidown

import (
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/api/speakerdeck/download", HttpWithContext(MustParams(DownloadFromSpeakerDeck))).Methods("POST")
	router.HandleFunc("/api/slideshare/download", HttpWithContext(MustParams(DownloadFromSlideShare))).Methods("POST")
	http.Handle("/", router)
}

