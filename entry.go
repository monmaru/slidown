package slidown

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/monmaru/slidown/common"
	"github.com/monmaru/slidown/slideshare"
	"github.com/monmaru/slidown/speakerdeck"
	"google.golang.org/appengine/log"
)

// Run ...
func Run() {
	router := httprouter.New()
	router.GET("/_ah/start", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Infof(common.NewContext(r), "START")
	})
	router.GET("/_ah/stop", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		log.Infof(common.NewContext(r), "STOP")
	})
	router.POST("/api/speakerdeck/download", common.HTTPWithContext(common.MustParams(speakerdeck.DownloadHandler)))
	router.POST("/api/slideshare/download", common.HTTPWithContext(common.MustParams(slideshare.DownloadHandler)))
	http.Handle("/", router)
}
