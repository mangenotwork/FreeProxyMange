package pool

import (
	"context"
	"sync"
	"time"

	gt "github.com/mangenotwork/gathertool"
)

/*

池子ip可用验证:
1. https://icanhazip.com/  (全球最快)
2. https://ipinfo.io/ip
3. https://httpbin.org/ip
4. https://myip.ipip.net/    （中国最快）

*/

func Run(ctx context.Context, wg *sync.WaitGroup) {
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-ctx.Done():
			gt.Info("启动池子维护任务 收到退出信号，开始停止服务...")
			// 这里可以写收尾逻辑：关闭数据库连接、清理临时文件、保存状态等
			time.Sleep(1 * time.Second) // 模拟收尾耗时
			gt.Info("服务已安全停止")
			return

		default:
			gt.Info("启动池子维护任务....")

		}

	}(ctx, wg)
}
