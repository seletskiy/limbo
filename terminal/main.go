package main

import (
	"image/png"
	"os"

	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/limbo/terminal/engine"
)

const (
	width  = 500
	height = 500
)

var (
	log lorg.Logger
)

func main() {
	log = lorg.NewLog()

	engine, err := engine.New(log)
	if err != nil {
		panic(err)
	}

	defer engine.Stop()

	_, err = engine.CreateWindow(640, 460, "test")
	if err != nil {
		panic(err)
	}

	_, err = engine.CreateWindow(640, 460, "test")
	if err != nil {
		panic(err)
	}

	file, err := os.Open("font.png")
	if err != nil {
		panic(err)
	}

	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	engine.SetFont(img)

	for {
		if engine.Empty() {
			break
		}

		err := engine.Render()
		if err != nil {
			panic(err)
		}
	}
}
