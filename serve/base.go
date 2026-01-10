package serve

import (
	"context"
	"sync"
	"time"

	gt "github.com/mangenotwork/gathertool"
)

// 模拟业务逻辑：比如一个持续运行的服务
func Run(ctx context.Context, wg *sync.WaitGroup) {
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			gt.Info("业务逻辑收到退出信号，开始停止服务...")
			// 这里可以写收尾逻辑：关闭数据库连接、清理临时文件、保存状态等
			time.Sleep(1 * time.Second) // 模拟收尾耗时
			gt.Info("服务已安全停止")
			return
		default:
			// 模拟正常业务运行
			gt.Info("服务正在运行中（按 Ctrl+C 退出）")
			time.Sleep(500 * time.Millisecond)
		}
	}(ctx, wg)

}
