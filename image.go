package main

import (
	"bytes"
	"path"
	"strings"

	"os"

	"math"
	"math/rand"
	"time"

	"image"
	"image/color"
	"image/draw"
	"image/png"

	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/vgimg"
	// "gonum.org/v1/plot/plotter"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// WaterMark for adding a watermark on the image
func WaterMark(img image.Image, address, qrcodeType string) (image.Image, error) {
	// image's length to canvas's length
	bounds := img.Bounds()
	w := vg.Length(bounds.Max.X) * vg.Inch / vgimg.DefaultDPI
	h := vg.Length(bounds.Max.Y) * vg.Inch / vgimg.DefaultDPI
	diagonal := vg.Length(math.Sqrt(float64(w*w + h*h)))
	// create a canvas, which width and height are diagonal
	c := vgimg.New(diagonal, diagonal/2)

	// make a fontStyle, which width is vg.Inch * 0.7
	fontStyle, _ := vg.MakeFont("Courier", diagonal/42)

	// set the color of markText
	c.SetColor(color.RGBA{0, 0, 0, 200})
	c.FillString(fontStyle, vg.Point{X: vg.Length(bounds.Min.X + 10), Y: diagonal/2 - 20}, "Ethereum Address: ")
	c.FillString(fontStyle, vg.Point{X: vg.Length(bounds.Min.X + 10), Y: diagonal/2 - 35}, address)

	c.FillString(fontStyle, vg.Point{X: vg.Length(bounds.Min.X + 10), Y: diagonal/2 - 60}, strings.Join([]string{"Generate Time:", time.Now().Format("2006-01-02 15:04:05")}, " "))
	c.FillString(fontStyle, vg.Point{X: vg.Length(bounds.Min.X + 10), Y: diagonal/2 - 80}, strings.Join([]string{"qrcode type:", qrcodeType}, " "))

	// canvas writeto jpeg
	// canvas.img is private
	// so use a buffer to transfer
	jc := vgimg.PngCanvas{Canvas: c}
	buff := new(bytes.Buffer)
	jc.WriteTo(buff)
	img, _, err := image.Decode(buff)
	if err != nil {
		return nil, err
	}

	// get the center point of the image
	ctp := int(diagonal * vgimg.DefaultDPI / vg.Inch / 2)

	// cutout the marked image
	size := bounds.Size()
	bounds = image.Rect(ctp-size.X/2, ctp-size.Y/2, ctp+size.X/2, ctp+size.Y/2)
	rv := image.NewRGBA(bounds)
	draw.Draw(rv, bounds, img, image.Point{0, 0}, draw.Src)
	return rv, nil
}

// MarkingPicture for marking picture with text
func MarkingPicture(filepath, address, qrcodeType string) (image.Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	img, err = WaterMark(img, address, qrcodeType)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func wm(target, address, qrcodeType string) {
	img, err := MarkingPicture(target, address, qrcodeType)
	if err != nil {
		log.Fatalln(err.Error())
	}

	srcFile, err := os.Open(target)
	if err != nil {
		log.Fatalln(err.Error())
	}

	srcImage, _, err := image.Decode(srcFile)
	if err != nil {
		log.Fatalln(err.Error())
	}

	sb := srcImage.Bounds()
	r2 := image.Rectangle{}
	r2.Min.X = 0
	r2.Min.Y = sb.Max.Y
	r2.Max.X = sb.Max.X
	r2.Max.Y = sb.Max.Y * 2
	r := image.Rectangle{image.Point{0, 0}, r2.Max}
	rgba := image.NewRGBA(r)

	draw.Draw(rgba, sb, srcImage, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, r2, img, image.Point{img.Bounds().Min.X, img.Bounds().Min.Y - 10}, draw.Src)

	ext := path.Ext(target)
	base := strings.Split(path.Base(target), ".")[0] + "_marked"
	wmFileName := strings.Join([]string{base, ext}, "")
	f, err := os.Create(strings.Join([]string{path.Dir(target), wmFileName}, "/"))
	if err != nil {
		log.Fatalln(err.Error())
	}
	png.Encode(f, rgba)
}
