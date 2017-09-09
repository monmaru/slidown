package slideshare

type Slide struct {
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

type link struct {
	Full, Normal string
}
