package meta

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
func UpdateFileMetaDB(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// 获取文件元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}
