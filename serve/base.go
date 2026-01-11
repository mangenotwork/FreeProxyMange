package serve

import (
	"FreeProxyMange/pool"
	"FreeProxyMange/target"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	gt "github.com/mangenotwork/gathertool"
)

// 模拟业务逻辑：比如一个持续运行的服务
func Run(ctx context.Context, wg *sync.WaitGroup) {

	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		mux := http.NewServeMux()
		mux.HandleFunc("/all", allHandler)
		mux.HandleFunc("/add", addHandler)
		mux.HandleFunc("/check", checkHandler)
		mux.HandleFunc("/get", getHandler)
		mux.HandleFunc("/useList", useShowHandler)
		mux.HandleFunc("/notuseList", notuseShowHandler)

		// 启动 HTTP 服务，监听 8080 端口
		httpServer := &http.Server{
			Addr:    ":8082",
			Handler: ResponseHeaderMiddleware(mux), // 中间件包裹所有路由
		}

		// 2. 启动 HTTP 服务的 goroutine（非阻塞）
		serverErrChan := make(chan error, 1)
		go func() {
			gt.Info("HTTP 服务启动，监听地址：:8080")
			// ListenAndServe 是阻塞调用，放入 goroutine 避免阻塞主逻辑
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverErrChan <- err
			}
			close(serverErrChan)
		}()

		// 3. 监听退出信号（ctx.Done() 或 HTTP 服务错误）
		select {
		case <-ctx.Done():
			gt.Info("业务逻辑收到退出信号，开始优雅关停 HTTP 服务...")

			// 创建超时上下文，5秒内强制关停
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 优雅关闭 HTTP 服务（关键：停止接收新请求，等待已有请求处理完成）
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				gt.Info("HTTP 服务优雅关停失败，强制退出")
			} else {
				gt.Info("HTTP 服务已优雅关停")
			}

			// 模拟其他收尾逻辑（如关闭数据库、清理资源）
			time.Sleep(1 * time.Second)
			gt.Info("所有服务已安全停止")

		case err := <-serverErrChan:
			// HTTP 服务启动失败
			gt.Error("HTTP 服务启动失败，退出程序: ", err)
		}

	}(ctx, wg)

}

type Response struct {
	Code    int         `json:"code"`    // 状态码 200=成功
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 业务数据
}

func allHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 仅允许 GET 方法
	if r.Method != http.MethodGet {
		// 设置响应头和状态码
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		// 返回 JSON 错误响应
		json.NewEncoder(w).Encode(Response{
			Code:    405,
			Message: "仅支持 GET 方法",
			Data:    nil,
		})
		return
	}

	pageStr := r.URL.Query().Get("page")
	gt.Info("pageStr = ", pageStr)

	// 2. 模拟业务数据（你可以替换为实际逻辑，比如从 BadgerDB 查询数据）
	// 示例：返回一些测试数据
	testData, err := pool.GetAllKey(gt.Any2Int(pageStr))
	if err != nil {
		_ = json.NewEncoder(w).Encode(Response{
			Code:    200,
			Message: "获取数据失败",
			Data:    err.Error(),
		})
		return
	}

	// 4. 返回 JSON 响应
	_ = json.NewEncoder(w).Encode(Response{
		Code:    200,
		Message: "获取数据成功",
		Data:    testData,
	})
	return
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	ipStr := r.URL.Query().Get("ip")

	if ipStr == "" {
		_ = json.NewEncoder(w).Encode(Response{
			Code:    200,
			Message: "ip不能为空",
			Data:    "",
		})
	}

	ipData := &pool.ProxyIP{
		IP: ipStr,
	}
	err := ipData.Add()
	if err != nil {
		_ = json.NewEncoder(w).Encode(Response{
			Code:    200,
			Message: "获取数据失败",
			Data:    err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(Response{
		Code:    200,
		Message: "添加成功",
		Data:    "",
	})
	return
}

func checkHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 仅允许 GET 方法
	if r.Method != http.MethodGet {
		// 设置响应头和状态码
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		// 返回 JSON 错误响应
		json.NewEncoder(w).Encode(Response{
			Code:    405,
			Message: "仅支持 GET 方法",
			Data:    nil,
		})
		return
	}

	ipStr := r.URL.Query().Get("ip")

	if ipStr == "" {
		_ = json.NewEncoder(w).Encode(Response{
			Code:    200,
			Message: "ip不能为空",
			Data:    "",
		})
	}

	res := pool.Check(ipStr)

	// 4. 返回 JSON 响应
	_ = json.NewEncoder(w).Encode(Response{
		Code:    200,
		Message: "",
		Data:    res,
	})

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	ip := target.UseIP()
	// 4. 返回 JSON 响应
	_ = json.NewEncoder(w).Encode(Response{
		Code:    200,
		Message: "",
		Data:    ip,
	})

}

func useShowHandler(w http.ResponseWriter, r *http.Request) {
	ip := target.ShowUse()
	// 4. 返回 JSON 响应
	_ = json.NewEncoder(w).Encode(Response{
		Code:    200,
		Message: "",
		Data:    ip,
	})

}

func notuseShowHandler(w http.ResponseWriter, r *http.Request) {
	ip := target.ShowNotUse()
	// 4. 返回 JSON 响应
	_ = json.NewEncoder(w).Encode(Response{
		Code:    200,
		Message: "",
		Data:    ip,
	})

}

// ========== 核心：通用响应头中间件 ==========
// ResponseHeaderMiddleware 中间件：设置通用响应头（JSON + 跨域）
// next: 下一个处理器（被包装的路由函数）
func ResponseHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 设置通用响应头（所有路由都会生效）
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*") // 开发环境允许所有跨域
		// 可选：支持更多跨域请求头（如 POST/PUT 需要）
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// 2. 处理 OPTIONS 预检请求（跨域必备）
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 3. 调用下一个处理器（执行实际的路由逻辑）
		next.ServeHTTP(w, r)
	})
}
