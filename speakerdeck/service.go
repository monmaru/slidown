package speakerdeck

import (
	"errors"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/monmaru/slidown/common"
	"golang.org/x/net/context"
)

type service struct {
	httpClient *http.Client
}

func newService(client *http.Client) *service {
	return &service{httpClient: client}
}

func (s *service) fetch(ctx context.Context, url string) (*Slide, error) {
	var cache Slide
	key := "SD-" + common.TailURL(url)
	if err := common.GetCache(ctx, key, &cache); err == nil {
		return &cache, nil
	}

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

	slide := &Slide{
		Title:       doc.Find("h1.mb-4").Text(),
		DownloadURL: downloadURL,
		FileName:    common.TailURL(downloadURL),
	}

	common.SetCache(ctx, key, slide)
	return slide, nil
}
