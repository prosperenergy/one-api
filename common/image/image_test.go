package image_test

// These tests used to download sample images from Wikimedia Commons. That
// request is not an image anymore: Go's default client sends User-Agent
// "Go-http-client/1.1", which Commons rejects with 403, and the 2560px
// thumbnail URL also returns 400 ("Use thumbnail sizes listed on
// https://w.wiki/GHai"). image.Decode then reports "image: unknown format".
// Fixtures are generated locally so the unit tests do not depend on Commons.

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/songquanpeng/one-api/common/client"
	img "github.com/songquanpeng/one-api/common/image"
	"github.com/stretchr/testify/require"

	_ "golang.org/x/image/webp"
)

// oneByOneWebP is a 1x1 lossless WebP. golang.org/x/image/webp decodes WebP
// but does not encode it, so the fixture is stored directly.
var oneByOneWebP = []byte{
	0x52, 0x49, 0x46, 0x46, 0x1a, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50,
	0x56, 0x50, 0x38, 0x4c, 0x0d, 0x00, 0x00, 0x00, 0x2f, 0x00, 0x00, 0x00,
	0x10, 0x07, 0x10, 0x11, 0x11, 0x88, 0x88, 0xfe, 0x07, 0x00,
}

type imageCase struct {
	format string
	mime   string
	width  int
	height int
	data   []byte
}

func TestMain(m *testing.M) {
	client.Init()
	m.Run()
}

func solid(w, h int) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	c := color.RGBA{R: 20, G: 80, B: 160, A: 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			im.Set(x, y, c)
		}
	}
	return im
}

func mustEncode(t *testing.T, encode func(io.Writer, image.Image) error, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, encode(&buf, solid(w, h)))
	return buf.Bytes()
}

func cases(t *testing.T) []imageCase {
	t.Helper()
	return []imageCase{
		{
			format: "jpeg",
			mime:   "image/jpeg",
			width:  8,
			height: 4,
			data: mustEncode(t, func(w io.Writer, im image.Image) error {
				return jpeg.Encode(w, im, nil)
			}, 8, 4),
		},
		{
			format: "png",
			mime:   "image/png",
			width:  3,
			height: 5,
			data: mustEncode(t, func(w io.Writer, im image.Image) error {
				return png.Encode(w, im)
			}, 3, 5),
		},
		{
			format: "gif",
			mime:   "image/gif",
			width:  6,
			height: 2,
			data: mustEncode(t, func(w io.Writer, im image.Image) error {
				pal := []color.Color{color.RGBA{R: 20, G: 80, B: 160, A: 255}}
				src := im.Bounds()
				pm := image.NewPaletted(src, pal)
				for y := src.Min.Y; y < src.Max.Y; y++ {
					for x := src.Min.X; x < src.Max.X; x++ {
						pm.Set(x, y, pal[0])
					}
				}
				return gif.Encode(w, pm, nil)
			}, 6, 2),
		},
		{
			format: "webp",
			mime:   "image/webp",
			width:  1,
			height: 1,
			data:   oneByOneWebP,
		},
	}
}

func TestDecode(t *testing.T) {
	for _, c := range cases(t) {
		t.Run("Decode:"+c.format, func(t *testing.T) {
			decoded, format, err := image.Decode(bytes.NewReader(c.data))
			require.NoError(t, err)
			require.Equal(t, c.format, format)
			require.Equal(t, c.width, decoded.Bounds().Dx())
			require.Equal(t, c.height, decoded.Bounds().Dy())
		})
		t.Run("DecodeConfig:"+c.format, func(t *testing.T) {
			config, format, err := image.DecodeConfig(bytes.NewReader(c.data))
			require.NoError(t, err)
			require.Equal(t, c.format, format)
			require.Equal(t, c.width, config.Width)
			require.Equal(t, c.height, config.Height)
		})
	}
}

func TestBase64(t *testing.T) {
	for _, c := range cases(t) {
		encoded := base64.StdEncoding.EncodeToString(c.data)
		t.Run("Decode:"+c.format, func(t *testing.T) {
			body := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(encoded)))
			decoded, format, err := image.Decode(body)
			require.NoError(t, err)
			require.Equal(t, c.format, format)
			require.Equal(t, c.width, decoded.Bounds().Dx())
			require.Equal(t, c.height, decoded.Bounds().Dy())
		})
		t.Run("DecodeConfig:"+c.format, func(t *testing.T) {
			body := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(encoded)))
			config, format, err := image.DecodeConfig(body)
			require.NoError(t, err)
			require.Equal(t, c.format, format)
			require.Equal(t, c.width, config.Width)
			require.Equal(t, c.height, config.Height)
		})
	}
}

func TestGetImageSize(t *testing.T) {
	images := cases(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, c := range images {
			if r.URL.Path == "/"+c.format {
				w.Header().Set("Content-Type", c.mime)
				_, _ = w.Write(c.data)
				return
			}
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	for _, c := range images {
		t.Run(c.format, func(t *testing.T) {
			width, height, err := img.GetImageSize(srv.URL + "/" + c.format)
			require.NoError(t, err)
			require.Equal(t, c.width, width)
			require.Equal(t, c.height, height)
		})
	}
}

func TestGetImageSizeFromBase64(t *testing.T) {
	for _, c := range cases(t) {
		t.Run(c.format, func(t *testing.T) {
			encoded := base64.StdEncoding.EncodeToString(c.data)
			width, height, err := img.GetImageSizeFromBase64(encoded)
			require.NoError(t, err)
			require.Equal(t, c.width, width)
			require.Equal(t, c.height, height)

			dataURL := "data:" + c.mime + ";base64," + encoded
			width, height, err = img.GetImageSize(dataURL)
			require.NoError(t, err)
			require.Equal(t, c.width, width)
			require.Equal(t, c.height, height)
		})
	}
}
