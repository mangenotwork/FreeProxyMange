package collect

import (
	"context"
	"sync"
)

func Run(ctx context.Context, wg *sync.WaitGroup) {
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		ZdopenTask(ctx)
	}(ctx, wg)
}
