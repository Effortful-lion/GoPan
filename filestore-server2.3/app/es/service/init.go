package service

import (
	"filestore-server/config"
	"github.com/olivere/elastic/v7"
	"log"
)

// es 服务
var ES *elastic.Client

// 初始化es
func InitES() {
	var err error
	ES, err = elastic.NewClient(
		elastic.SetURL(config.Config.ESConfig.ES_URL),
		elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}
	log.Println("ES connect success")
}
