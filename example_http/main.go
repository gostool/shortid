package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gostool/shortid"
)

func main() {
	// 命令行参数
	addr := flag.String("addr", ":8080", "HTTP server address")
	redisAddr := flag.String("redis", "localhost:6379", "Redis server address")
	businessType := flag.String("business", "order", "Business type (order, user, etc.)")
	flag.Parse()

	// 解析业务类型
	var bt shortid.BusinessType
	switch *businessType {
	case "order":
		bt = shortid.BusinessOrder
	case "user":
		bt = shortid.BusinessUser
	default:
		bt = shortid.BusinessOrder
	}

	// 创建HTTP服务器
	server, err := shortid.NewHTTPServer(*addr, *redisAddr, bt)
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	// 启动服务器（在goroutine中）
	go func() {
		log.Printf("Starting HTTP server on %s", *addr)
		log.Printf("Redis address: %s", *redisAddr)
		log.Printf("Business type: %s", *businessType)
		log.Printf("API endpoints:")
		log.Printf("  GET/POST /nextid - Generate NextID")
		log.Printf("  GET /health - Health check")
		if err := server.Start(); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

