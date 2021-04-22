package main

import (
	"fmt"
	client2 "zookeeper/client"
)

func main() {
	// 先安转zookeeper
	// 服务器地址列表
	servers := []string{"192.168.5.216:2181"} // 192.168.5.216
	client, err := client2.NewClient(servers, "/api", 10)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	node1 := &client2.ServiceNode{"db", "127.0.0.1", 4000}
	node2 := &client2.ServiceNode{"db", "127.0.0.1", 4001}
	node3 := &client2.ServiceNode{"img", "127.0.0.1", 3309}
	node4 := &client2.ServiceNode{"img", "127.0.0.2", 3309}
	if err := client.Register(node1); err != nil {
		panic(err)
	}
	if err := client.Register(node2); err != nil {
		panic(err)
	}
	if err := client.Register(node3); err != nil {
		panic(err)
	}
	if err := client.Register(node4); err != nil {
		panic(err)
	}
	// db
	dbNodes, err := client.GetNodes("db")
	if err != nil {
		panic(err)
	}
	for _, node := range dbNodes {
		fmt.Println("userNode=", node.Host, node.Port)
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
