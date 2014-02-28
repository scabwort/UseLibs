package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	. "github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

type ResFile struct {
	Name string
	Url  string
	Path string
}

type VersionItem struct {
	XMLName xml.Name `xml:"VersionItem"`
	Md5     string   `xml:"md5,attr"`
	Path    string   `xml:"path,attr"`
}

type VersionList struct {
	XMLName xml.Name      `xml:"VersionList"`
	BaseUrl string        `xml:"baseUrl,attr"`
	List    []VersionItem `xml:"VersionItem"`
}

type VerFile struct {
	XMLName     *xml.Name   `xml:"VersionTable"`
	VList       VersionList `xml:"VersionList"`
	Description string      `xml:",innerxml"`
}

/**
* parallelism to load res files
 */
func getHttpFile(fileRes ResFile, execChans chan bool, doneChans chan string) {
	execChans <- true
	res, err := http.Get(fileRes.Url)
	if err != nil {
		doneChans <- "load fail:" + fileRes.Url
	} else {
		os.MkdirAll(fileRes.Path, 0755)
		ioFile, err := os.Create(fileRes.Path + "/" + fileRes.Name)
		if err != nil {
			doneChans <- "create file fail:" + fileRes.Path + "/" + fileRes.Name
			return
		}
		defer ioFile.Close()

		io.Copy(ioFile, res.Body)
		doneChans <- "ok"
	}
}

/**
* from chrom cache html file get file list
 */
func getFileListFromChromeCache(fileFrom, filterHead string) ([]ResFile, error) {
	var doc *Document
	var e error

	file, err := os.Open(fileFrom)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	if doc, e = NewDocumentFromReader(file); e != nil {
		panic(e.Error())
	}

	filterHeadLen := len(filterHead)
	fileList := []ResFile{}
	doc.Find("table tbody tr").Each(func(i int, s *Selection) {
		fileUrl := s.Text()
		if filterHeadLen > 0 && (len(fileUrl) < filterHeadLen || (fileUrl[0:filterHeadLen] != filterHead)) {
			return
		}
		tmpFile := ResFile{}
		tmpFile.Url = fileUrl
		purl, err := url.Parse(tmpFile.Url)
		if err != nil {
			fmt.Printf("parse url:%s error!", tmpFile.Url)
			return
		}

		idx := strings.LastIndex(purl.Path, "/")
		tmpFile.Name = purl.Path[idx+1 : len(purl.Path)]
		idxName := strings.LastIndex(tmpFile.Name, "_")
		if idxName > 0 && len(tmpFile.Name) > 36 {
			idxPot := strings.LastIndex(tmpFile.Name, ".")
			if idxPot > idxName {
				tmpFile.Name = tmpFile.Name[0:idxName] + tmpFile.Name[idxPot:len(tmpFile.Name)]
			}
		}

		tmpFile.Path = purl.Host + purl.Path[0:idx]
		fileList = append(fileList, tmpFile)
	})
	return fileList, nil
}

/**
* from version xml file get file list
 */
func getFromXmlVersion(fileFrom string, filterHead string) ([]ResFile, error) {
	vfile, err := ioutil.ReadFile(fileFrom)
	if err != nil {
		return nil, errors.New("version xml file not exist!")
	}
	xmlFile := VerFile{}
	err = xml.Unmarshal(vfile, &xmlFile)
	if err != nil {
		return nil, errors.New("version xml file parse error!")
	}

	filterHeadLen := len(filterHead)
	fileList := []ResFile{}
	topUrl := xmlFile.VList.BaseUrl
	backUrl := topUrl[0:strings.LastIndex(topUrl, "/")]
	for _, vUrl := range xmlFile.VList.List {
		fileUrl := ""
		if vUrl.Path[0:3] == "../" {
			fileUrl = backUrl + vUrl.Path[2:len(vUrl.Path)]
		} else {
			fileUrl = topUrl + "/" + vUrl.Path
		}
		if filterHeadLen > 0 && (len(fileUrl) < filterHeadLen || (fileUrl[0:filterHeadLen] != filterHead)) {
			continue
		}

		tmpFile := ResFile{}
		extIdx := strings.LastIndex(fileUrl, ".")
		tmpFile.Url = fileUrl[0:extIdx] + "_" + vUrl.Md5 + fileUrl[extIdx:len(fileUrl)] + "?ver=managed"
		purl, err := url.Parse(tmpFile.Url)
		if err != nil {
			fmt.Printf("parse url:%s error!", tmpFile.Url)
			return nil, errors.New("parse error")
		}

		idx := strings.LastIndex(purl.Path, "/")
		//tmpFile.Name = purl.Path[idx+1 : len(purl.Path)]
		tmpFile.Name = fileUrl[strings.LastIndex(fileUrl, "/")+1 : len(fileUrl)]
		tmpFile.Path = purl.Host + purl.Path[0:idx]
		fileList = append(fileList, tmpFile)
	}
	return fileList, nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("start")

	fileFrom := "纵横九州4.htm"
	//fileFrom := "version.xml"
	filterHead := "http://res.9z.qq.com"
	//filterHead := "http://res.9z.qq.com/flash"

	fileList, err := getFileListFromChromeCache(fileFrom, filterHead)
	//fileList, err := getFromXmlVersion(fileFrom, filterHead)

	if err != nil {
		fmt.Printf("read file list error:%v\n", err)
		return
	}

	fmt.Printf("find %d resource\n", len(fileList))

	doneChans := make(chan string, 1)
	execChans := make(chan bool, 10)
	for _, fileNode := range fileList {
		go getHttpFile(fileNode, execChans, doneChans)
	}

	total := len(fileList)
	for i := 0; i < total; i++ {
		r := <-doneChans
		<-execChans
		//fmt.Println(r)
		if r == "ok" {
			fmt.Printf("load over %d/%d ok\n", i+1, total)
		} else {
			fmt.Printf("load %d/%d, %s\n", i+1, total, r)
		}
	}
	close(doneChans)

	fmt.Printf("ok\n")
}
