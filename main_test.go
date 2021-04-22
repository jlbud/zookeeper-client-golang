package main

import (
	"encoding/json"
	"testing"
	"time"
	client2 "zookeeper/client"
)

var client *client2.SdClient

// 自己客户端的服务地址，
// 只注册自己能提供的服务，
// 如果注册其它IP提供的服务（这里可以做个限制，自动获取本机IP），那么其它IP服务是否可用自己不清楚
const Self_Node = "127.0.0.1"

func init() {
	var err error
	servers := []string{"192.168.5.216:2181"}
	client, err = client2.NewClient(servers, "/server", 10)
	if err != nil {
		panic(err)
	}
}

// 节点注册后会一直存在zookeeper中，
// 节点下的信息如果客户端断开后，心跳消失，信息会自动消除，
// 注册本节点能提供的服务信息
func TestRegister(t *testing.T) {
	defer client.Close()
	// 注册本节点能提供的服务
	// 127.0.0.1:4001端口提供消息队列服务
	// 8080端口提供http服务
	// 本节点能提供两个nsqd服务，和一个http服务
	node1 := &client2.ServiceNode{Name: "nsqd", Host: Self_Node, Port: 4001}
	node2 := &client2.ServiceNode{Name: "nsqd", Host: Self_Node, Port: 4002}
	node3 := &client2.ServiceNode{Name: "http", Host: Self_Node, Port: 8080}
	if err := client.Register(node1); err != nil {
		panic(err)
	}
	if err := client.Register(node2); err != nil {
		panic(err)
	}
	if err := client.Register(node3); err != nil {
		panic(err)
	}
	time.Sleep(100 * time.Second)
}

func TestModify(t *testing.T) {
	//defer client.Close()
	childs, err := client.GetChildren("http")
	if err != nil {
		t.Error(err)
	}
	if len(childs) > 0 {
		node3 := &client2.ServiceNode{Name: "http", Host: "127.0.0.1", Port: 9090}
		b, _ := json.Marshal(node3)
		err := client.Modify("http/"+childs[0], b)
		if err != nil {
			t.Error(err)
		}
	} else {
		t.Log("children len is 0")
	}

}

func TestDelete(t *testing.T) {
	//defer client.Close()
	err := client.Delete("http")
	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	//defer client.Close()
	// 注意
	// 如果节点下有数据，该节点不能删除
	dbNodes, err := client.GetNodes("http")
	if err != nil {
		panic(err)
	}
	t.Log(dbNodes)
	for _, node := range dbNodes {
		t.Log("db node=", node.Host, node.Port)
	}
}
