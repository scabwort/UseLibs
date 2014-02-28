package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
)

//-ldflags="-H windowsgui -s -w"
func main() {
	dir := flag.String("dir", "./", "服务器地址")
	port := flag.Int("port", 80, "服务器端口")
	flag.Parse()

	http.HandleFunc("/", http.FileServer(http.Dir(*dir)).ServeHTTP)
	fmt.Print("bb")
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)

	fmt.Print("aa")
}

// 静态文件处理
//func StaticServer(w http.ResponseWriter, req *http.Request) {
//staticHandler.ServeHTTP(w, req)
//if req.URL.Path != "/" {
//	staticHandler.ServeHTTP(w, req)
//	return
//}

//io.WriteString(w, "hello, world!\n")
//}
