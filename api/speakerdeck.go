package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/monmaru/slidown/library/log"
	"github.com/monmaru/slidown/service"
)

func HandleSpeakerDeck(speakerdeck service.SpeakerDeckService, storage service.Storage) http.Handler {
	return &SpeakerDeckHandler{
		speakerdeck: speakerdeck,
		storage:     storage,
	}
}

type SpeakerDeckHandler struct {
	speakerdeck service.SpeakerDeckService
	storage     service.Storage
}

func (h *SpeakerDeckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := parseFrom(r)
	if err != nil {
		log.Infof(ctx, "failed to parse request : %v", err)
		writeMessage(w, "Invalid request format!!", http.StatusBadRequest)
	}

	slide, err := h.speakerdeck.Fetch(ctx, strings.Split(req.URL, "?")[0])
	if err != nil {
		log.Errorf(ctx, "fetch error: %#v", err)
		writeMessage(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}
	log.Debugf(ctx, "slide: %+v", slide)

	resp, err := httpGetWithTimeout(ctx, slide.DownloadURL, 60*time.Second)
	if err != nil {
		log.Errorf(ctx, "download error: %#v", err)
		writeMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
		return
	}

	if resp.ContentLength > maxFileSize {
		url, err := h.storage.Upload(ctx, resp.Body, slide.FileName)
		if err != nil {
			log.Errorf(ctx, "Upload error %v", err)
			writeMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}
		writeMessage(w, url, http.StatusCreated)
	} else {
		log.Debugf(ctx, "content length is %d", resp.ContentLength)
		copyResponseAndClose(ctx, w, resp, slide.FileName)
	}
}
