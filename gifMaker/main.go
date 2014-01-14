package main

import (
	"flag"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("start")

	var folder string
	var frameNum int
	flag.StringVar(&folder, "in", "./", "source folder")
	flag.IntVar(&frameNum, "frame", 12, "frame rate")
	flag.Parse()

	delay := int(100 / frameNum)
	finfo, err := os.Stat(folder)
	if err != nil {
		fmt.Printf("folder:%s is not exist!\n", folder)
		return
	} else if !finfo.IsDir() {
		fmt.Printf("folder:%s is not folder!\n", folder)
		return
	}

	fileName := folder
	if fileName == "./" {
		fileName, _ = os.Getwd()
		fmt.Printf("path3:%s\n", fileName)
	}
	_, fileName = filepath.Split(fileName)
	fmt.Printf("path:%s\n", fileName)

	gifInfo := gif.GIF{}
	files, err := ioutil.ReadDir(folder)

	if err != nil {
		fmt.Printf("read folder:%s error!\n", folder)
		return
	}

	for _, fileInfo := range files {
		ext := strings.ToLower(path.Ext(fileInfo.Name()))

		fmt.Printf("file:%s, ext:%s\n", fileInfo.Name(), ext)
		if ext != ".png" && ext != ".jpg" {
			continue
		}

		fullPath := folder + "/" + fileInfo.Name()
		file, _ := os.Open(fullPath)
		defer file.Close()

		if ext == ".png" {
			pic, err := png.Decode(file)
			if err != nil {
				fmt.Printf("png file:%s format is error!\n", fullPath)
			}
			gifInfo.Image = append(gifInfo.Image, ImageToPaletted(pic))
		} else {
			pic, err := jpeg.Decode(file)
			if err != nil {
				fmt.Printf("jpeg file:%s format is error!\n", fullPath)
			}
			gifInfo.Image = append(gifInfo.Image, ImageToPaletted(pic))
		}

		gifInfo.Delay = append(gifInfo.Delay, delay)
	}

	if len(gifInfo.Image) == 0 {
		fmt.Printf("folder:%s has not image file!\n", folder)
		return
	}

	outFile, _ := os.Create(fileName + ".gif")
	defer outFile.Close()

	fmt.Printf("create gif:%s, and delay:%d success!\n", fileName, delay)
	gif.EncodeAll(outFile, &gifInfo)
}

func ImageToPaletted(img image.Image) *image.Paletted {
	b := img.Bounds()
	pm := image.NewPaletted(b, palette.Plan9)
	draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
	return pm
}
