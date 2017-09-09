package common

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
)

// DecodeJSON ...
func DecodeJSON(rc io.ReadCloser, out interface{}) error {
	defer rc.Close()
	return json.NewDecoder(rc).Decode(out)
}

// DecodeXML ...
func DecodeXML(rc io.ReadCloser, out interface{}) error {
	defer rc.Close()
	return xml.NewDecoder(rc).Decode(out)
}

// TailURL ...
func TailURL(url string) string {
	tmp := strings.Split(url, "/")
	return tmp[len(tmp)-1]
}
