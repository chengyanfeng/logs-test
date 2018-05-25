package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"time"
	"gopkg.in/olivere/elastic.v6"
	"context"
	. "logs/models"
	"log"
	"os"
	"strings"
)

var (
	client *elastic.Client
	msgs   *tail.Line
	ok     bool
	now    = time.Now()
)
var message = make(chan string, 10)

type Tweet struct {
	Message string
}

func init() {
	errorlog := log.New(os.Stdout, "APP", log.LstdFlags)
	var err error
	client, err = elastic.NewClient(elastic.SetErrorLog(errorlog), elastic.SetSniff(false), elastic.SetURL(ES_HOST+ES_PORT))
	if err != nil {
		fmt.Println("connect es error", err)
		return
	}
	info, code, err := client.Ping(ES_HOST + ES_PORT).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion, err := client.ElasticsearchVersion(ES_HOST + ES_PORT)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch version %s\n", esversion)
}
func main() {

	go TEST()
	tails, err := tail.TailFile(LOG_PATH, tail.Config{
		ReOpen: true,
		Follow: true,
		// Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	})
	if err != nil {
		fmt.Println("tail file err:", err)
		return
	}

	for true {
		msgs, ok = <-tails.Lines
		if !ok {
			fmt.Printf("tail file close reopen, filename:%s\n", tails.Filename)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		esMessage := string(msgs.Text)
		if !strings.Contains(esMessage, "ELB-HEALTHcHECKER") {
			message <- esMessage
		}
	}

}

func TEST() {
	//检查和创建索引
	for {
		select {
		case a := <-message:
			CreatIndex()
			_, err := client.Index().Index(ES_INDEX).Type(ES_TYPE).BodyJson(string(a)).Do(context.Background())
			if err != nil {
				// Handle error
				panic(err)
				return
			}
		}
		fmt.Println("insert succ")
	}
}

//初始化索引
func IndexExis(index string) {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(index).Do(context.Background())

	if err != nil {
		// Handle error
		panic(err)
	}
	if !exists {
		// Create a new index.
		mapping := `
 {
	"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
	},
	"mapping" : {
      "api-log" : {
        "properties" : {
          "@timestamp" : {
            "type" : "date"
          },
          "@version" : {
            "type" : "integer"
          },
          "body_bytes_sent" : {
            "type" :"long"
          },
          "remote_addr" : {
            "type" :"ip"
          },
          "request_time" : {
            "type" :"float"
          },
          "server_addr" : {
            "type" : "ip"
          },
          "status" : {
            "type" : "integer"
          },
          "upstream_response_time" : {
            "type" : "string"
          }
        }
      }
    }
}
`
		createIndex, err := client.CreateIndex(index).Body(mapping).Do(context.Background())
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}
}

func CreatIndex() {
	time.LoadLocation("Asia/Shanghai")
	day := now.Day()
	//获取当月
	ES_INDEX = ES_INDEX_PREFIX + now.Format("2006-01")

	//如果是月初第一天删除前三个月index
	if day == 25 {
		//检查索引是否存在,不存在即创建
		IndexExis(ES_INDEX)
		//删除前三个月索引
		for i := 1; i < 4; i++ {
			beforTimeIndex := now.AddDate(0, -i, 0).Format("2006-01")
			DelecIndex(beforTimeIndex)
		}
	}
}

//删除index
func DelecIndex(beforTime string) {

	exists, err := client.IndexExists(beforTime).Do(context.Background())

	if err != nil {
		// Handle error
		panic(err)
	}
	if exists {
		// Delete an index.
		deleteIndex, err := client.DeleteIndex(beforTime).Do(context.Background())
		if err != nil {
			// Handle error
			panic(err)
		}
		if !deleteIndex.Acknowledged {
			// Not acknowledged
		}
		println("success")

	}

}
