package mq

import cmn "filestore-server/common"

// 定义消息格式
type TransferData struct {
	FileHash      string        // 文件hash
	CurLocation   string        // 当前存储位置（可能是临时文件的地址，最终是目标位置）
	DestLocation  string        // 目标存储位置
	DestStoreType cmn.StoreType // 存储类型
}
