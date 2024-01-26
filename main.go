package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sap/gorfc/gorfc"
	"github.com/secsy/goftp"
)

type req struct {
	InterfaceID string `json:"interface_id"`
	Pattern     struct {
		PatternType string `json:"pattern_type"`
		Datefrom    string `json:"datefrom"`
		Dateto      string `json:"dateto"`
	} `json:"pattern"`
	AsURL     string `json:"as_url"`
	Debug     string `json:"debug"`
	Spras     string `json:"spras"`
	AttaMode  string `json:"atta_mode"`
	Condition []struct {
		TABLE_NAME string `json:"table_name"`
		TABLE_TYPE string `json:"table_type"`
		FIELDNAME  string `json:"fieldname"`
		LOW        string `json:"low"`
		HIGH       string `json:"high"`
	} `json:"condition"`
}

type Meta struct {
	InterfaceID string `json:"INTERFACE_ID"`
	Ext         string `json:"EXT"`
	TotalStep   int    `json:"TOTAL_STEP"`
	CurrentStep int    `json:"CURRENT_STEP"`
	Output      struct {
		Data []struct {
			SapObject string `json:"SAP_OBJECT"`
			ObjectID  string `json:"OBJECT_ID"`
			ArchivID  string `json:"ARCHIV_ID"`
			ArcDocID  string `json:"ARC_DOC_ID"`
			ArDate    string `json:"AR_DATE"`
			Reserve   string `json:"RESERVE"`
			DelFlag   string `json:"DEL_FLAG"`
			Filename  string `json:"FILENAME"`
			Creator   string `json:"CREATOR"`
			Descr     string `json:"DESCR"`
			Creatime  string `json:"CREATIME"`
			AttaMode  string `json:"ATTA_MODE"`
		} `json:"DATA"`
	} `json:"OUTPUT"`
}

type authStruct struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type fileBeginReqStruct struct {
	Docid  string `json:"docid"`
	Length int    `json:"length"`
	Name   string `json:"name"`
}

type fileBeginRspStruct struct {
	Authrequest []string `json:"authrequest"`
	Docid       string   `json:"docid"`
	Name        string   `json:"name"`
	Rev         string   `json:"rev"`
}

// type filePorcessingRspStruct struct {
// 	Authrequest []string `json:"authrequest"`
// 	Docid       string   `json:"docid"`
// 	Name        string   `json:"name"`
// 	Rev         string   `json:"rev"`
// }

type fileEndReqStruct struct {
	Docid string `json:"docid"`
	Rev   string `json:"rev"`
}

func main() {

	router := gin.Default()

	router.POST("/api/store", storeFile)
	router.GET("/api/query", query)

	router.POST("/api/sync_data", sync)

	//Upload file into AnyShare
	router.POST("/api/uploadfile", uploadFile)

	router.Run(":50016")
}
func query(c *gin.Context) {

	decoder := json.NewDecoder(c.Request.Body)
	var reqBody map[string]interface{}

	err := decoder.Decode(&reqBody)
	if err != nil {
		panic(err)

	}

	c.JSON(http.StatusOK, reqBody)
}

func storeFile(c *gin.Context) {

	//解析Form-data 传入参数
	name := c.PostForm("name")

	fmt.Printf("name: %s", name)

	decoder := json.NewDecoder(c.Request.Body)
	var reqBody map[string]interface{}

	err := decoder.Decode(&reqBody)
	if err != nil {
		panic(err)

	}
	//c.String(http.StatusOK, str)

	callFTP(reqBody)
}

func callFTP(reqBody map[string]interface{}) {

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

	body, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 将json格式转换成io.reader 类型，直接发送到FTP
	// 参考：https://gosamples.dev/struct-to-io-reader/
	reader := bytes.NewReader(body)
	if err != nil {
		log.Fatal(err)
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("20060102_150405")

	//文件路径 + 文件名
	file := "file/" + formattedTime + ".json"

	//func (c *Client) Store(path string, src io.Reader) error
	err = client.Store(file, reader)
	if err != nil {
		panic(err)
	}

}

func sync(c *gin.Context) {

	var reqParam req

	decoder := json.NewDecoder(c.Request.Body)

	err := decoder.Decode(&reqParam)
	if err != nil {
		panic(err)

	}

	CallRfc(reqParam)
}

func CallRfc(reqParam req) {
	c, err := gorfc.ConnectionFromParams(abapSystem())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connected:", c.Alive())

	table := reqParam.Condition
	params := map[string]interface{}{
		"CAPP_INPUT": map[string]interface{}{
			"INTERFACE_ID": reqParam.InterfaceID,
			"PATTERN": map[string]interface{}{
				"PATTERN_TYPE": reqParam.Pattern.PatternType,
				"DATEFROM":     reqParam.Pattern.Datefrom,
				"DATETO":       reqParam.Pattern.Dateto,
			},
			"EXT":    "",
			"SPRAS":  reqParam.Spras,
			"AS_URL": reqParam.AsURL,
			"DEBUG":  reqParam.Debug,
		},
		"CONDITION": table,
	}

	funcname := "ZAS_CAPP_NOTICE"

	//attrs, _ := c.GetConnectionAttributes()
	//fmt.Println("Connection attributes", attrs)

	//params := map[string]interface{}{}
	r, e := c.Call(funcname, params)

	if e != nil {
		fmt.Println(e)
		return
	}
	// 输出结果
	fmt.Printf("Response: %#v \n", r)

	c.Close()
}
func abapSystem() gorfc.ConnectionParameters {
	return gorfc.ConnectionParameters{
		"user":   "AS_HUI",
		"passwd": "Sap@202201",
		"ashost": "10.4.112.97",
		"sysnr":  "00",
		"client": "800",
		"lang":   "ZH",
	}
}

func uploadFile(c *gin.Context) {

	// // 从请求body中读取内容
	// body, err := io.ReadAll(c.Request.Body)
	// if err != nil {
	// 	c.String(http.StatusBadRequest, "Bad request")
	// 	return
	// }
	// len := len(body)
	// str := string(body)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "文件上传失败")
	}
	body, err := io.ReadAll(file)

	//fmt.Println(file, header, body)

	//获取Token
	token := getToken()

	//开始上传文件
	fileBeginRsp := fileBegin(token, int(header.Size), header.Filename)

	// 处理上传文件
	fileProcessing(token, body, fileBeginRsp)

	// //结束上传文件
	fileEnd(token, fileBeginRsp)
}

func getToken() string {

	authorization := "Basic MTNjOGQ2NmUtN2Q0OC00ZWY2LWE0Y2EtYzY3NGU2ODExNTgyOjExMTExMQ=="
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	urlValues := url.Values{}

	//添加Form Body字段值
	urlValues.Add("grant_type", "client_credentials")
	urlValues.Add("scope", "all")

	reqBody := urlValues.Encode()
	requestPostURL := "https://10.4.132.181:443/oauth2/token"
	req, err := http.NewRequest(http.MethodPost, requestPostURL, strings.NewReader(reqBody))

	//添加Header
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Add("Authorization", authorization)

	if err != nil {
		log.Println(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	var auth authStruct
	err = json.NewDecoder(resp.Body).Decode(&auth)
	if err != nil {
		log.Println(err)
	}
	token := auth.AccessToken

	defer resp.Body.Close()
	return token
}

func fileBegin(token string, len int, filename string) fileBeginRspStruct {

	authorization := "Bearer " + token
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var data fileBeginReqStruct
	data.Docid = "gns://AFC10D84B461408EAD3CEBA6E0EC136F"
	data.Length = len
	data.Name = filename
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)

	}

	requestPostURL := "https://10.4.132.181:443/api/efast/v1/file/osbeginupload"
	req, err := http.NewRequest(http.MethodPost, requestPostURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
	}
	//添加Header
	// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Add("Authorization", authorization)
	// fmt.Println(string(authorization))

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	var fileBeginRsp fileBeginRspStruct
	err = json.NewDecoder(resp.Body).Decode(&fileBeginRsp)
	if err != nil {
		log.Println(err)
	}
	return fileBeginRsp
}

func fileProcessing(token string, body []byte, fileBeginRsp fileBeginRspStruct) {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// var data fileBeginReqStruct
	// data.Docid = "gns://AFC10D84B461408EAD3CEBA6E0EC136F"
	// // data.Length = len
	// data.Name = "1.docx"
	// jsonData, err := json.Marshal(data)
	// if err != nil {
	// 	fmt.Println("Error marshalling JSON:", err)

	// }
	requestPostURL := fileBeginRsp.Authrequest[1]

	// var body1 =
	req, err := http.NewRequest(http.MethodPut, requestPostURL, bytes.NewReader(body))

	//添加Header

	//Content-Type
	req.Header.Add("Content-Type", "application/octet-stream")

	//date
	str := fileBeginRsp.Authrequest[4]
	fmt.Println(str)
	lenStr := len(str)
	slice := str[12:lenStr]
	//slice := strings.Split(str, ":")
	req.Header.Add("x-amz-date", slice)

	//Authorization
	str = fileBeginRsp.Authrequest[2]
	lenStr = len(str)
	slice = str[15:lenStr]
	req.Header.Add("Authorization", slice)

	// fmt.Println(string(authorization))
	if err != nil {
		log.Println(err)
	}
	fmt.Println(req.Header)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	respData, _ := io.ReadAll(req.Body)
	// var fileProcessingRsp filePorcessingRspStruct
	// err = json.NewDecoder(resp.Body).Decode(&fileProcessingRsp)
	// if err != nil {
	// 	log.Println(err)
	// }
	fmt.Println("fileProcessing", resp.Status)
	fmt.Println("res", respData)
	// return fileBeginRsp
}
func fileEnd(token string, fileBeginRsp fileBeginRspStruct) {

	authorization := "Bearer " + token
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var data fileEndReqStruct
	data.Docid = fileBeginRsp.Docid
	data.Rev = fileBeginRsp.Rev
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)

	}

	requestPostURL := "https://10.4.132.181:443/api/efast/v1/file/osendupload"
	req, err := http.NewRequest(http.MethodPost, requestPostURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
	}
	//添加Header
	// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Add("Authorization", authorization)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(resp.Status)
}
