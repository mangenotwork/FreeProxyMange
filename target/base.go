package target

import (
	"FreeProxyMange/pool"
	"context"
	"sync"
	"time"

	gt "github.com/mangenotwork/gathertool"
)

func Run(ctx context.Context, wg *sync.WaitGroup) {
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		NotUsedTask()
		UsedTask()
		NotUsedTask2()
	}(ctx, wg)
}

var NotUsed sync.Map
var Used sync.Map

func UsedTask() {
	go func() {
		for {
			time.Sleep(4 * time.Second)
			Used.Range(func(key, value any) bool {
				time.Sleep(1 * time.Second)
				now := time.Now().Unix()
				if value.(int64) < now-60*2 {
					NotUsed.Store(key, now)
					Used.Delete(key)
				}
				return true
			})

		}
	}()
}

func NotUsedTask() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, p := range pool.AllDBPath() {
				ips, err := pool.BadgerGetAllKeys(p)
				if err != nil {
					gt.Error(err)
				}
				gt.Info("ips = ", ips)
				for _, v := range ips {
					time.Sleep(1 * time.Second)
					go func() {
						gt.Info("从池子里找可用ip ", v)
						_, err := pool.CheckMs(v)
						if err == nil {
							now := time.Now().Unix()
							NotUsed.Store(v, now)
						}
					}()

				}
			}
		}
	}()
}

func NotUsedTask2() {
	go func() {
		for {
			time.Sleep(4 * time.Second)
			NotUsed.Range(func(key, value any) bool {
				item := 0
			R:
				if item > 4 {
					NotUsed.Delete(key.(string))
					return true
				}
				time.Sleep(1 * time.Second)
				_, err := pool.CheckMs(key.(string))
				if err != nil {
					item++
					goto R
				}

				return true
			})
		}
	}()
}

func UseIP() string {
	ip := ""
	NotUsed.Range(func(key, value any) bool {
		ip = key.(string)
		now := time.Now().Unix()
		Used.Store(key.(string), now)
		NotUsed.Delete(key.(string))
		return false
	})
	return ip
}

func ShowUse() []string {
	rse := make([]string, 0)
	Used.Range(func(key, value any) bool {
		rse = append(rse, key.(string))
		return true
	})
	return rse
}

func ShowNotUse() []string {
	rse := make([]string, 0)
	NotUsed.Range(func(key, value any) bool {
		rse = append(rse, key.(string))
		return true
	})
	return rse
}
