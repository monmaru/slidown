package slideshare

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jung-kurt/gofpdf"
	"github.com/monmaru/slidown/common"
	"google.golang.org/appengine/log"
)

// DownloadHandler ...
func DownloadHandler(ctx context.Context, w http.ResponseWriter, data common.ReqData) {
	httpClient := common.DefaultHTTPClient(ctx)
	svc := newService(
		os.Getenv("APIKEY"),
		os.Getenv("SHAREDSECRET"),
		httpClient)

	slide, err := svc.fetch(ctx, data.URL)
	if err != nil {
		log.Infof(ctx, "GetSlideShareInfo error: %#v", err)
		common.WriteMessage(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	log.Debugf(ctx, "slide: %+v", slide)
	fileName := fmt.Sprint(slide.ID) + "." + slide.Format
	if slide.Download {
		resp, err := common.GetWithTimeout(ctx, slide.DownloadURL)
		if err != nil {
			log.Errorf(ctx, "download error: %#v", err)
			common.WriteMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}

		if resp.ContentLength > common.MaxFileSize {
			url, err := common.Upload2GCS(ctx, resp.Body, fileName)
			if err != nil {
				log.Errorf(ctx, "Upload2GCS error %v", err)
				common.WriteMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
				return
			}
			common.WriteMessage(w, url, http.StatusCreated)
		} else {
			common.CopyResponseAndClose(ctx, w, resp, fileName)
		}
		return
	}

	createPDF, _ := strconv.ParseBool(os.Getenv("CREATEPDF"))
	if !createPDF {
		msg := "指定されたスライドはダウンロードが禁止されています。"
		log.Infof(ctx, msg)
		common.WriteMessage(w, msg, http.StatusBadRequest)
		return
	}

	links, err := svc.fetchImageLinks(data.URL)
	if err != nil {
		log.Errorf(ctx, "GetSlideImageLinks error: %#v", err)
		common.WriteMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
		return
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	for i, link := range links {
		imageLink := link.Normal
		resp, err := httpClient.Get(imageLink)
		if err != nil {
			log.Errorf(ctx, "Error HTTP Get %s idx = %d: %v", imageLink, i, err)
			common.WriteMessage(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}

		if err := addPage2PDF(pdf, imageLink, resp); err != nil {
			log.Errorf(ctx, "addPage2PDF error idx = %d: %v", i, err)
			common.WriteMessage(w, "PDF作成中にエラーが発生しました。", http.StatusInternalServerError)
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

func addPage2PDF(pdf *gofpdf.Fpdf, link string, resp *http.Response) error {
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
