package test

import (
	"context"
	"filestore-server/app/dbproxy/mysql"
	"filestore-server/app/gateway/rpc"
	"filestore-server/config"
	"filestore-server/idl/dbproxy/dbproxyPb"
	"testing"
)

// 其他服务需要rpc方式调用数据库代理服务
// 这里在不使用rpc的情况下，测试数据库代理服务功能
func TestDBProxyConn(t *testing.T) {
	config.InitConfig()
	mysql.InitDB()
	db := mysql.DBConn()
	if db == nil {
		t.Log("DBConn failed")
		return
	}
	t.Log("DBConn success")
}

func TestDBProxyExec(t *testing.T) {
	config.InitConfig()
	mysql.InitDB()
	db := mysql.DBConn()

	if db == nil {
		t.Log("DBConn failed")
		return
	}
	t.Log("DBConn success")

	filehash := "1f5e35f31e42ede556c163c3a6eaa4991a1fcb7e"
	filemeta, err := rpc.GetFileMetaHandler(context.Background(), &dbproxyPb.GetFileMetaRequest{
		FileSha1: filehash,
	})
	if err != nil {
		t.Log("GetFileMeta failed")
		return
	}
	t.Log("GetFileMeta success")
	t.Log(filemeta)
}
