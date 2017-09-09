package slideshare

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/monmaru/slidown/common"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type service struct {
	APIKey, SharedSecret string
	httpClient           *http.Client
}

func newService(apiKey, sharedSecret string, client *http.Client) *service {
	return &service{
		APIKey:       apiKey,
		SharedSecret: sharedSecret,
		httpClient:   client,
	}
}

func (s *service) fetch(ctx context.Context, url string, optArgs ...map[string]string) (*Slide, error) {
	var cache Slide
	key := "SS-" + common.TailURL(url)
	if err := common.GetCache(ctx, key, &cache); err == nil {
		return &cache, nil
	}

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

	slide := &Slide{}
	err = common.DecodeXML(resp.Body, slide)

	// compare with zero value.
	if slide.ID == 0 && slide.Title == "" {
		return nil, errors.New("指定されたURLからはスライドが見つかりませんでした。")
	}

	common.SetCache(ctx, key, slide)
	return slide, err
}

// fetchImageLinks ...
func (s *service) fetchImageLinks(url string) ([]link, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	links := []link{}
	doc.Find("div.slide_container > section > img.slide_image").Each(func(i int, s *goquery.Selection) {
		normal := strings.Split(s.AttrOr("data-normal", ""), "?")[0]
		full := strings.Split(s.AttrOr("data-full", ""), "?")[0]
		link := link{
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

func (s *service) buildEndpoint(method string, args map[string]string) string {
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
