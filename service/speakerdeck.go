package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SpeakerDeckService interface {
	Fetch(ctx context.Context, url string) (*SpeakerDeckSlide, error)
}

type SpeakerDeckSlide struct {
	Title, Description, DownloadURL, FileName string
}

type SpeakerDeckServiceImpl struct {
	httpClient *http.Client
}

func NewSpeakerDeckService(client *http.Client) SpeakerDeckService {
	return &SpeakerDeckServiceImpl{httpClient: client}
}

func (s *SpeakerDeckServiceImpl) Fetch(ctx context.Context, url string) (*SpeakerDeckSlide, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	downloadURL := ""
	doc.Find("div.sd-main div.deck div.deck-meta a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.AttrOr("title", ""), "Download PDF") {
			downloadURL = s.AttrOr("href", "")
		}
	})

	// compare with zero value.
	if downloadURL == "" {
		return nil, errors.New("指定されたURLからはスライドが見つかりませんでした。")
	}

	slide := &SpeakerDeckSlide{
		Title:       doc.Find("h1.mb-4").Text(),
		DownloadURL: downloadURL,
		FileName:    tailURL(downloadURL),
	}

	return slide, nil
}
