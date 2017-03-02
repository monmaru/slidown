package slidown

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func DownloadFromSlideShare(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	data, err := json2ReqData(r.Body)
	if err != nil {
		log.Debugf(ctx, "Error json2Data : %v", err)
		writeErrorResponse(w, "Invalid request format!!", http.StatusBadRequest)
		return
	}

	svc := createSlideShareSvc(ctx)
	slide, err := svc.GetSlideShareInfo(data.URL)
	if err != nil {
		log.Debugf(ctx, "GetSlideShareInfo error: %#v", err)
		writeErrorResponse(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	if slide.Download {
		resp, err := download(ctx, slide.DownloadURL)
		if err != nil {
			log.Errorf(ctx, "download error: %#v", err)
			writeErrorResponse(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		fileName := fmt.Sprint(slide.ID) + "." + slide.Format
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		w.Header().Set("X-FileName", fileName)
		io.Copy(w, resp.Body)
	} else {
		// TODO: ダウンロード不可のスライドの場合は、画像ファイルをもとにPDFを作成して返す。
		msg := "指定されたスライドはダウンロード禁止です。"
		log.Debugf(ctx, msg)
		writeErrorResponse(w, msg, http.StatusBadRequest)
	}
}

func createSlideShareSvc(ctx context.Context) *SlideShareSvc {
	return NewSlideShareSvc(
		os.Getenv("APIKEY"),
		os.Getenv("SHAREDSECRET"),
		GetHTTPClient(ctx))
}

func DownloadFromSpeakerDeck(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	data, err := json2ReqData(r.Body)
	if err != nil {
		log.Debugf(ctx, "Error json2Data : %v", err)
		writeErrorResponse(w, "Invalid request format!!", http.StatusBadRequest)
		return
	}

	svc := NewSpeakerDeckSvc(GetHTTPClient(ctx))
	info, err := svc.GetSpeakerDeckInfo(data.URL)
	if err != nil {
		log.Debugf(ctx, "GetSpeakerDeckInfo error: %#v", err)
		writeErrorResponse(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	resp, err := download(ctx, info.DownloadURL)
	if err != nil {
		log.Errorf(ctx, "download error: %#v", err)
		writeErrorResponse(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+info.FileName)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	w.Header().Set("X-FileName", info.FileName)
	io.Copy(w, resp.Body)
}

var GetHTTPClient func(ctx context.Context) *http.Client = func(ctx context.Context) *http.Client {
	return urlfetch.Client(ctx)
}

func json2ReqData(rc io.ReadCloser) (*ReqData, error) {
	defer rc.Close()
	var data ReqData
	err := json.NewDecoder(rc).Decode(&data)
	return &data, err
}

func download(ctx context.Context, url string) (resp *http.Response, err error) {
	ctxWithDeadline, _ := context.WithTimeout(ctx, 60*time.Second)
	return GetHTTPClient(ctxWithDeadline).Get(url)
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ResError{Message: message})
}
