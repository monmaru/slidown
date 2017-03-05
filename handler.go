package slidown

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type HandlerWithContext func(context.Context, http.ResponseWriter, *http.Request)

func HttpWithContext(h HandlerWithContext) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r)
		defer timeTrack(ctx, time.Now(), r.RequestURI)
		h(ctx, w, r)
	})
}

var NewContext func(*http.Request) context.Context = func(r *http.Request) context.Context {
	return appengine.NewContext(r)
}

func timeTrack(ctx context.Context, start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debugf(ctx, "%s took %s", name, elapsed)
}

type HandlerWithReqData func(context.Context, http.ResponseWriter, *ReqData)

func MustParams(h HandlerWithReqData) HandlerWithContext {
	return HandlerWithContext(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		data, err := json2ReqData(r.Body)
		if err != nil || data.URL == "" {
			log.Debugf(ctx, "Error json2Data : %v", err)
			writeError(w, "Invalid request format!!", http.StatusBadRequest)
			return
		}
		h(ctx, w, data)
	})
}

func DownloadFromSlideShare(ctx context.Context, w http.ResponseWriter, data *ReqData) {
	svc := createSlideShareSvc(ctx)
	slide, err := svc.GetSlideShareInfo(data.URL)
	if err != nil {
		log.Infof(ctx, "GetSlideShareInfo error: %#v", err)
		writeError(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	if slide.Download {
		resp, err := download(ctx, slide.DownloadURL)
		if err != nil {
			log.Errorf(ctx, "download error: %#v", err)
			writeError(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		fileName := fmt.Sprint(slide.ID) + "." + slide.Format
		copyResponseAsFile(w, resp, fileName)
	} else {
		// TODO: ダウンロード不可のスライドの場合は、画像ファイルをもとにPDFを作成して返す。
		msg := "指定されたスライドはダウンロード禁止です。"
		log.Infof(ctx, msg)
		writeError(w, msg, http.StatusBadRequest)
	}
}

func createSlideShareSvc(ctx context.Context) *SlideShareSvc {
	return NewSlideShareSvc(
		os.Getenv("APIKEY"),
		os.Getenv("SHAREDSECRET"),
		NewHTTPClient(ctx))
}

func DownloadFromSpeakerDeck(ctx context.Context, w http.ResponseWriter, data *ReqData) {
	svc := NewSpeakerDeckSvc(NewHTTPClient(ctx))
	info, err := svc.GetSpeakerDeckInfo(data.URL)
	if err != nil {
		log.Infof(ctx, "GetSpeakerDeckInfo error: %#v", err)
		writeError(w, "スライドが見つかりませんでした。", http.StatusNotFound)
		return
	}

	resp, err := download(ctx, info.DownloadURL)
	if err != nil {
		log.Errorf(ctx, "download error: %#v", err)
		writeError(w, "ダウンロード中にエラーが発生しました。", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	copyResponseAsFile(w, resp, info.FileName)
}

var NewHTTPClient func(ctx context.Context) *http.Client = func(ctx context.Context) *http.Client {
	return urlfetch.Client(ctx)
}

func json2ReqData(rc io.ReadCloser) (*ReqData, error) {
	defer rc.Close()
	var data ReqData
	err := json.NewDecoder(rc).Decode(&data)
	return &data, err
}

func download(ctx context.Context, url string) (resp *http.Response, err error) {
	ctxWithTimeout, _ := context.WithTimeout(ctx, 60*time.Second)
	return NewHTTPClient(ctxWithTimeout).Get(url)
}

func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ResError{Message: message})
}

func copyResponseAsFile(w http.ResponseWriter, resp *http.Response, fileName string) {
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	w.Header().Set("X-FileName", fileName)
	io.Copy(w, resp.Body)
}
