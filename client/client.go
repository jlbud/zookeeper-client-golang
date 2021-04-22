package client

import (
	"encoding/json"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

type ServiceNode struct {
	Name string `json:"name"` // 服务名称
	Host string `json:"host"`
	Port int    `json:"port"`
	Url  string `json:"url"`
}

// 在定义一个服务发现的客户端结构体SdClient
type SdClient struct {
	zkServers []string // 多个节点地址
	zkRoot    string   // 服务根节点，这里是/api
	conn      *zk.Conn // zk的客户端连接
}

//编写构造器，创建根节点
func NewClient(zkServers []string, zkRoot string, timeout int) (*SdClient, error) {
	client := new(SdClient)
	client.zkServers = zkServers
	client.zkRoot = zkRoot
	// 连接服务器
	conn, _, err := zk.Connect(zkServers, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}
	client.conn = conn
	// 创建服务根节点
	if err := client.ensureRoot(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

// 关闭连接，释放临时节点
func (s *SdClient) Close() {
	s.conn.Close()
}

func (s *SdClient) ensureRoot() error {
	exists, _, err := s.conn.Exists(s.zkRoot)
	if err != nil {
		return err
	}
	if !exists {
		_, err := s.conn.Create(s.zkRoot, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}

//值得注意的是代码中的Create调用可能会返回节点已存在错误，这是正常现象，因为会存在多进程同时创建节点的可能。如果创建根节点出错，还需要及时关闭连接。我们不关心节点的权限控制，所以使用zk.WorldACL(zk.PermAll)表示该节点没有权限限制。Create参数中的flag = 0表示这是一个持久化的普通节点。
//接下来我们编写服务注册方法
func (s *SdClient) Register(node *ServiceNode) error {
	if err := s.ensureName(node.Name); err != nil {
		return err
	}
	path := s.zkRoot + "/" + node.Name + "/n"
	data, err := json.Marshal(node)
	if err != nil {
		return err
	}
	_, err = s.conn.CreateProtectedEphemeralSequential(path, data, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err
	}
	return nil
}
func (s *SdClient) ensureName(name string) error {
	path := s.zkRoot + "/" + name
	exists, _, err := s.conn.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		_, err := s.conn.Create(path, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}

//先要创建/api/db节点作为服务列表的父节点。然后创建一个保护顺序临时(ProtectedEphemeralSequential)子节点，同时将地址信息存储在节点中。什么叫保护顺序临时节点，首先它是一个临时节点，会话关闭后节点自动消失。其它它是个顺序节点，zookeeper自动在名称后面增加自增后缀，确保节点名称的唯一性。同时还是个保护性节点，节点前缀增加了GUID字段，确保断开重连后临时节点可以和客户端状态对接上。
//接下来我们实现消费者获取服务列表方法
func (s *SdClient) GetNodes(name string) ([]*ServiceNode, error) {
	path := s.zkRoot + "/" + name
	// 获取字节点名称
	childs, _, err := s.conn.Children(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return []*ServiceNode{}, nil
		}
		return nil, err
	}
	nodes := []*ServiceNode{}
	for _, child := range childs {
		fullPath := path + "/" + child
		data, _, err := s.conn.Get(fullPath)
		if err != nil {
			if err == zk.ErrNoNode {
				continue
			}
			return nil, err
		}
		node := new(ServiceNode)
		err = json.Unmarshal(data, node)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (s *SdClient) GetChildren(name string) ([]string, error) {
	path := s.zkRoot + "/" + name
	// 获取字节点名称 [_c_9ca18b1c5c821874d7241eaeef76f105-n0000000007,_c_9ca18b1c5c821874d7241eaeef76f105-n0000000008]
	children, _, err := s.conn.Children(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return make([]string, 0), nil
		}
		return nil, err
	}
	return children, nil
}

func (s *SdClient) Delete(name string) error {
	path := s.zkRoot + "/" + name
	_, sate, err := s.conn.Get(path)
	if err != nil {
		return err
	}
	err = s.conn.Delete(path, sate.Version)
	if err != nil {
		return err
	}
	return nil
}

// 删改与增不同在于其函数中的version参数,其中version是用于 CAS支持
// 可以通过此种方式保证原子性
// 改
func (s *SdClient) Modify(childPath string, newData []byte) error {
	path := s.zkRoot + "/" + childPath
	_, sate, _ := s.conn.Get(path)
	_, err := s.conn.Set(path, newData, sate.Version)
	if err != nil {
		return err
	}
	return nil
}
