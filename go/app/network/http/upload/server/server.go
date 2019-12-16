package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {

	http.HandleFunc("/upload", upload)
	err := http.ListenAndServe("127.0.0.1:8088", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.Method) //GET

	//这个很重要,必须写
	//r.ParseForm()
	_ = r.ParseMultipartForm(32 << 20)
	user := r.Form.Get("user")
	password := r.Form.Get("password")

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err, "--------1------------") //上传错误
	}
	defer file.Close()

	tmpFile, err := os.Create(handler.Filename)
	if err != nil {
		fmt.Println("error opening file")
		os.Exit(1)
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		fmt.Println("error copying file")
		os.Exit(1)
	}

	buffer := bytes.Buffer{}
	_, _ = buffer.ReadFrom(tmpFile)

	fmt.Println(user, password, handler.Filename, buffer.String())

	/*deviceType := r.Form.Get("deviceType")
	filename := r.Form.Get("filename")

	fmt.Println(deviceType, filename) // 1 test.zip

	deviceType3 := r.PostForm.Get("deviceType")
	fmt.Println(deviceType3)

	//第二种方式,底层是r.Form
	deviceType2 := r.FormValue("deviceType")
	filename2 := r.FormValue("filename")
	fmt.Println(deviceType2, filename2) // 1 test.zip*/

}
