package slidown

import (
	"encoding/json"
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
		log.Errorf(ctx, "Error json2Data : %v", err)
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	svc := createSvc(ctx)

	id, err := svc.GetSlideID(data.URL)
	if err != nil {
		log.Errorf(ctx, "GetSlideID error: %#v", err)
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slide, err := svc.GetSlideShareInfo(id)
	if err != nil {
		log.Errorf(ctx, "GetSlideShareInfo error: %#v", err)
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if slide.Download {
		resp, err := download(ctx, slide.DownloadURL)
		if err != nil {
			log.Errorf(ctx, "download error: %#v", err)
			writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		fileName := id + "." + slide.Format
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("X-FileName", fileName)
		io.Copy(w, resp.Body)
	} else {
		// TODO: ダウンロード不可のスライドの場合は、画像ファイルをもとにPDFを作成して返す。
		writeErrorResponse(w, "このスライドはダウンロード禁止です。", http.StatusBadRequest)
	}
}

func createSvc(ctx context.Context) *SlideShareSvc {
	return NewSlideShareSvc(
		os.Getenv("APIKEY"),
		os.Getenv("SHAREDSECRET"),
		getDefaultHTTPClient(ctx))
}

func DownloadFromSpeakerDeck(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	data, err := json2ReqData(r.Body)
	if err != nil {
		log.Errorf(ctx, "Error json2Data : %v", err)
		writeErrorResponse(w, "Invalid json format", http.StatusBadRequest)
		return
	}

	service := NewSpeakerDeckSvc(getDefaultHTTPClient(ctx))
	info, err := service.GetSpeakerDeckInfo(data.URL)
	if err != nil {
		log.Errorf(ctx, "GetSpeakerDeckInfo error: %#v", err)
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := download(ctx, info.DownloadURL)
	if err != nil {
		log.Errorf(ctx, "download error: %#v", err)
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+info.FileName)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("X-FileName", info.FileName)
	io.Copy(w, resp.Body)
}

func getDefaultHTTPClient(ctx context.Context) *http.Client {
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
	return urlfetch.Client(ctxWithDeadline).Get(url)
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ResError{Message: message})
}
