package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"math"
	"os"
	"runtime"
	"time"
)

func cutImageChan(x, y, cellWidth, cellHeight int, src image.Image, folder string, quality int, doneChans chan bool) {
	//execChans <- true
	cutImage(x, y, cellWidth, cellHeight, src, folder, quality)
	doneChans <- true
}
func cutImage(x, y, cellWidth, cellHeight int, src image.Image, folder string, quality int) {

	cellPos := image.Pt(x*cellWidth, y*cellHeight)
	cellRect := image.Rect(0, 0, cellWidth, cellHeight)
	if (x+1)*cellWidth > src.Bounds().Max.X {
		cellRect.Max.X = src.Bounds().Max.X - x*cellWidth
	}
	if (y+1)*cellHeight > src.Bounds().Max.Y {
		cellRect.Max.Y = src.Bounds().Max.Y - y*cellHeight
	}

	//fmt.Printf("(%d, %d)=%v-%v\n", x, y, cellPos, cellRect.Max)

	m := image.NewRGBA(cellRect)
	draw.Draw(m, cellRect, src, cellPos, draw.Src)

	file, err := os.Create(fmt.Sprintf("%s/%d_%d.jpg", folder, x, y))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err2 := jpeg.Encode(file, m, &jpeg.Options{quality})
	if err2 != nil {
		panic(err2)
	}
}

func main() {
	from := flag.String("from", "bg.jpg", "from info")
	out := flag.String("out", "out", "out info")
	cw := flag.Int("cw", 256, "cw info")
	ch := flag.Int("ch", 256, "ch info")
	ql := flag.Int("q", 90, "quality")
	flag.Parse()
	fmt.Printf("cell size:(%d,%d), quality:%d, from:%s, out:%s\n", *cw, *ch, *ql, *from, *out)

	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Printf("start, use cpu num:%d\n", numCPU)
	t := time.Now()
	srcJpgFile, err := os.Open(*from)
	if err != nil {
		panic(err)
	}
	defer srcJpgFile.Close()

	srcJpg, err := jpeg.Decode(srcJpgFile)
	if err != nil {
		panic(err)
	}

	folder := *out
	os.MkdirAll(folder, os.ModeDir)

	fmt.Printf("parse time:%s\n", time.Now().Sub(t).String())
	t = time.Now()

	cellWidth := *cw
	cellHeight := *ch
	quality := *ql
	stcSize := srcJpg.Bounds().Size()
	cellRect := image.Rect(0, 0, cellWidth, cellHeight)
	cellSize := cellRect.Size()
	cellNumX := int(math.Ceil(float64(stcSize.X) / float64(cellWidth)))
	cellNumY := int(math.Ceil(float64(stcSize.Y) / float64(cellHeight)))
	fmt.Printf("src size:%v, cell size:%v, cell num:(%d, %d)\n", stcSize, cellSize, cellNumX, cellNumY)

	total := cellNumX * cellNumY
	//execChans := make(chan bool, numCPU)
	doneChans := make(chan bool, 1)
	for y := 0; y < cellNumY; y++ {
		for x := 0; x < cellNumX; x++ {
			go cutImageChan(x, y, cellWidth, cellHeight, srcJpg, folder, quality, doneChans)
		}
	}
	for i := 0; i < total; i++ {
		r := <-doneChans
		if !r {
			fmt.Printf("index of %d failed!\n", i)
		}
	}
	close(doneChans)

	fmt.Printf("cut time:%s\n", time.Now().Sub(t).String())

	fmt.Printf("ok\n")
}
