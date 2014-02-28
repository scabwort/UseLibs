package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	fmt.Println("start")
	r, err := zip.OpenReader("背包.fla")
	if err != nil {
		fmt.Println("error:", err)
	}
	defer r.Close()

	for _, f := range r.File {

		f.Name = strings.Replace(f.Name, "\\", "/", -1)
		//fmt.Printf("file:%s\n", f.Name)

		rc, err := f.Open()
		if err != nil {
			fmt.Println("error2:", err)
		}

		if f.Name == "DOMDocument.xml" {
			bytes, err := ioutil.ReadAll(rc)
			if err != nil {
				fmt.Println("error3:", err)
			}
			v := FlaDocument{}
			err = xml.Unmarshal(bytes, &v)
			if err != nil {
				fmt.Println("error4:", err)
			}
			fmt.Printf("DOMFolderItem[0].Name:%v\n", v.Timelines)
			fmt.Printf("xml len:%d\n", len(bytes))
		}
		rc.Close()
	}

	fmt.Println("success!")
}
