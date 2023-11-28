package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	pb "distributed-cache/cache"
	"google.golang.org/grpc"
)

// grpcServer 是 gRPC 服务的实现。
type grpcServer struct {
	pb.UnimplementedCacheServiceServer
	cache *Cache
}

// NewGRPCServer 初始化并返回一个新的 grpcServer 实例。
func NewGRPCServer(cache *Cache) *grpcServer {
	return &grpcServer{cache: cache}
}

// startGRPCServer 启动 gRPC 服务器。
func startGRPCServer(address string, cache *Cache) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCacheServiceServer(s, NewGRPCServer(cache))
	log.Printf("gRPC server listening on %s", address)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Get 实现了 gRPC 服务接口中的 Get 方法。
func (s *grpcServer) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	value, found := s.cache.Get(in.Key)
	return &pb.GetResponse{Value: value, Found: found}, nil
}

// Set 实现了 gRPC 服务接口中的 Set 方法。
func (s *grpcServer) Set(ctx context.Context, in *pb.SetRequest) (*pb.SetResponse, error) {
	s.cache.Set(in.Key, in.Value)
	return &pb.SetResponse{}, nil
}

// Delete 实现了 gRPC 服务接口中的 Delete 方法。
func (s *grpcServer) Delete(ctx context.Context, in *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	_, found := s.cache.Get(in.Key)
	if found {
		s.cache.Delete(in.Key)
		return &pb.DeleteResponse{DeletedCount: 1}, nil
	}
	return &pb.DeleteResponse{DeletedCount: 0}, nil
}

// forwardRequest 使用 gRPC 转发请求到其他节点。
func forwardRequest(w http.ResponseWriter, r *http.Request, targetNodeIndex int, cache *Cache) {
	targetGRPCAddress := fmt.Sprintf("localhost:%d", 1000+9527+targetNodeIndex)

	conn, err := grpc.Dial(targetGRPCAddress, grpc.WithInsecure(), grpc.WithBlock()) // 与目标节点建立连接
	if err != nil {
		http.Error(w, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewCacheServiceClient(conn)                                // 创建 gRPC 客户端
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 创建上下文
	defer cancel()                                                          // 调用结束后取消上下文

	key := extractKeyFromRequest(r)

	switch r.Method {
	case http.MethodGet:
		// 调用 Get 方法
		res, err := client.Get(ctx, &pb.GetRequest{Key: key})
		if err != nil {
			http.Error(w, "Failed to retrieve value", http.StatusInternalServerError)
			return
		}
		if res.Value != "" {
			respondWithJSON(w, http.StatusOK, map[string]string{key: res.Value})
		} else {
			http.NotFound(w, r)
		}

	case http.MethodPost:
		var data map[string]string
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			log.Printf("Error decoding JSON: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf("Parsed data: %v", data)
		_, err = client.Set(ctx, &pb.SetRequest{Key: key, Value: data[key]})
		if err != nil {
			http.Error(w, "Failed to set value", http.StatusInternalServerError)
			return
		}
		//w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		res, err := client.Delete(ctx, &pb.DeleteRequest{Key: key})
		if err != nil {
			http.Error(w, "Failed to delete key", http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, res.DeletedCount)
	}
}
