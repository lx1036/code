package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

/**
https://my.oschina.net/solate/blog/741039
 */


func postFile(url, filename, path, deviceType, deviceId string, filePath string) error {
	//打开文件句柄操作
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer file.Close()
	
	//创建一个模拟的form中的一个选项,这个form项现在是空的
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	
	//关键的一步操作, 设置文件的上传参数叫uploadfile, 文件名是filename,
	//相当于现在还没选择文件, form项里选择文件的选项
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}
	
	//iocopy 这里相当于选择了文件,将文件放到form中
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err
	}
	
	//获取上传文件的类型,multipart/form-data; boundary=...
	contentType := bodyWriter.FormDataContentType()
	
	//这个很关键,必须这样写关闭,不能使用defer关闭,不然会导致错误
	bodyWriter.Close()
	
	
	//这里就是上传的其他参数设置,可以使用 bodyWriter.WriteField(key, val) 方法
	//也可以自己在重新使用  multipart.NewWriter 重新建立一项,这个再server 会有例子
	params := map[string]string{
		"filename" : filename,
		"path" : path,
		"deviceType" : deviceType,
		"deviceId" : deviceId,
		
	}
	//这种设置值得仿佛 和下面再从新创建一个的一样
	for key, val := range params {
		_ = bodyWriter.WriteField(key, val)
	}
	
	//发送post请求到服务端
	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}


func httpGet() {
	//发送get 请求
	resp, err := http.Get("http://127.0.0.1:9090/upload?id=1&filename=test.zip")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
}




func postFile2() error {
	
	//打开文件句柄操作
	file, err := os.Open("air.log")
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer file.Close()
	
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", "json.log")
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}
	
	//iocopy
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err
	}
	
	////设置其他参数
	/*params := map[string]string{
		"user": "test",
		"password": "123456",
	}*/
	//
	////这种设置值得仿佛 和下面再从新创建一个的一样
	/*for key, val := range params {
		_ = bodyWriter.WriteField(key, val)
	}*/
	
	//和上面那种效果一样
	//建立第二个fields
	if fileWriter, err = bodyWriter.CreateFormField("user"); err != nil {
		fmt.Println(err, "----------4--------------")
	}
	if _, err = fileWriter.Write([]byte("test")); err != nil {
		fmt.Println(err, "----------5--------------")
	}
	//建立第三个fieds
	if fileWriter, err = bodyWriter.CreateFormField("password"); err != nil {
		fmt.Println(err, "----------4--------------")
	}
	if _, err = fileWriter.Write([]byte("123456")); err != nil {
		fmt.Println(err, "----------5--------------")
	}
	
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	
	resp, err := http.Post("http://127.0.0.1:8088/upload", contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}
// sample usage
func main() {
	
	postFile2()
	
	/*url := "http://localhost:8088/upload"
	filename := "file1"
	path := "/eagleeye"
	deviceType := "iphone"
	deviceId := "e6c5a83c5e20420286bb00b90b938d92"
	
	filePath := "./air.log" //上传的文件
	
	
	postFile(url, filename, path, deviceType, deviceId,  filePath)*/
}
