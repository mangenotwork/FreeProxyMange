package pool

import (
	"context"
	"strings"
	"time"

	gt "github.com/mangenotwork/gathertool"
)

func CheckTask(ctx context.Context) {

	go func(ctx context.Context) {
		for {

			select {
			case <-ctx.Done():
				gt.Info("启动池子维护任务 收到退出信号，开始停止服务...")
				// 这里可以写收尾逻辑：关闭数据库连接、清理临时文件、保存状态等
				time.Sleep(1 * time.Second) // 模拟收尾耗时
				gt.Info("池子检查已安全停止")
				return

			default:
				gt.Info("启动池子维护任务....")

				time.Sleep(4 * time.Second)
				for _, p := range AllDBPath() {
					ips, err := BadgerGetAllKeys(p)
					if err != nil {
						gt.Error(err)
					}
					gt.Info("ips = ", ips)
					for _, v := range ips {
						gt.Info("准备检查 ", v)
						time.Sleep(1 * time.Second)
						ip, ok, err := BadgerReadStruct(p, v)
						if err != nil {
							gt.Error(err)
						}
						if ok {
							if ip.FailNum > 4 {
								gt.Info(ip.IP, "验证4次都失败了,执行删除")
								BadgerDeleteStruct(p, ip.IP)
							}
							cms, err := CheckMs(ip.IP)
							if err != nil {
								ip.FailNum++
							} else {
								ip.CheckNum++
								ip.LastCheckTime = time.Now().GoString()
								ip.LastCheckMs = cms
							}
							BadgerUpsertStruct(p, ip.IP, ip)
						}
					}
				}
			}

		}
	}(ctx)

}

func Check(ip string) string {

	if !strings.HasPrefix(ip, "http://") && !strings.HasPrefix(ip, "https://") {
		ip = "http://" + ip
	}

	// https://myip.ipip.net

	ctx, err := gt.Get("https://www.doubao.com/chat/", gt.ProxyUrl(ip), gt.ReqTimeOut(30))
	if err != nil {
		gt.Error(err)
		return err.Error()
	}
	gt.Info(ctx.Ms)
	gt.Info(ctx.RespBodyString())
	return ctx.RespBodyString() + "  ms:" + ctx.Ms.String()
}

func CheckMs(ip string) (string, error) {
	if !strings.HasPrefix(ip, "http://") && !strings.HasPrefix(ip, "https://") {
		ip = "http://" + ip
	}

	// https://myip.ipip.net
	gt.Info("Check ", ip)

	ctx, err := gt.Get("https://www.doubao.com/chat/", gt.ProxyUrl(ip), gt.ReqTimeOut(10))
	if err != nil {
		gt.Error(err)
		return "", err
	}
	gt.Info(ctx.Ms)
	//gt.Info(ctx.RespBodyString())
	return ctx.Ms.String(), nil
}
