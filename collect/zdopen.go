package collect

import (
	"FreeProxyMange/pool"
	"context"
	"time"

	gt "github.com/mangenotwork/gathertool"
)

func ZdopenTask(ctx context.Context) {
	go func(ctx context.Context) {
		for {

			select {
			case <-ctx.Done():
				gt.Info("启动采集任务 收到退出信号，开始停止服务...")
				// 这里可以写收尾逻辑：关闭数据库连接、清理临时文件、保存状态等
				time.Sleep(1 * time.Second) // 模拟收尾耗时
				gt.Info("任务安全停止")
				return
			default:
				// 模拟正常业务运行
				gt.Info("启动采集任务....")
				time.Sleep(14 * time.Second)
				ctx, err := gt.Get("http://www.zdopen.com/ShortProxy/GetIP/?api=202601112328085632&akey=3b61ce2c13043ee9&count=1&timespan=3&type=1")
				if err != nil {
					gt.Error("提取ip失败:", err)
				} else {
					ip := ctx.RespBodyString()
					gt.Info("提取到的ip = ", ip)
					proxyIP := &pool.ProxyIP{
						IP: ip,
					}
					err = proxyIP.Add()
					if err != nil {
						gt.Error("存储ip失败，err = ", err)
					}
				}
			}
		}
	}(ctx)

}
