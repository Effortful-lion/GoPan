package discovery

// 服务注册

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Register 结构体用于管理服务在 etcd 中的注册信息
type Register struct {
	EtcdAddrs   []string // etcd 服务器的地址列表
	DialTimeout int      // 连接 etcd 的超时时间

	closeCh     chan struct{}                           // 用于关闭服务注册的通道
	leasesID    clientv3.LeaseID                        // etcd 的租约 ID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse // 用于保持租约活跃的通道

	srvInfo Server           // 服务的详细信息
	srvTTL  int64            // 服务租约的过期时间
	cli     *clientv3.Client // etcd 客户端实例
	logger  *logrus.Logger   // 日志记录器
}

// NewRegister 创建一个基于 etcd 的服务注册实例
// 参数 etcdAddrs 是 etcd 服务器的地址列表，logger 是日志记录器
// 返回一个指向 Register 结构体的指针
func NewRegister(etcdAddrs []string, logger *logrus.Logger) *Register {
	return &Register{
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Register 用于将服务信息注册到 etcd 中
// 参数 srvInfo 是服务的详细信息，ttl 是服务租约的过期时间
// 返回一个用于关闭注册的通道和可能出现的错误
func (r *Register) Register(srvInfo Server, ttl int64) (chan<- struct{}, error) {
	var err error

	// 检查服务地址的 IP 部分是否有效
	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return nil, errors.New("invalid ip address")
	}

	// 创建 etcd 客户端实例
	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return nil, err
	}

	// 保存服务信息和租约过期时间
	r.srvInfo = srvInfo
	r.srvTTL = ttl

	// 执行注册操作
	if err = r.register(); err != nil {
		return nil, err
	}

	// 创建关闭通道
	r.closeCh = make(chan struct{})

	// 启动保持租约活跃的协程
	go r.keepAlive()

	return r.closeCh, nil
}

// register 执行实际的注册操作，包括获取租约、保持租约活跃和将服务信息存入 etcd
// 返回可能出现的错误
func (r *Register) register() error {
	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	// 获取 etcd 租约
	leaseResp, err := r.cli.Grant(ctx, r.srvTTL)
	if err != nil {
		return err
	}

	// 保存租约 ID
	r.leasesID = leaseResp.ID

	// 启动租约保持活跃机制
	if r.keepAliveCh, err = r.cli.KeepAlive(context.Background(), r.leasesID); err != nil {
		return err
	}

	// 将服务信息序列化为 JSON 格式
	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}

	// 将服务信息存入 etcd，并关联租约
	_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))

	return err
}

// Stop 停止服务注册，包括取消注册和撤销租约
func (r *Register) Stop() {
	// 注释掉的代码原计划通过 closeCh 通道触发关闭操作
	// r.closeCh <- struct{}{}
	// 执行取消注册操作
	if err := r.unregister(); err != nil {
		r.logger.Error("unregister failed, error: ", err)
	}

	// 撤销租约
	if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
		r.logger.Error("revoke failed, error: ", err)
	}
}

// unregister 从 etcd 中删除服务注册信息
// 返回可能出现的错误
func (r *Register) unregister() error {
	// 从 etcd 中删除服务信息
	_, err := r.cli.Delete(context.Background(), BuildRegisterPath(r.srvInfo))
	return err
}

// keepAlive 保持服务租约活跃，处理租约过期或丢失的情况
func (r *Register) keepAlive() {
	// 创建一个定时器，周期为租约过期时间
	ticker := time.NewTicker(time.Duration(r.srvTTL) * time.Second)

	for {
		select {
		// 注释掉的代码原计划处理关闭信号
		// issues:https://github.com/CocaineCong/grpc-todoList/issues/19
		// case <-r.closeCh:
		//	if err := r.unregister(); err != nil {
		//		r.logger.Error("unregister failed, error: ", err)
		//	}
		//
		//	if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
		//		r.logger.Error("revoke failed, error: ", err)
		//	}
		// 处理租约保持活跃的响应
		case res := <-r.keepAliveCh:
			if res == nil {
				// 若响应为空，重新注册服务
				if err := r.register(); err != nil {
					r.logger.Error("register failed, error: ", err)
				}
			}
		// 定时器触发
		case <-ticker.C:
			if r.keepAliveCh == nil {
				// 若租约保持活跃通道为空，重新注册服务
				if err := r.register(); err != nil {
					r.logger.Error("register failed, error: ", err)
				}
			}
		}
	}
}

// UpdateHandler 返回一个 HTTP 处理函数，用于更新服务的权重信息
func (r *Register) UpdateHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 从请求 URL 中获取权重参数
		weightstr := req.URL.Query().Get("weight")
		// 将权重参数转换为整数
		weight, err := strconv.Atoi(weightstr)
		if err != nil {
			// 若转换失败，返回 400 错误
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// 定义更新服务信息的匿名函数
		var update = func() error {
			// 更新服务信息中的权重
			r.srvInfo.Weight = int64(weight)
			// 将更新后的服务信息序列化为 JSON 格式
			data, err := json.Marshal(r.srvInfo)
			if err != nil {
				return err
			}

			// 将更新后的服务信息存入 etcd，并关联租约
			_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
			return err
		}

		// 执行更新操作
		if err := update(); err != nil {
			// 若更新失败，返回 500 错误
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// 更新成功，返回成功信息
		_, _ = w.Write([]byte("update server weight success"))
	})
}

// GetServerInfo 从 etcd 中获取服务的详细信息
// 返回服务信息和可能出现的错误
func (r *Register) GetServerInfo() (Server, error) {
	// 从 etcd 中获取服务信息
	resp, err := r.cli.Get(context.Background(), BuildRegisterPath(r.srvInfo))
	if err != nil {
		return r.srvInfo, err
	}

	// 初始化服务信息结构体
	server := Server{}
	if resp.Count >= 1 {
		// 若获取到服务信息，将其反序列化为 Server 结构体
		if err := json.Unmarshal(resp.Kvs[0].Value, &server); err != nil {
			return server, err
		}
	}

	return server, err
}
