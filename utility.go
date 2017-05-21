package slidown

import (
	"encoding/json"
	"encoding/xml"
	"io"
)

func decodeJSON(rc io.ReadCloser, out interface{}) error {
	defer rc.Close()
	return json.NewDecoder(rc).Decode(out)
}

func decodeXML(rc io.ReadCloser, out interface{}) error {
	defer rc.Close()
	return xml.NewDecoder(rc).Decode(out)
}
