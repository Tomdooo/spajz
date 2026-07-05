package storage

/*

{
  "id": "e4d909c290d0fb1ca068ffaddf22cbd0",
  "bucket": "kamosuv-web",
  "object_key": "products/boty-01.jpg",
  "file_name": "boty-01.jpg",
  "file_size": 245120,
  "content_type": "image/jpeg",
  "etag": "\"1349b1a28a301a2e30bf33e143b83ef1\"",
  "width": 2400,
  "height": 1800,
  "created_at": "2026-07-05T10:28:00Z",
  "updated_at": "2026-07-05T10:28:00Z",
  "custom_metadata": {
    "alt": "Červené sportovní boty",
    "uploaded_by": "admin-cms"
  }
}

*/

type FileMeta struct {
	Id          string `json:"id"`
	Bucket      string `json:"bucket"`
	ObjectKey   string `json:"object_key"`
	Filename    string `json:"file_name"`
	Size        int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	Etag        string `json:"etag"`
	Ext         string `json:"ext"`

	CustomMetadata map[string]string `json:"custom_metadata"`

	Image *ImageMeta `json:"image,omitempty"`
	Video *VideoMeta `json:"video,omitempty"`
	Audio *AudioMeta `json:"audio,omitempty"`
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
