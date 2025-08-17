package meta

import (
	mydb "filestore-server/app/dbproxy/op"
	"filestore-server/common"
	"sort"
)

// 对外不可见：通过fileSha1获取文件元信息对象
var fileMetas map[string]common.FileMeta

func init() {
	fileMetas = make(map[string]common.FileMeta)
}

// 新增/更新文件元信息
func UpdateFileMeta(fmeta common.FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// 新增/更新文件元信息到mysql中
func UpdateFileMetaDB(fmeta common.FileMeta) bool {
	return mydb.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// 获取文件元信息
func GetFileMeta(fileSha1 string) common.FileMeta {
	return fileMetas[fileSha1]
}

// 从mysql中获取文件元信息
func GetFileMetaDB(fileSha1 string) (*common.FileMeta, error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if err != nil || tfile == nil {
		return nil, err
	}
	fmeta := common.FileMeta{
		FileSha1: fileSha1,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return &fmeta, nil
}

// 批量获取文件元信息列表
func GetLastFileMetas(count int) []common.FileMeta {
	// 修正初始化切片的长度
	fMetaArray := make([]common.FileMeta, 0, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}
	// 显式转换为 ByUploadTime 类型
	byUploadTime := ByUploadTime(fMetaArray)
	sort.Sort(byUploadTime)
	// 处理 count 大于切片长度的情况
	if count > len(byUploadTime) {
		count = len(byUploadTime)
	}
	return byUploadTime[0:count]
}

// 删除文件元信息: TODO 这里只是简单的删除，还需要考虑线程安全问题
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
