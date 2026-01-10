package main

import (
	"FreeProxyMange/collect"
	"FreeProxyMange/pool"
	"FreeProxyMange/serve"
	"FreeProxyMange/target"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	gt "github.com/mangenotwork/gathertool"
)

func main() {

	gt.Info("free proxy mange")
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	collect.Run(ctx, &wg)
	wg.Add(1)
	pool.Run(ctx, &wg)
	wg.Add(1)
	target.Run(ctx, &wg)
	wg.Add(1)
	serve.Run(ctx, &wg)

	sigChan := make(chan os.Signal, 1)
	// 监听的信号：SIGINT（Ctrl+C）、SIGTERM（kill 命令）
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞主协程，等待信号
	gt.Info("程序已启动，按 Ctrl+C 退出...")
	sig := <-sigChan
	gt.Infof("接收到信号：%v，开始优雅退出...\n", sig)

	// 触发退出逻辑
	cancel()
	wg.Wait()
	gt.Info("程序已完全退出")

}
