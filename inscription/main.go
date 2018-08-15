package main

import (
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/reconquest/karma-go"
)

var version = "1.0"

var usage = `inscription - graphical font tool.

Usage:
  inscription --help
  inscription -G -w= -h= -d= -m= <x> <y> <file>

Options:
  --help                 Show this help.
  -G --generate          Generate font image layout.
  -w --width <width>     Glyph width.
  -h --height <height>   Glyph height.
  -d --descend <height>  Descend size.
  -m --medium <height>   Medium (x-height) size.
`

type Opts struct {
	Generate bool   `docopt:"--generate"`
	Width    int    `docopt:"--width"`
	Height   int    `docopt:"--height"`
	Descend  int    `docopt:"--descend"`
	Medium   int    `docopt:"--medium"`
	File     string `docopt:"<file>"`
	X        int    `docopt:"<x>"`
	Y        int    `docopt:"<y>"`
}

func main() {
	args, err := docopt.ParseArgs(usage, nil, version)
	if err != nil {
		panic(err)
	}

	var opts Opts

	err = args.Bind(&opts)
	if err != nil {
		panic(err)
	}

	switch {
	case opts.Generate:
		err = generate(opts)
	}

	if err != nil {
		panic(err)
	}
}

func generate(opts Opts) error {
	rgb := func(r, g, b uint8) color.RGBA {
		return color.RGBA{r, g, b, 255}
	}

	palette := color.Palette{
		rgb(0xff, 0xff, 0xff),
		rgb(0x00, 0x00, 0x00),

		rgb(0xff, 0xf2, 0xe0),
		rgb(0xff, 0xeb, 0xd1),

		rgb(0xe8, 0xee, 0xfa),
	}

	img := image.NewPaletted(
		image.Rect(0, 0, opts.X*opts.Width, opts.Y*opts.Height),
		palette,
	)

	for y := 0; y < opts.Y*opts.Height; y++ {
		for x := 0; x < opts.X*opts.Width; x++ {
			var (
				index  uint8
				medium bool
			)

			gy := opts.Height - y%opts.Height
			if gy > opts.Descend {
				if gy < opts.Descend+opts.Medium {
					medium = true
				}
			}

			if (x/opts.Width)%2 == (y/opts.Height)%2 {
				if medium {
					index = 4
				} else {
					index = 0
				}
			} else {
				if medium {
					index = 3
				} else {
					index = 2
				}
			}

			img.SetColorIndex(x, y, index)
		}
	}

	file, err := os.Create(opts.File)
	if err != nil {
		return karma.Format(
			err,
			"{generate} unable to open file: %s",
			opts.File,
		)
	}

	err = png.Encode(file, img)
	if err != nil {
		return karma.Format(
			err,
			"{generate} unable to encode PNG",
		)
	}

	return nil
}
