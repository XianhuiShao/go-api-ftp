package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sap/gorfc/gorfc"
	"github.com/secsy/goftp"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
	//Array string  `json:"array"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

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
	router.GET("/albums", getAlbums)

	router.POST("/api/store", postAlbums)

	router.POST("/api/sync_data", sync)

	router.Run(":50016")
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {

	// a := c.Request

	// var a string
	// a = "edsr"
	// c.IndentedJSON(http.StatusOK, a)

}

func postAlbums(c *gin.Context) {
	name := c.PostForm("name")

	fmt.Printf("name: %s", name)

	//a := c.Request.Body

	//c.IndentedJSON(http.StatusOK, name)
	//c.String(http.StatusOK, string(name))

	//len := c.Request.ContentLength

	//c.String(http.StatusOK, len)
	//a1 :=c.Bind()

	decoder := json.NewDecoder(c.Request.Body)
	var metaData Meta

	err := decoder.Decode(&metaData)
	if err != nil {
		panic(err)

	}
	b, err := json.Marshal(metaData)
	if err != nil {
		fmt.Println(err)
		return
	}

	// len := c.Request.ContentLength // 获取请求实体长度
	// body := make([]byte, len)      // 创建存放请求实体的字节切片
	// c.Request.Body.Read(body)
	// bytes := make([]byte, len)
	// bytes, err = io.ReadAll(c.Request.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// body1 := string(bytes)
	// body2 := string(body)
	// fmt.Println(body1)
	// fmt.Println(body2)

	str := string(b)

	var filename string

	currentTime := time.Now()
	formattedTime := currentTime.Format("20060102_150405")
	filename = formattedTime + ".json"

	//if err := os.WriteFile("file1.txt", []byte("Hello GOSAMPLES!"), 0666); err != nil {
	if err := os.WriteFile(filename, []byte(str), 0666); err != nil {
		log.Fatal(err)
	}
	c.String(http.StatusOK, str)

	ftp(filename)
}

func AddPost(w http.ResponseWriter, r *http.Request) {
	len := r.ContentLength          // 获取请求实体长度
	body := make([]byte, len)       // 创建存放请求实体的字节切片
	r.Body.Read(body)               // 调用 Read 方法读取请求实体并将返回内容存放到上面创建的字节切片
	io.WriteString(w, string(body)) // 将请求实体作为响应实体返回
}

func ftp(filename string) {

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

	//	file := "file.txt"

	// Upload a file from disk
	bigFile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	err = client.Store(filename, bigFile)
	if err != nil {
		panic(err)
	}

	err = os.Remove(filename)
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

func sync(c *gin.Context) {

	var reqParam req

	decoder := json.NewDecoder(c.Request.Body)

	// c.Request.Body.Read()
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
			"EXT":       "",
			"SPRAS":     reqParam.Spras,
			"AS_URL":    reqParam.AsURL,
			"DEBUG":     reqParam.Debug,
			"ATTA_MODE": reqParam.AttaMode,
		},
		"CONDITION": table,
	}

	funcname := "ZAS_CAPP_NOTICE"

	attrs, _ := c.GetConnectionAttributes()
	fmt.Println("Connection attributes", attrs)

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
		"user":   "user01",
		"passwd": "Sap@202301",
		"ashost": "10.4.112.97",
		"sysnr":  "00",
		"client": "800",
		"lang":   "ZH",
	}
}
