package storage

type FileMeta struct {
	Filename string `json:"filename"`
	Ext      string `json:"ext"`
	Size     int64  `json:"size"`
	Etag     string `json:"etag"`
}

type ImageMeta struct {
	Width       int `json:"width"`
	Height      int `json:"height"`
	Orientation int `json:"orientation"`
	Megapixels  int `json:"megapixels"`
}

type VideoMeta struct {
	Duration int     `json:"duration"`
	Codec    string  `json:"codec"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Fps      float32 `json:"fps"`
	Bitrate  int     `json:"bitrate"`
}

type AudioMeta struct {
	Duration   int    `json:"duration"`
	Codec      string `json:"codec"`
	SampleRate int    `json:"sampleRate"`
	Channels   int    `json:"channels"`
}
