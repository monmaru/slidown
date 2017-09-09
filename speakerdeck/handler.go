package speakerdeck

import (
	"net/http"

	"github.com/monmaru/slidown/common"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// DownloadHandler ...
func DownloadHandler(ctx context.Context, w http.ResponseWriter, data common.ReqData) {
	s := newService(common.DefaultHTTPClient(ctx))
	slide, err := s.fetch(ctx, data.URL)
	if err != nil {
		log.Infof(ctx, "fetch error: %#v", err)
		common.WriteMessage(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	resp, err := common.GetWithTimeout(ctx, slide.DownloadURL)
	if err != nil {
		log.Errorf(ctx, "download error: %#v", err)
		common.WriteMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
		return
	}

	if resp.ContentLength > common.MaxFileSize {
		url, err := common.Upload2GCS(ctx, resp.Body, slide.FileName)
		if err != nil {
			log.Errorf(ctx, "Upload2GCS error %v", err)
			common.WriteMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}
		common.WriteMessage(w, url, http.StatusCreated)
	} else {
		common.CopyResponseAndClose(ctx, w, resp, slide.FileName)
	}
}
