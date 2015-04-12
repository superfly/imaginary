package main

import (
	"encoding/json"
	"gopkg.in/h2non/bimg.v0"
)

type ImageOptions struct {
	Width       int
	Height      int
	AreaWidth   int
	AreaHeight  int
	Quality     int
	Compression int
	Rotate      int
	Top         int
	Left        int
	Margin      int
	Factor      int
	DPI         int
	TextWidth   int
	NoReplicate bool
	Opacity     float64
	Text        string
	Font        string
	Type        string
	Color       []uint8
	Mime        string
}

type Image struct {
	Body []byte
	Mime string
}

type Operation func([]byte, ImageOptions) (Image, error)

func (o Operation) Run(buf []byte, opts ImageOptions) (Image, error) {
	return o(buf, opts)
}

func BimgOptions(o ImageOptions) bimg.Options {
	return bimg.Options{
		Width:       o.Width,
		Height:      o.Height,
		Quality:     o.Quality,
		Compression: o.Quality,
		Type:        ImageType(o.Type),
	}
}

type ImageInfo struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Type        string `json:"type"`
	Space       string `json:"space"`
	Alpha       bool   `json:"alpha"`
	Profile     bool   `json:"profile"`
	Channels    int    `json:"channels"`
	Orientation int    `json:"orientation"`
}

func Info(buf []byte, o ImageOptions) (Image, error) {
	image := Image{}

	meta, err := bimg.Metadata(buf)
	if err != nil {
		return image, NewError("Cannot retrieve image medatata: %s"+err.Error(), BAD_REQUEST)
	}

	info := ImageInfo{
		Width:       meta.Size.Width,
		Height:      meta.Size.Height,
		Type:        meta.Type,
		Space:       meta.Space,
		Alpha:       meta.Alpha,
		Profile:     meta.Profile,
		Channels:    meta.Channels,
		Orientation: meta.Orientation,
	}

	body, _ := json.Marshal(info)
	image.Body = body
	image.Mime = "application/json"

	return image, nil
}

func Resize(buf []byte, o ImageOptions) (Image, error) {
	if o.Width == 0 || o.Height == 0 {
		return Image{}, NewError("Missing required params: height, width", BAD_REQUEST)
	}

	opts := BimgOptions(o)
	return Process(buf, opts)
}

func Enlarge(buf []byte, o ImageOptions) (Image, error) {
	if o.Width == 0 || o.Height == 0 {
		return Image{}, NewError("Missing required params: height, width", BAD_REQUEST)
	}

	opts := BimgOptions(o)
	opts.Enlarge = true
	return Process(buf, opts)
}

func Crop(buf []byte, o ImageOptions) (Image, error) {
	opts := BimgOptions(o)
	opts.Crop = true
	return Process(buf, opts)
}

func Rotate(buf []byte, o ImageOptions) (Image, error) {
	if o.Rotate == 0 {
		return Image{}, NewError("Missing rotate param", BAD_REQUEST)
	}

	opts := BimgOptions(o)
	opts.Rotate = bimg.Angle(o.Rotate)
	return Process(buf, opts)
}

func Flip(buf []byte, o ImageOptions) (Image, error) {
	opts := BimgOptions(o)
	opts.Flip = true
	return Process(buf, opts)
}

func Flop(buf []byte, o ImageOptions) (Image, error) {
	opts := BimgOptions(o)
	opts.Flop = true
	return Process(buf, opts)
}

func Thumbnail(buf []byte, o ImageOptions) (Image, error) {
	if o.Width == 0 && o.Height == 0 {
		return Image{}, NewError("Missing required params: width or height", BAD_REQUEST)
	}

	opts := BimgOptions(o)
	return Process(buf, opts)
}

func Zoom(buf []byte, o ImageOptions) (Image, error) {
	if o.Factor == 0 {
		return Image{}, NewError("Missing required param: factor", BAD_REQUEST)
	}

	opts := BimgOptions(o)

	if o.Top > 0 || o.Left > 0 {
		if o.AreaWidth == 0 && o.AreaHeight == 0 {
			return Image{}, NewError("Missing required extract area params: areawidth, areaheight", BAD_REQUEST)
		}
		opts.Top = o.Top
		opts.Left = o.Left
		opts.AreaWidth = o.AreaWidth
		opts.AreaHeight = o.AreaHeight
	}

	opts.Zoom = o.Factor
	return Process(buf, opts)
}

func Convert(buf []byte, o ImageOptions) (Image, error) {
	if o.Type == "" {
		return Image{}, NewError("Missing required params: type", BAD_REQUEST)
	}

	opts := BimgOptions(o)
	return Process(buf, opts)
}

func Watermark(buf []byte, o ImageOptions) (Image, error) {
	if o.Text == "" {
		return Image{}, NewError("Missing required params: text", BAD_REQUEST)
	}

	if o.TextWidth == 0 {
		o.TextWidth = 100
	}

	opts := BimgOptions(o)
	opts.Watermark.DPI = o.DPI
	opts.Watermark.Text = o.Text
	opts.Watermark.Font = o.Font
	opts.Watermark.Margin = o.Margin
	opts.Watermark.Width = o.TextWidth
	opts.Watermark.NoReplicate = o.NoReplicate
	opts.Watermark.Opacity = float32(o.Opacity)

	if len(o.Color) > 2 {
		opts.Watermark.Background = bimg.Color{o.Color[0], o.Color[1], o.Color[2]}
	}

	return Process(buf, opts)
}

func Extract(buf []byte, o ImageOptions) (Image, error) {
	if o.Top == 0 || o.Left == 0 {
		return Image{}, NewError("Missing required params: top, left", BAD_REQUEST)
	}

	opts := BimgOptions(o)
	return Process(buf, opts)
}

func Process(buf []byte, opts bimg.Options) (Image, error) {
	buf, err := bimg.Resize(buf, opts)
	if err != nil {
		return Image{}, err
	}

	mime := GetImageMimeType(bimg.DetermineImageType(buf))
	return Image{Body: buf, Mime: mime}, nil
}