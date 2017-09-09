package speakerdeck

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/monmaru/slidown/common"
	"github.com/pkg/errors"
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

	details := doc.Find("#talk-details")
	downloadURL := doc.Find("#share_pdf").AttrOr("href", "")

	// compare with zero value.
	if downloadURL == "" {
		return nil, errors.New("指定されたURLからはスライドが見つかりませんでした。")
	}

	slide := &Slide{
		Title:       details.Find("h1").Text(),
		Description: strings.TrimSpace(details.Find(".description").Text()),
		DownloadURL: downloadURL,
		FileName:    common.TailURL(downloadURL),
	}

	common.SetCache(ctx, key, slide)
	return slide, nil
}
