package discovery

// 服务发现解析器
// 作用：动态获取服务的地址信息

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

// 定义常量，指定此解析器支持的协议方案为 etcd
const (
	schema = "etcd"
)

// Resolver 实现了 gRPC 的解析器接口，用于在 etcd 中发现服务地址
// 它负责与 etcd 交互，监听服务地址的变化，并更新 gRPC 客户端的连接状态
type Resolver struct {
	schema      string   // 支持的协议方案
	EtcdAddrs   []string // etcd 服务器的地址列表
	DialTimeout int      // 连接 etcd 的超时时间

	closeCh      chan struct{}      // 用于关闭解析器的通道
	watchCh      clientv3.WatchChan // etcd 的 watch 通道，用于监听键值对的变化
	cli          *clientv3.Client   // etcd 客户端实例
	keyPrifix    string             // etcd 中存储服务信息的键前缀
	srvAddrsList []resolver.Address // 服务地址列表

	cc     resolver.ClientConn // gRPC 客户端连接
	logger *logrus.Logger      // 日志记录器
}

// NewResolver 创建一个新的基于 etcd 的解析器实例
// 参数 etcdAddrs 是 etcd 服务器的地址列表，logger 是日志记录器
func NewResolver(etcdAddrs []string, logger *logrus.Logger) *Resolver {
	return &Resolver{
		schema:      schema,
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Scheme 返回该解析器支持的协议方案 : etcd
func (r *Resolver) Scheme() string {
	return r.schema
}

// Build 根据给定的目标创建一个新的解析器实例
// 参数 target 是解析的目标，cc 是 gRPC 客户端连接，opts 是构建选项
// 返回值为解析器实例和可能出现的错误
func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc // 设置客户端连接

	// 生成 etcd 键前缀
	r.keyPrifix = BuildPrefix(Server{Name: target.Endpoint(), Version: target.URL.Host})
	// 启动解析器
	if _, err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

// ResolveNow 实现了 resolver.Resolver 接口，用于立即触发解析操作
// 目前此方法为空实现
func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

// Close 实现了 resolver.Resolver 接口，用于关闭解析器
// 向 closeCh 通道发送一个空结构体来触发关闭操作
func (r *Resolver) Close() {
	r.closeCh <- struct{}{}
}

// start 启动解析器，包括创建 etcd 客户端、注册解析器、同步服务地址信息和启动 watch 协程
// 返回一个用于关闭解析器的通道和可能出现的错误
func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	// 创建 etcd 客户端
	r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	// 注册解析器
	resolver.Register(r)

	// 创建关闭通道
	r.closeCh = make(chan struct{})

	// 同步服务地址信息
	if err = r.sync(); err != nil {
		return nil, err
	}

	// 启动 watch 协程
	go r.watch()

	return r.closeCh, nil
}

// watch 监听 etcd 中键值对的变化
// 处理关闭信号、watch 事件和定期同步服务地址信息
func (r *Resolver) watch() {
	// 创建一个每分钟触发一次的定时器
	ticker := time.NewTicker(time.Minute)
	// 启动 etcd 的 watch 操作，监听以 keyPrifix 为前缀的键值对变化
	r.watchCh = r.cli.Watch(context.Background(), r.keyPrifix, clientv3.WithPrefix())

	for {
		select {
		case <-r.closeCh:
			return // 收到关闭信号，退出循环
		case res, ok := <-r.watchCh:
			if ok {
				// 处理 watch 事件
				r.update(res.Events)
			}
		case <-ticker.C:
			// 定期同步服务地址信息
			if err := r.sync(); err != nil {
				r.logger.Error("sync failed", err)
			}
		}
	}
}

// update 处理 etcd 中的事件，包括添加和删除服务地址信息
// 并更新客户端连接的状态
func (r *Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		var info Server
		var err error

		switch ev.Type {
		case clientv3.EventTypePut:
			// 解析新添加的服务信息
			info, err = ParseValue(ev.Kv.Value)
			if err != nil {
				continue
			}
			// 创建服务地址结构体
			addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
			if !Exist(r.srvAddrsList, addr) {
				// 如果地址不存在，则添加到地址列表
				r.srvAddrsList = append(r.srvAddrsList, addr)
				// 更新客户端连接状态
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		case clientv3.EventTypeDelete:
			// 解析要删除的服务信息
			info, err = SplitPath(string(ev.Kv.Key))
			if err != nil {
				continue
			}
			// 创建服务地址结构体
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.srvAddrsList, addr); ok {
				// 如果地址存在，则从地址列表中移除
				r.srvAddrsList = s
				// 更新客户端连接状态
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		}
	}
}

// sync 同步获取 etcd 中所有服务地址信息，并更新客户端连接的状态
func (r *Resolver) sync() error {
	// 创建一个带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 从 etcd 中获取以 keyPrifix 为前缀的所有键值对
	res, err := r.cli.Get(ctx, r.keyPrifix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	// 清空服务地址列表
	r.srvAddrsList = []resolver.Address{}

	for _, v := range res.Kvs {
		// 解析服务信息
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}
		// 创建服务地址结构体
		addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
		// 添加到服务地址列表
		r.srvAddrsList = append(r.srvAddrsList, addr)
	}
	// 更新客户端连接状态
	r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
	return nil
}
