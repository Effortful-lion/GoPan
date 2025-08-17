package service

import (
	"context"
	"encoding/json"
	"filestore-server/idl/es/esPb"
	"github.com/olivere/elastic/v7"
	"log"
)

type ESService struct {
	esPb.UnimplementedESServiceServer
}

func NewESService() *ESService {
	return &ESService{}
}

// 查询所有的文件哈希集合
func (s *ESService) GetFileHashList(ctx context.Context, req *esPb.SearchReq) (*esPb.SearchResp, error) {
	filename := req.Key
	username := req.Username
	index := req.Index
	start := req.Start
	size := req.Size

	// 直接搜索到全部内容
	// 创建复杂查询
	query := elastic.NewBoolQuery()
	// 准备查询条件
	es_match_query := elastic.NewMatchQuery("file_name", filename)
	es_match_query2 := elastic.NewMatchQuery("username", username)
	query.Must(es_match_query)
	query.Must(es_match_query2)
	res, err := ES.Search().Index(index).Query(query).From(int(start)).Size(int(size)).Do(context.Background())
	if err != nil {
		log.Println(err)
		return &esPb.SearchResp{}, err
	}

	result := &esPb.SearchResp{}
	var re []*esPb.SearchResult
	// 准备结果，序列化到结构体
	for _, hit := range res.Hits.Hits {
		temp := esPb.SearchResult{}
		source := hit.Source
		err := json.Unmarshal(source, &temp)
		if err != nil {
			log.Println(err)
			return &esPb.SearchResp{}, err
		}
		re = append(re, &temp)
	}
	result.Res = re
	return result, nil
}
