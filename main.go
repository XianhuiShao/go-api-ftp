package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

func main() {

	router := gin.Default()

	router.POST("/api/store", storeFile)
	router.GET("/api/query", query)

	router.POST("/api/sync_data", sync)

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
