package slidown

// SlideShareInfo ...
type SlideShareInfo struct {
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

// Link ...
type Link struct {
	Full, Normal string
}

// SpeakerDeckInfo ...
type SpeakerDeckInfo struct {
	Title, Description, DownloadURL, FileName string
}

// ReqData ...
type ReqData struct {
	URL string `json:"url"`
}

// ResMessage ...
type ResMessage struct {
	Message string `json:"message"`
}
