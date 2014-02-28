package main

import (
	//"bytes"
	//"compress/zlib"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
)

func main() {
	//fmt.Println("start")
	var inFile string
	var outFile string
	var quality int
	flag.IntVar(&quality, "quality", 90, "out image quality")
	flag.StringVar(&inFile, "in", "0.png", "inFile image path")
	flag.StringVar(&outFile, "out", "tmp", "outFile image path")
	flag.Parse()

	file, err := os.Open(inFile)
	if err != nil {
		fmt.Printf("file not exist, file:%s\n", inFile)
		return
	}
	defer file.Close()
	pngfile, err := png.Decode(file)
	if err != nil {
		fmt.Printf("png decode fail, file:%s\n", inFile)
		return
	}

	rect := pngfile.Bounds()
	cutPos := image.Pt(0, 0)
	cutRect := image.Rect(0, 0, 0, 0)

	cutPos.X = getEge(pngfile, rect.Max.X, rect.Max.Y, true, true)
	cutPos.Y = getEge(pngfile, rect.Max.Y, rect.Max.X, true, false)
	cutRect.Max.X = getEge(pngfile, rect.Max.X, rect.Max.Y, false, true)
	cutRect.Max.Y = getEge(pngfile, rect.Max.Y, rect.Max.X, false, false)

	cutRect.Max.X -= cutPos.X
	cutRect.Max.Y -= cutPos.Y

	if cutRect.Max.X == 0 || cutRect.Max.Y == 0 {
		fmt.Printf("png is emplty, file:%s\n", inFile)
		return
	}

	//fmt.Printf("%v, %v\n", cutPos, cutRect)

	cutPic := image.NewRGBA(cutRect)
	draw.Draw(cutPic, cutRect, pngfile, cutPos, draw.Src)

	out, _ := os.Create(outFile + ".jpg")
	defer out.Close()
	jpeg.Encode(out, cutPic, &jpeg.Options{quality})

	alphaOut, _ := os.Create(outFile + ".alpha")
	defer alphaOut.Close()
	alphaOut.Write(getAlpha(pngfile, cutPos.X, cutPos.Y, cutRect.Max.X, cutRect.Max.Y))

	posOut, _ := os.Create(outFile + ".pos")
	defer posOut.Close()
	posOut.WriteString(fmt.Sprintf("%d,%d,%d,%d", cutPos.X-(rect.Max.X>>1), cutPos.Y-(rect.Max.Y>>1), cutRect.Max.X, cutRect.Max.Y))
}

func getAlpha(pic image.Image, x, y, w, h int) []byte {
	var byteData []byte
	for i := y; i < y+h; i++ {
		for j := x; j < x+w; j++ {
			c := pic.At(j, i)
			_, _, _, a := c.RGBA()
			//fmt.Println("alpha:", a>>8)
			byteData = append(byteData, uint8(a>>8))
		}
	}
	return byteData

	//var buf bytes.Buffer

	//writer, err := zlib.NewWriterLevelDict(&buf, zlib.BestCompression, byteData)
	//if err != nil {
	//	fmt.Println("压缩失败")
	//	return nil
	//}
	//writer.Write(byteData)
	//writer.Close()

	//writer := zlib.NewWriter(&buf)
	//writer.Write(byteData)
	//writer.Close()

	//return buf.Bytes()
}

func getEge(pic image.Image, w, h int, fromWay bool, xyWay bool) int {
	px := 0
	py := 0

	if fromWay {
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				if xyWay {
					px, py = x, y
				} else {
					px, py = y, x
				}
				c := pic.At(px, py)
				_, _, _, a := c.RGBA()
				if a != 0 {
					return x
				}
			}
		}
	} else {
		for x := w; x >= 0; x-- {
			for y := h; y >= 0; y-- {
				if xyWay {
					px, py = x, y
				} else {
					px, py = y, x
				}
				c := pic.At(px, py)
				_, _, _, a := c.RGBA()
				if a != 0 {
					return x
				}
			}
		}
	}
	return -1
}
