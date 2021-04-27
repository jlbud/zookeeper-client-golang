package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	client2 "zookeeper/client"
)

func callback(event zk.Event) {

}

func main() {
	// 先安转zookeeper
	// 服务器地址列表
	servers := []string{"192.168.5.216:2181"}

	client, err := client2.NewClient(servers, "/api", 10, func(event zk.Event) {
		// zk.EventNodeCreated
		// zk.EventNodeDeleted
		fmt.Println("path: ", event.Path)
		fmt.Println("type: ", event.Type.String())
		fmt.Println("state: ", event.State.String())
		fmt.Println("---------------------------")
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()
	node1 := &client2.ServiceNode{Name: "db", Host: "127.0.0.1", Port: 4000}
	node2 := &client2.ServiceNode{Name: "img", Host: "127.0.0.1", Port: 4001}
	if err := client.Register(node1); err != nil {
		panic(err)
	}
	if err := client.Register(node2); err != nil {
		panic(err)
	}
	// db
	dbNodes, err := client.GetNodes("db")
	if err != nil {
		panic(err)
	}
	for _, node := range dbNodes {
		fmt.Println("dbNode=", node.Host, node.Port)
	}
	// img
	imgNodes, err := client.GetNodes("img")
	if err != nil {
		panic(err)
	}
	for _, node := range imgNodes {
		fmt.Println("mqNode=", node.Host, node.Port)
	}
}
