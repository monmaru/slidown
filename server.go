package slidown

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine/log"
)

func init() {
	router := httprouter.New()
	router.GET("/_ah/start", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Infof(NewContext(r), "START")
	})
	router.GET("/_ah/stop", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Infof(NewContext(r), "STOP")
	})
	router.POST("/api/speakerdeck/download", HTTPWithContext(MustParams(DownloadFromSpeakerDeck)))
	router.POST("/api/slideshare/download", HTTPWithContext(MustParams(DownloadFromSlideShare)))
	http.Handle("/", router)
}
