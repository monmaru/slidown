package slidown

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

// HandlerWithContext ...
type HandlerWithContext func(context.Context, http.ResponseWriter, *http.Request)

// HTTPWithContext ...
func HTTPWithContext(h HandlerWithContext) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)
		defer timeTrack(ctx, time.Now(), r.RequestURI)
		h(ctx, w, r)
	})
}

// NewContext ...
var NewContext func(*http.Request) context.Context = func(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func timeTrack(ctx context.Context, start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debugf(ctx, "%s took %s", name, elapsed)
}

// HandlerWithReqData ...
type HandlerWithReqData func(context.Context, http.ResponseWriter, *ReqData)

// MustParams ...
func MustParams(h HandlerWithReqData) HandlerWithContext {
	return HandlerWithContext(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		data, err := json2ReqData(r.Body)
		if err != nil || data.URL == "" {
			log.Infof(ctx, "Error json2ReqData : %v", err)
			writeErrorResponse(w, "Invalid request format!!", http.StatusBadRequest)
			return
		}
		h(ctx, w, data)
	})
}

// DownloadFromSlideShare ...
func DownloadFromSlideShare(ctx context.Context, w http.ResponseWriter, data *ReqData) {
	httpClient := NewHTTPClient(ctx)
	svc := NewSlideShareSvc(
		os.Getenv("APIKEY"),
		os.Getenv("SHAREDSECRET"),
		httpClient)

	slide, err := svc.GetSlideShareInfo(data.URL)
	if err != nil {
		log.Infof(ctx, "GetSlideShareInfo error: %#v", err)
		writeErrorResponse(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	fileName := fmt.Sprint(slide.ID) + "." + slide.Format
	if slide.Download {
		resp, err := getWithTimeout(ctx, slide.DownloadURL)
		if err != nil {
			log.Errorf(ctx, "download error: %#v", err)
			writeErrorResponse(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}
		copyResponseAndClose(w, resp, fileName)
		return
	}

	createPDF, _ := strconv.ParseBool(os.Getenv("CREATEPDF"))
	if !createPDF {
		msg := "指定されたスライドはダウンロードが禁止されています。"
		log.Infof(ctx, msg)
		writeErrorResponse(w, msg, http.StatusBadRequest)
		return
	}

	links, err := svc.GetSlideImageLinks(data.URL)
	if err != nil {
		log.Errorf(ctx, "GetSlideImageLinks error: %#v", err)
		writeErrorResponse(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
		return
	}

	pdf := gofpdf.New("L", "mm", "A4", "")
	for i, link := range links {
		imageLink := link.Normal
		resp, err := httpClient.Get(imageLink)
		if err != nil {
			log.Errorf(ctx, "Error HTTP Get %s idx = %d: %v", imageLink, i, err)
			writeErrorResponse(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}

		if err := addPage2PDF(pdf, imageLink, resp); err != nil {
			log.Errorf(ctx, "addPage2PDF error idx = %d: %v", i, err)
			writeErrorResponse(w, "PDF作成中にエラーが発生しました。", http.StatusInternalServerError)
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

// DownloadFromSpeakerDeck ...
func DownloadFromSpeakerDeck(ctx context.Context, w http.ResponseWriter, data *ReqData) {
	svc := NewSpeakerDeckSvc(NewHTTPClient(ctx))
	info, err := svc.GetSpeakerDeckInfo(data.URL)
	if err != nil {
		log.Infof(ctx, "GetSpeakerDeckInfo error: %#v", err)
		writeErrorResponse(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	resp, err := getWithTimeout(ctx, info.DownloadURL)
	if err != nil {
		log.Errorf(ctx, "download error: %#v", err)
		writeErrorResponse(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
		return
	}

	copyResponseAndClose(w, resp, info.FileName)
}

// NewHTTPClient ...
var NewHTTPClient func(ctx context.Context) *http.Client = func(ctx context.Context) *http.Client {
	return urlfetch.Client(ctx)
}

func json2ReqData(rc io.ReadCloser) (*ReqData, error) {
	defer rc.Close()
	var data ReqData
	err := json.NewDecoder(rc).Decode(&data)
	return &data, err
}

func getWithTimeout(ctx context.Context, url string) (resp *http.Response, err error) {
	const timeout = 60 * time.Second
	ctxWithTimeout, _ := context.WithTimeout(ctx, timeout)
	return NewHTTPClient(ctxWithTimeout).Get(url)
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ResError{Message: message})
}

func copyResponseAndClose(w http.ResponseWriter, resp *http.Response, fileName string) {
	defer resp.Body.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	w.Header().Set("X-FileName", fileName)
	io.Copy(w, resp.Body)
}
