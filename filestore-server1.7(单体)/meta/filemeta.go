package meta

import (
	mydb "filestore-server/db"
	"sort"
)

// 文件元信息
type FileMeta struct {
	FileSha1 string // 基于 SHA1 算法计算的文件哈希值，唯一标志
	FileName string // 文件名
	FileSize int64  // 文件大小
	Location string // 文件存储路径
	UploadAt string // 文件上传时间
}

// 对外不可见：通过fileSha1获取文件元信息对象
var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// 新增/更新文件元信息
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// 新增/更新文件元信息到mysql中
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return mydb.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// 获取文件元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// 从mysql中获取文件元信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if err != nil || tfile == nil {
		return nil, err
	}
	fmeta := FileMeta{
		FileSha1: fileSha1,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return &fmeta, nil
}

// 批量获取文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}
	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}

// 删除文件元信息: TODO 这里只是简单的删除，还需要考虑线程安全问题
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
