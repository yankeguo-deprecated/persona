package main

import (
	"bytes"
	"fmt"
	"github.com/skip2/go-qrcode"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/rasterizer"
	"gopkg.in/yaml.v2"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Item struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Tokens  []Item `yaml:"tokens"`
}

var fontFamily *canvas.FontFamily

func main() {
	var err error
	defer func(err *error) {
		if *err != nil {
			panic(*err)
		}
	}(&err)

	fontFamily = canvas.NewFontFamily("Custom")
	fontFamily.Use(canvas.CommonLigatures)
	if err = fontFamily.LoadFontFile(filepath.Join("src", "custom-font.ttf"), canvas.FontRegular); err != nil {
		return
	}

	var buf []byte
	if buf, err = ioutil.ReadFile(filepath.Join("src", "addresses.yml")); err != nil {
		return
	}

	dec := yaml.NewDecoder(bytes.NewReader(buf))

	md := &bytes.Buffer{}
	md.WriteString("# Persona\n\n")
	md.WriteString("Personal Crypto-Currency Addresses\n")

	for {
		var item Item
		if err = dec.Decode(&item); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if err = generate(item.ID, item.Name, item.Address); err != nil {
			return
		}
		for _, token := range item.Tokens {
			if err = generate(token.ID, token.Name, item.Address); err != nil {
				return
			}
			md.WriteString(fmt.Sprintf("\n![%s](dist/%s.png)\n", token.Name, token.ID))
		}
		md.WriteString(fmt.Sprintf("\n![%s](dist/%s.png)\n", item.Name, item.ID))
	}

	if err = ioutil.WriteFile("README.md", md.Bytes(), 0640); err != nil {
		return
	}
}

func generate(id, name, address string) (err error) {
	gray := color.RGBA{R: 51, G: 51, B: 51, A: 255}
	log.Println(id, name, address)

	var q *qrcode.QRCode
	if q, err = qrcode.New(address, qrcode.High); err != nil {
		return
	}
	q.ForegroundColor = gray
	b := &bytes.Buffer{}
	if err = q.Write(512, b); err != nil {
		return
	}

	buf := b.Bytes()

	var img image.Image
	if img, err = png.Decode(bytes.NewReader(buf)); err != nil {
		return
	}

	if buf, err = ioutil.ReadFile(filepath.Join("src", "logos", id+"-logo.png")); err != nil {
		return
	}
	var logo image.Image
	if logo, err = png.Decode(bytes.NewReader(buf)); err != nil {
		return
	}

	logoW, _ := float64(logo.Bounds().Max.X), float64(logo.Bounds().Max.Y)

	logoSize := float64(64)

	c := canvas.New(600, 800)
	ctx := canvas.NewContext(c)
	ctx.SetFillColor(color.White)
	bgLine := &canvas.Polyline{}
	bgLine.Add(600, 0).Add(600, 800).Add(0, 800).Add(0, 0)
	ctx.DrawPath(0, 0, bgLine.ToPath())
	ctx.DrawImage((600.0-512.0)/2.0, (800.0-512.0)-((600.0-512.0)/2.0), img, 1)
	bgCircle := canvas.Circle(50)
	ctx.DrawPath(297, 500, bgCircle)
	ctx.DrawImage(265, 469, logo, logoW/logoSize)
	ctx.SetFillColor(gray)
	borderLine := &canvas.Polyline{}
	borderLine.Add(0, 0).Add(560, 0).Add(560, 760).Add(0, 760).Add(0, 0)
	ctx.DrawPath(20, 20, borderLine.ToPath().Stroke(4.0, canvas.RoundCap, canvas.ArcsJoin))

	headerFace := fontFamily.Face(128.0, gray, canvas.FontRegular, canvas.FontNormal)
	tb := canvas.NewTextBox(headerFace, name, 600, 200, canvas.Center, canvas.Center, 0.0, 0.0)
	ctx.DrawText(0, 250, tb)

	if err = os.MkdirAll(filepath.Join("dist"), 0755); err != nil {
		return
	}
	if err = c.WriteFile(filepath.Join("dist", id+".png"), rasterizer.PNGWriter(1)); err != nil {
		return
	}
	return
}
