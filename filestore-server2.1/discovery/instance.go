// 定义包名为 discovery，用于实现服务发现相关功能
package discovery

// 服务发现

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/resolver"
)

// Server 结构体表示一个服务实例，包含服务的名称、地址、版本和权重信息
type Server struct {
	Name    string `json:"name"`    // 服务名称
	Addr    string `json:"addr"`    // 服务地址
	Version string `json:"version"` // 服务版本
	Weight  int64  `json:"weight"`  // 服务权重
}

// BuildPrefix 根据服务信息构建 etcd 键的前缀
// 如果服务版本为空，则前缀仅包含服务名称；否则，前缀包含服务名称和版本
func BuildPrefix(server Server) string {
	if server.Version == "" {
		return fmt.Sprintf("/%s/", server.Name)
	}

	return fmt.Sprintf("/%s/%s/", server.Name, server.Version)
}

// BuildRegisterPath 根据服务信息构建完整的 etcd 键路径
// 路径由前缀和服务地址拼接而成
func BuildRegisterPath(server Server) string {
	return fmt.Sprintf("%s%s", BuildPrefix(server), server.Addr)
}

// ParseValue 将字节切片解析为 Server 结构体
// 使用 json.Unmarshal 进行解析，如果解析失败则返回错误
func ParseValue(value []byte) (Server, error) {
	server := Server{}
	if err := json.Unmarshal(value, &server); err != nil {
		return server, err
	}

	return server, nil
}

// SplitPath 从 etcd 键路径中提取服务地址信息
// 如果路径格式无效，则返回错误
func SplitPath(path string) (Server, error) {
	server := Server{}
	strs := strings.Split(path, "/")
	if len(strs) == 0 {
		return server, errors.New("invalid path")
	}

	server.Addr = strs[len(strs)-1]

	return server, nil
}

// Exist 辅助函数，用于检查一个地址是否存在于地址列表中
// 遍历地址列表，如果找到匹配的地址则返回 true，否则返回 false
func Exist(l []resolver.Address, addr resolver.Address) bool {
	for i := range l {
		if l[i].Addr == addr.Addr {
			return true
		}
	}

	return false
}

// Remove 辅助函数，用于从地址列表中移除指定地址
// 遍历地址列表，找到匹配的地址后将其移除并返回新的地址列表和移除成功标志
func Remove(s []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr.Addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}

// BuildResolverUrl 根据应用名称构建解析器的 URL
// URL 格式为 schema:///app，其中 schema 为协议方案，app 为应用名称
func BuildResolverUrl(app string) string {
	return schema + ":///" + app
}
