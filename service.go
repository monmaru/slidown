package slidown

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://www.slideshare.net/api/2"

type SlideShareSvc struct {
	APIKey       string
	SharedSecret string
	httpClient   *http.Client
}

func NewSlideShareSvc(apiKey, sharedSecret string, client *http.Client) *SlideShareSvc {
	return &SlideShareSvc{
		APIKey:       apiKey,
		SharedSecret: sharedSecret,
		httpClient:   client,
	}
}

func (s *SlideShareSvc) GetSlideShareInfo(url string, optArgs ...map[string]string) (*SlideShareInfo, error) {
	args := map[string]string{
		"slideshow_url": url,
	}

	if len(optArgs) > 0 {
		for k, v := range optArgs[0] {
			args[k] = v
		}
	}
	endpoint := s.buildEndpoint("get_slideshow", args)

	resp, err := s.httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ss := &SlideShareInfo{}
	err = parseXML(resp, ss)
	return ss, err
}

func (s *SlideShareSvc) buildEndpoint(method string, args map[string]string) string {
	values := url.Values{}
	for k, v := range args {
		values.Set(k, v)
	}
	values.Set("api_key", s.APIKey)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	values.Set("ts", timestamp)
	hash := sha1.New()
	io.WriteString(hash, s.SharedSecret+timestamp)
	values.Set("hash", fmt.Sprintf("%x", hash.Sum(nil)))
	return baseURL + "/" + method + "?" + values.Encode()
}

func parseXML(resp *http.Response, container interface{}) error {
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal([]byte(responseBody), container)
}

type SpeakerDeckSvc struct {
	httpClient *http.Client
}

func NewSpeakerDeckSvc(client *http.Client) *SpeakerDeckSvc {
	return &SpeakerDeckSvc{httpClient: client}
}

func (s *SpeakerDeckSvc) GetSpeakerDeckInfo(url string) (*SpeakerDeckInfo, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	details := doc.Find("#talk-details")
	sidebar := doc.Find(".sidebar")
	downloadURL := doc.Find("#share_pdf").AttrOr("href", "")
	strings.Split(downloadURL, "/")
	stars, err := strconv.Atoi(strings.Split(sidebar.Find(".stargazers").Text(), " ")[0])
	if err != nil {
		return nil, err
	}

	return &SpeakerDeckInfo{
		Title:       details.Find("h1").Text(),
		Description: strings.TrimSpace(details.Find(".description").Text()),
		DownloadURL: downloadURL,
		FileName:    extractFileName(downloadURL),
		Stars:       stars,
	}, nil
}

func extractFileName(downloadURL string) string {
	tmp := strings.Split(downloadURL, "/")
	return tmp[len(tmp)-1]
}
