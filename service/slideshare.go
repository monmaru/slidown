package service

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SlideShareSlide struct {
	ID                uint32 `xml:"ID"`
	Title             string `xml:"Title"`
	Description       string `xml:"Description"`
	Username          string `xml:"Username"`
	Status            uint8  `xml:"Status"`
	URL               string `xml:"URL"`
	ThumbnailURL      string `xml:"ThumbnailURL"`
	ThumbnailSize     string `xml:"ThumbnailSize"`
	ThumbnailSmallURL string `xml:"ThumbnailSmallURL"`
	Embed             string `xml:"Embed"`
	Created           string `xml:"Created"`
	Updated           string `xml:"Updated"`
	Language          string `xml:"Language"`
	Format            string `xml:"Format"`
	Download          bool   `xml:"Download"`
	DownloadURL       string `xml:"DownloadUrl"`
	SlideshowType     uint8  `xml:"SlideshowType"`
	InContest         bool   `xml:"InContest"`
}

type Link struct {
	Full, Normal string
}

type SlideShareService interface {
	Fetch(ctx context.Context, url string, optArgs ...map[string]string) (*SlideShareSlide, error)
	FetchImageLinks(url string) ([]Link, error)
}

type SlideShareServiceImpl struct {
	APIKey, SharedSecret string
	httpClient           *http.Client
}

func NewSlideShareService(apiKey, sharedSecret string, client *http.Client) SlideShareService {
	return &SlideShareServiceImpl{
		APIKey:       apiKey,
		SharedSecret: sharedSecret,
		httpClient:   client,
	}
}

func (s *SlideShareServiceImpl) Fetch(ctx context.Context, url string, optArgs ...map[string]string) (*SlideShareSlide, error) {
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

	slide := &SlideShareSlide{}
	err = decodeXML(resp.Body, slide)

	// compare with zero value.
	if slide.ID == 0 && slide.Title == "" {
		return nil, errors.New("指定されたURLからはスライドが見つかりませんでした。")
	}

	return slide, err
}

func (s *SlideShareServiceImpl) FetchImageLinks(url string) ([]Link, error) {
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

func (s *SlideShareServiceImpl) buildEndpoint(method string, args map[string]string) string {
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
