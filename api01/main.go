package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/secsy/goftp"
)

func main() {

	http.HandleFunc("/api/ftp", handler)
	http.HandleFunc("/api/store", StoreFile)
	http.ListenAndServe(":50016", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	len := r.ContentLength          // 获取请求实体长度
	body := make([]byte, len)       // 创建存放请求实体的字节切片
	r.Body.Read(body)               // 调用 Read 方法读取请求实体并将返回内容存放到上面创建的字节切片
	io.WriteString(w, string(body)) // 将请求实体作为响应实体返回

	file := "file.txt"
	//if err := os.WriteFile("file1.txt", []byte("Hello GOSAMPLES!"), 0666); err != nil {
	if err := os.WriteFile(file, []byte(body), 0666); err != nil {
		log.Fatal(err)
	}

	//ftp(body)
	config := goftp.Config{
		User:               "F01",
		Password:           "A@1qaz2wsx",
		ConnectionsPerHost: 10,
		Timeout:            10 * time.Second,
		Logger:             os.Stderr,
	}
	// Create client object with default config
	client, err := goftp.DialConfig(config, "10.4.33.118")
	if err != nil {
		panic(err)
	}

	defer client.Close()

	file = "file.txt"

	// Upload a file from disk
	bigFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	err = client.Store(file, bigFile)
	if err != nil {
		panic(err)
	}

	err = os.Remove(file)
	if err != nil {

		//如果删除失败则输出 file remove Error!

		fmt.Println("file remove Error!")

		//输出错误详细信息

		fmt.Printf("%s", err)

	} else {

		//如果删除成功则输出 file remove OK!

		fmt.Println("file remove OK!")

	}
}

func ftp(data []byte) {

	config := goftp.Config{
		User:               "F01",
		Password:           "A@1qaz2wsx",
		ConnectionsPerHost: 10,
		Timeout:            10 * time.Second,
		Logger:             os.Stderr,
	}
	// Create client object with default config
	client, err := goftp.DialConfig(config, "10.4.33.118")
	if err != nil {
		panic(err)
	}

	defer client.Close()

	file := "file.txt"

	// Upload a file from disk
	bigFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	err = client.Store(file, bigFile)
	if err != nil {
		panic(err)
	}

	err = os.Remove(file)
	if err != nil {

		//如果删除失败则输出 file remove Error!

		fmt.Println("file remove Error!")

		//输出错误详细信息

		fmt.Printf("%s", err)

	} else {

		//如果删除成功则输出 file remove OK!

		fmt.Println("file remove OK!")

	}
}

func StoreFile(w http.ResponseWriter, r *http.Request) {

	config := goftp.Config{
		User:               "F01",
		Password:           "A@1qaz2wsx",
		ConnectionsPerHost: 10,
		Timeout:            10 * time.Second,
		Logger:             os.Stderr,
	}
	// Create client object with default config
	client, err := goftp.DialConfig(config, "10.4.33.118")
	if err != nil {
		panic(err)
	}

	defer client.Close()

	file := "file.txt"

	// Upload a file from disk
	bigFile, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	err = client.Store(file, bigFile)
	if err != nil {
		panic(err)
	}

	err = os.Remove(file)
	if err != nil {

		//如果删除失败则输出 file remove Error!

		fmt.Println("file remove Error!")

		//输出错误详细信息

		fmt.Printf("%s", err)

	} else {

		//如果删除成功则输出 file remove OK!

		fmt.Println("file remove OK!")

	}
}
