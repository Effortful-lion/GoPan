package common

// 存储类型(表示文件存到哪里)
type StoreType int

const (
	_ StoreType = iota
	// StoreLocal : 节点本地
	StoreLocal
	// StoreCeph : Ceph集群
	StoreCeph
	// StoreOSS : 阿里OSS
	StoreOSS
	// StoreMix : 混合(Ceph及OSS)
	StoreMix
	// StoreAll : 所有类型的存储都存一份数据
	StoreAll
)

// 文件元信息
type FileMeta struct {
	FileSha1 string // 基于 SHA1 算法计算的文件哈希值，唯一标志
	FileName string // 文件名
	FileSize int64  // 文件大小
	Location string // 文件存储路径
	UploadAt string // 文件上传时间
}

// 用户文件表结构
type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// 分块上传初始化信息
type MultipartUploadInfo struct {
	FileHash   string // 文件hash值
	FileSize   int    // 文件总大小
	UploadID   string // 分块上传的唯一标识符
	ChunkSize  int    // 分块大小
	ChunkCount int    // 分块数量
}

type User struct {
	UserName     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}
