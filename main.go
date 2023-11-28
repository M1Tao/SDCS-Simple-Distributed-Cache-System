package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	// 假设 totalNodes 是系统中节点的总数
	totalNodes = 3
	nodes      = []string{"localhost:9527", "localhost:9528", "localhost:9529"}
)

func main() {
	// 从命令行参数获取当前节点索引
	var currentNodeIndex int
	flag.IntVar(&currentNodeIndex, "index", 0, "Index of the current node")
	flag.Parse()

	if currentNodeIndex < 0 || currentNodeIndex >= totalNodes {
		log.Fatalf("Invalid node index: %d", currentNodeIndex)
	}

	// 创建缓存实例
	cache := NewCache()

	// 设置 gRPC 服务器地址（假设为 HTTP 端口 + 1000）
	grpcPort := fmt.Sprintf("localhost:%d", 1000+9527+currentNodeIndex)
	log.Printf("Starting gRPC server on %s\n", grpcPort)
	go startGRPCServer(grpcPort, cache)

	// 设置 HTTP 服务器地址
	httpPort := fmt.Sprintf("localhost:%d", 9527+currentNodeIndex)
	log.Printf("Starting HTTP server on %s\n", httpPort)
	startHTTPServer(nodes[currentNodeIndex], cache, nodes, currentNodeIndex)
}
