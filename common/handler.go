package common

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const MaxFileSize = 29360128 // 28MB

// HandlerWithContext ...
type HandlerWithContext func(context.Context, http.ResponseWriter, *http.Request)

// HTTPWithContext ...
func HTTPWithContext(h HandlerWithContext) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := NewContext(r)
		defer timeTrack(ctx, time.Now(), r.RequestURI)
		h(ctx, w, r)
	}
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
type HandlerWithReqData func(context.Context, http.ResponseWriter, ReqData)

// MustParams ...
func MustParams(h HandlerWithReqData) HandlerWithContext {
	return HandlerWithContext(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var data ReqData
		err := DecodeJSON(r.Body, &data)
		if err != nil || data.URL == "" {
			log.Infof(ctx, "Error json2ReqData : %v", err)
			WriteMessage(w, "Invalid request format!!", http.StatusBadRequest)
			return
		}
		h(ctx, w, data)
	})
}

// GetWithTimeout ...
func GetWithTimeout(ctx context.Context, url string) (resp *http.Response, err error) {
	const timeout = 60 * time.Second
	ctxWithTimeout, _ := context.WithTimeout(ctx, timeout)
	return CustomHTTPClient(ctxWithTimeout).Get(url)
}

// WriteMessage ...
func WriteMessage(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ResMessage{Message: message})
}

// CopyResponseAndClose ...
func CopyResponseAndClose(ctx context.Context, w http.ResponseWriter, resp *http.Response, fileName string) {
	defer resp.Body.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", string(resp.ContentLength))
	w.Header().Set("X-FileName", fileName)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Errorf(ctx, "io.Copy error: %#v", err)
	}
}
