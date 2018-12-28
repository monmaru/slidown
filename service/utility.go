package service

import (
	"encoding/xml"
	"io"
	"strings"
)

func decodeXML(rc io.ReadCloser, out interface{}) error {
	defer rc.Close()
	return xml.NewDecoder(rc).Decode(out)
}

func tailURL(url string) string {
	tmp := strings.Split(url, "/")
	return tmp[len(tmp)-1]
}
