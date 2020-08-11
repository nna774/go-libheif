package main

import (
	"image/jpeg"
	"log"
	"os"

	"github.com/nna774/go-libheif/heif"
)

func main() {
	file, err := os.Open("example.heic")
	if err != nil {
		log.Fatal(err)
	}

	img, err := heif.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("image.jpeg")
	if err != nil {
		log.Fatal(err)
	}

	opt := jpeg.Options{Quality: 100}
	if err := jpeg.Encode(f, img, &opt); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
