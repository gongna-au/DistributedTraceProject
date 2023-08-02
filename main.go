package main

import (
	"context"
	"log"
	"time"

	"github.com/DistributedTraceProject/config"
	"github.com/DistributedTraceProject/trace"
)

func main() {
	// 初始化追踪配置
	traceCfg := &config.Trace{
		Type:    "jaeger",
		Address: "http://localhost:14268/api/traces",
	}

	// 初始化追踪
	if err := trace.Initialize(context.Background(), traceCfg); err != nil {
		log.Fatalf("Failed to initialize trace: %v", err)
	}

	// 模拟一个请求上下文
	ctx := &config.Context{
		Context: context.Background(),
		// 在这里设置你的请求数据...
	}

	// 创建一个 hint，其中包含一个 trace context
	// 在实际使用中，这个 trace context 通常来自于请求的头部
	hint := &config.Hint{
		Type: config.TypeTrace,
		Inputs: []config.KeyValue{
			{
				K: "traceparent",
				V: "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
			},
		},
	}

	// 使用追踪提供者提取 trace context
	if ok := trace.Extract(ctx, []*config.Hint{hint}); !ok {
		log.Println("Failed to extract trace context")
	}

	// 以下是模拟的请求处理过程
	time.Sleep(100 * time.Millisecond)
}
