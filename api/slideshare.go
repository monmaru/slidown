package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/monmaru/slidown/library/log"
	"github.com/monmaru/slidown/service"
)

func HandleSlideShare(slideshare service.SlideShareService, storage service.Storage) http.Handler {
	return &SlideShareHandler{
		slideshare: slideshare,
		storage:    storage,
	}
}

type SlideShareHandler struct {
	slideshare service.SlideShareService
	storage    service.Storage
}

func (h *SlideShareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := parseFrom(r)
	if err != nil {
		log.Errorf(ctx, "failed to parse request : %v", err)
		writeMessage(w, "Invalid request format!!", http.StatusBadRequest)
	}

	slide, err := h.slideshare.Fetch(ctx, req.URL)
	if err != nil {
		log.Errorf(ctx, "GetSlideShareInfo error: %#v", err)
		writeMessage(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	log.Debugf(ctx, "slide: %+v", slide)
	fileName := fmt.Sprint(slide.ID) + "." + slide.Format
	if slide.Download {
		resp, err := httpGetWithTimeout(ctx, slide.DownloadURL, 60*time.Second)
		if err != nil {
			log.Errorf(ctx, "download error: %#v", err)
			writeMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}

		if resp.ContentLength > maxFileSize {
			url, err := h.storage.Upload(ctx, resp.Body, fileName)
			if err != nil {
				log.Errorf(ctx, "upload error %v", err)
				writeMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
				return
			}
			writeMessage(w, url, http.StatusCreated)
		} else {
			log.Debugf(ctx, "content length is %d", resp.ContentLength)
			copyResponseAndClose(ctx, w, resp, fileName)
		}
		return
	}

	links, err := h.slideshare.FetchImageLinks(ctx, req.URL)
	if err != nil {
		log.Errorf(ctx, "GetSlideImageLinks error: %#v", err)
		writeMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
		return
	}

	log.Debugf(ctx, "start write pdf")
	pdf := &pdfWriter{}
	pdf.write(ctx, w, links, fileName)
	log.Debugf(ctx, "finished write pdf")
}

type pdfWriter struct{}

func (p *pdfWriter) write(ctx context.Context, w http.ResponseWriter, links []service.Link, fileName string) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	for i, link := range links {
		imageLink := link.Normal
		resp, err := http.DefaultClient.Get(imageLink)
		if err != nil {
			log.Errorf(ctx, "Error HTTP Get %s idx = %d: %v", imageLink, i, err)
			writeMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}

		if err := p.addPage(pdf, imageLink, resp); err != nil {
			log.Errorf(ctx, "addPage2PDF error idx = %d: %v", i, err)
			writeMessage(w, "PDF作成中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-FileName", fileName)
	if err := pdf.Output(w); err != nil {
		log.Errorf(ctx, "PDF作成中にエラーが発生しました。 %#v", err)
	}
}

func (p *pdfWriter) addPage(pdf *gofpdf.Fpdf, link string, resp *http.Response) error {
	defer resp.Body.Close()
	pdf.AddPage()
	options := gofpdf.ImageOptions{
		ReadDpi:   false,
		ImageType: pdf.ImageTypeFromMime(resp.Header["Content-Type"][0]),
	}
	infoPtr := pdf.RegisterImageOptionsReader(link, options, resp.Body)
	imgWd, imgHt := infoPtr.Extent()
	pdf.ImageOptions(link, 35, 40, imgWd, imgHt, false, options, 0, "")

	if pdf.Err() {
		return pdf.Error()
	}
	return nil
}
