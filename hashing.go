package main

import "hash/fnv"

// getNodeForKey 根据提供的 key 计算并返回应该处理该 key 的节点索引。
func getNodeForKey(key string, totalNodes int) int {
	hasher := fnv.New32a() // 使用 FNV-1a 哈希算法
	hasher.Write([]byte(key))
	hash := hasher.Sum32()
	return int(hash) % totalNodes // 计算节点索引
}
