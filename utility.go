package slidown

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
)

func decodeJSON(rc io.ReadCloser, out interface{}) error {
	defer rc.Close()
	return json.NewDecoder(rc).Decode(out)
}

func decodeXML(rc io.ReadCloser, out interface{}) error {
	defer rc.Close()
	return xml.NewDecoder(rc).Decode(out)
}

func tailURL(url string) string {
	tmp := strings.Split(url, "/")
	return tmp[len(tmp)-1]
}
