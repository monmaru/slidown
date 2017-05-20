package slidown

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func init() {
	router := httprouter.New()
	router.GET("/_ah/start", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.GET("/_ah/stop", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {})
	router.POST("/api/speakerdeck/download", HTTPWithContext(MustParams(DownloadFromSpeakerDeck)))
	router.POST("/api/slideshare/download", HTTPWithContext(MustParams(DownloadFromSlideShare)))
	http.Handle("/", router)
}
