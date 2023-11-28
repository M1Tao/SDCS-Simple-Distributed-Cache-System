package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// startHTTPServer 启动 HTTP 服务器并定义路由。
func startHTTPServer(address string, cache *Cache, nodes []string, currentNodeIndex int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := extractKeyFromRequest(r)
		responsibleNodeIndex := getNodeForKey(key, len(nodes))

		if responsibleNodeIndex == currentNodeIndex {
			// 当前节点负责处理请求
			switch r.Method {
			case http.MethodGet:
				handleGet(w, r, cache, key)
			case http.MethodPost:
				handlePost(w, r, cache)
			case http.MethodDelete:
				handleDelete(w, r, cache, key)
			default:
				http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
			}
		} else {
			// 转发请求到负责的节点
			forwardRequest(w, r, responsibleNodeIndex, cache)
		}
	})

	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// handleGet 处理 GET 请求。
func handleGet(w http.ResponseWriter, r *http.Request, cache *Cache, key string) {
	log.Println("Received Get request:", r.URL.Path)
	if value, ok := cache.Get(key); ok {
		respondWithJSON(w, http.StatusOK, map[string]string{key: value})
	} else {
		http.NotFound(w, r)
	}
}

// handlePost 处理 POST 请求。
func handlePost(w http.ResponseWriter, r *http.Request, cache *Cache) {
	//输出日志
	log.Println("Received POST request:", r.URL.Path)

	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Parsed data: %v", data)
	for k, v := range data {
		cache.Set(k, v)
	}
}

// handleDelete 处理 DELETE 请求。
func handleDelete(w http.ResponseWriter, r *http.Request, cache *Cache, key string) {
	log.Println("Received Delete request:", r.URL.Path)
	if _, found := cache.Get(key); found {
		cache.Delete(key)
		respondWithJSON(w, http.StatusOK, 1)
	} else {
		respondWithJSON(w, http.StatusOK, 0)
	}
}

// respondWithJSON 发送 JSON 格式的响应。
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// extractKeyFromRequest 提取 HTTP 请求中的 key。
// 对于 POST 和 DELETE 请求，key 包含在 URL 路径中；
// 对于 POST 请求，我们假设请求体中只有一个键值对，并从中提取 key。
func extractKeyFromRequest(r *http.Request) string {
	// 对于 GET 和 DELETE 请求，key 在 URL 路径中
	if r.Method == http.MethodGet || r.Method == http.MethodDelete {
		return r.URL.Path[1:]
	}

	// 对于 POST 请求，提取请求体中的 key
	if r.Method == http.MethodPost {
		// 读取并备份请求体
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return ""
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // 重新设置请求体

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return ""
		}
		for k := range data {
			return k // 返回第一个 key
		}
	}

	return ""
}
