package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/monmaru/slidown/library/log"
	"github.com/monmaru/slidown/model"
)

const maxFileSize = 29360128 // 28MB

func parseFrom(r *http.Request) (*model.Request, error) {
	var req model.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func httpGetWithTimeout(ctx context.Context, url string, timeout time.Duration) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	ctxWithTimeout, _ := context.WithTimeout(ctx, timeout)
	req = req.WithContext(ctxWithTimeout)
	return http.DefaultClient.Do(req)
}

func writeMessage(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(model.Response{Message: message})
}

func copyResponseAndClose(ctx context.Context, w http.ResponseWriter, resp *http.Response, fileName string) {
	defer resp.Body.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	//	w.Header().Set("Content-Length", string(resp.ContentLength))
	w.Header().Set("X-FileName", fileName)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Errorf(ctx, "io.Copy error: %#v", err)
	}
}
