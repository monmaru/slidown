package slidown

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

type SlideShareSvc struct {
	APIKey, SharedSecret string
	httpClient           *http.Client
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

	// compare with zero value.
	if ss.ID == 0 && ss.Title == "" {
		return nil, errors.New("指定されたURLからはスライドが見つかりませんでした。")
	}

	return ss, err
}

func (s *SlideShareSvc) GetSlideImageLinks(url string) ([]Link, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	links := []Link{}
	doc.Find("div.slide_container > section > img.slide_image").Each(func(i int, s *goquery.Selection) {
		normal := strings.Split(s.AttrOr("data-normal", ""), "?")[0]
		full := strings.Split(s.AttrOr("data-full", ""), "?")[0]
		link := Link{
			Normal: normal,
			Full:   full,
		}
		links = append(links, link)
	})

	if len(links) == 0 {
		return nil, errors.New("指定されたURLからはスライドが見つかりませんでした。")
	}

	return links, nil
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
	baseURL := "https://www.slideshare.net/api/2"
	return baseURL + "/" + method + "?" + values.Encode()
}

func parseXML(resp *http.Response, container interface{}) error {
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal([]byte(bytes), container)
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
	downloadURL := doc.Find("#share_pdf").AttrOr("href", "")
	strings.Split(downloadURL, "/")

	// compare with zero value.
	if downloadURL == "" {
		return nil, errors.New("指定されたURLからはスライドが見つかりませんでした。")
	}

	return &SpeakerDeckInfo{
		Title:       details.Find("h1").Text(),
		Description: strings.TrimSpace(details.Find(".description").Text()),
		DownloadURL: downloadURL,
		FileName:    extractFileName(downloadURL),
	}, nil
}

func extractFileName(downloadURL string) string {
	tmp := strings.Split(downloadURL, "/")
	return tmp[len(tmp)-1]
}
