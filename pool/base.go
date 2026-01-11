package pool

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"sync"
	"time"

	gt "github.com/mangenotwork/gathertool"
)

var tableCount = 32

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

type ProxyIP struct {
	IP            string `json:"ip"`
	Type          string `json:"type"` // http https socket5
	Site          string `json:"site"`
	LastCheckTime string `json:"lastCheckTime"` // 最后检查时间
	CheckNum      int    `json:"checkNum"`      // 检查次数
	LastCheckMs   string `json:"lastCheckMs"`   // 最后检查IP响应时间ms
}

func (p *ProxyIP) Add() error {
	hash := fnv.New64a()
	_, err := hash.Write([]byte(p.IP))
	if err != nil {
		return err
	}
	hashValue := hash.Sum64()
	table := fmt.Sprintf("./data/%d", (hashValue % uint64(tableCount)))
	err = BadgerUpsertStruct(table, p.IP, p)
	if err != nil {
		return err
	}
	return err
}

func AllDBPath() []string {
	fList, err := GetSubdirectories("./data")
	if err != nil {
		gt.Error(err)
		return make([]string, 0)
	}
	sort.Slice(fList, func(i, j int) bool {
		return fList[i] > fList[j]
	})
	return fList
}

func GetSubdirectories(dirPath string) ([]string, error) {
	// 用于存储子目录路径
	var subDirs []string

	// 读取目录下的所有项（文件/目录）
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	// 遍历所有项，筛选出目录
	for _, entry := range entries {
		// 判断当前项是否为目录
		if entry.IsDir() {
			// 拼接完整路径（推荐使用 filepath.Join 处理路径分隔符兼容问题）
			fullPath := dirPath + "/" + entry.Name()
			subDirs = append(subDirs, fullPath)
		}
	}

	return subDirs, nil
}

// page 分页
func GetAllKey(page int) ([]string, error) {
	dbList := AllDBPath()
	gt.Info("dbList = ", dbList)
	if page > len(dbList)-1 {
		page = len(dbList) - 1
	}
	if page < 0 {
		return make([]string, 0), nil
	}
	path := dbList[page]
	gt.Info("path = ", path)
	return BadgerGetAllKeys(path)
}
