package pool

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	gt "github.com/mangenotwork/gathertool"
)

// 必须定义空日志器（放在函数外）
type nullLogger struct{}

func (l *nullLogger) Errorf(format string, v ...interface{})   {}
func (l *nullLogger) Warningf(format string, v ...interface{}) {}
func (l *nullLogger) Infof(format string, v ...interface{})    {}
func (l *nullLogger) Debugf(format string, v ...interface{})   {}

// ======================================
// BadgerUpsertStruct：插入或更新数据（不存在则新增，存在则修改）
// dbPath: 数据库路径（入参）
// key: 键（字符串）
// value: 结构体类型的值
// 返回：错误信息
// ======================================
func BadgerUpsertStruct(dbPath string, key string, value *ProxyIP) error {
	// 严格参数校验（复用原有逻辑）
	if dbPath == "" {
		return errors.New("数据库路径不能为空")
	}
	if key == "" {
		return errors.New("键不能为空")
	}
	if value == nil {
		return errors.New("结构体值不能为空")
	}

	// 结构体序列化为 JSON 字节数组（仅一次序列化，减少冗余）
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("结构体序列化失败: %w", err)
	}

	// 统一配置 BadgerDB（保留原有高性能配置）
	opts := badger.DefaultOptions(dbPath).
		WithMemTableSize(16 << 20).
		WithValueLogFileSize(64 << 20).
		WithSyncWrites(false).
		WithCompression(options.ZSTD).WithLogger(&nullLogger{})

	// 打开数据库
	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	defer db.Close()

	// 核心逻辑：无需检查键是否存在，直接 Set（BadgerDB 的 Set 操作天然支持覆盖）
	var isCreate bool // 标记是新增还是更新
	err = db.Update(func(txn *badger.Txn) error {
		// 先查询键是否存在（仅用于日志标记，不影响核心逻辑）
		_, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			isCreate = true // 键不存在，本次是新增
		} else if err != nil {
			return fmt.Errorf("查询键是否存在失败: %w", err)
		} // 否则 isCreate 为 false，本次是更新

		// 写入/覆盖数据（核心：Set 操作既可以新增也可以更新）
		if err := txn.Set([]byte(key), valueBytes); err != nil {
			return fmt.Errorf("写入/更新数据失败: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Upsert 数据失败: %w", err)
	}

	// 区分日志输出（保持原有日志格式）
	if isCreate {
		log.Printf("[新增成功] 路径：%s | 键：%s | 值：%+v", dbPath, key, value)
	} else {
		log.Printf("[更新成功] 路径：%s | 键：%s | 新值：%+v", dbPath, key, value)
	}
	return nil
}

// 2. 查询数据（Read）
// dbPath: 数据库路径（入参）
// key: 要查询的键
// 返回：结构体值、是否存在、错误信息
// ======================================
func BadgerReadStruct(dbPath string, key string) (*ProxyIP, bool, error) {
	// 参数校验
	if dbPath == "" {
		return nil, false, errors.New("数据库路径不能为空")
	}
	if key == "" {
		return nil, false, errors.New("键不能为空")
	}

	// 打开数据库（只读模式）
	opts := badger.DefaultOptions(dbPath).WithLogger(&nullLogger{})
	db, err := badger.Open(opts)
	if err != nil {
		return nil, false, fmt.Errorf("打开数据库失败: %w", err)
	}
	defer db.Close()

	var valueBytes []byte
	// 读事务查询数据
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // 键不存在，后续判断
			}
			return fmt.Errorf("查询数据失败: %w", err)
		}

		// 拷贝值（BadgerDB 要求拷贝，避免引用失效）
		valueBytes, err = item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("拷贝值失败: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, false, fmt.Errorf("查询结构体数据失败: %w", err)
	}

	// 键不存在
	if valueBytes == nil {
		return nil, false, nil
	}

	// 反序列化为结构体
	var result ProxyIP
	err = json.Unmarshal(valueBytes, &result)
	if err != nil {
		return nil, false, fmt.Errorf("结构体反序列化失败: %w", err)
	}

	return &result, true, nil
}

// ======================================
// 4. 删除数据（Delete）
// （与普通 KV 删除逻辑一致，无需修改，复用即可）
// ======================================
func BadgerDeleteStruct(dbPath string, key string) error {
	if dbPath == "" {
		return errors.New("数据库路径不能为空")
	}
	if key == "" {
		return errors.New("键不能为空")
	}

	opts := badger.DefaultOptions(dbPath).WithLogger(&nullLogger{})
	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	defer db.Close()

	err = db.Update(func(txn *badger.Txn) error {
		// 检查键是否存在
		_, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("键【%s】不存在，无需删除", key)
			}
			return fmt.Errorf("查询键是否存在失败: %w", err)
		}

		// 删除数据
		if err := txn.Delete([]byte(key)); err != nil {
			return fmt.Errorf("删除数据失败: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("删除结构体数据失败: %w", err)
	}
	log.Printf("[删除成功] 路径：%s | 键：%s", dbPath, key)
	return nil
}

// ======================================
// 新增：极速查询所有 Key 函数
// dbPath: 数据库路径（入参）
// 返回：所有 Key 的字符串切片、错误信息
// 核心优化：仅遍历 Key，不加载 Value，速度极快
// ======================================
func BadgerGetAllKeys(dbPath string) ([]string, error) {
	// 参数校验
	if dbPath == "" {
		return nil, errors.New("数据库路径不能为空")
	}

	// 配置：只读模式 + 禁用不必要的缓存（极致轻量）
	opts := badger.DefaultOptions(dbPath).
		WithSyncWrites(false).                                  // 关闭同步写（只读场景无意义）
		WithCompression(options.None).WithLogger(&nullLogger{}) // 遍历 Key 无需解压 Value，关闭压缩加速

	// 打开数据库
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	defer db.Close()

	// 预分配切片（减少内存分配次数，提升速度）
	keys := make([]string, 0, 1024) // 初始容量 1024，可根据实际数据量调整

	// 开启只读迭代器（核心：仅遍历 Key，不加载 Value）
	err = db.View(func(txn *badger.Txn) error {
		// 创建迭代器：默认遍历所有 Key，按字典序排列
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close() // 确保迭代器关闭

		// 极速遍历所有 Key
		for iter.Rewind(); iter.Valid(); iter.Next() {
			// 仅获取 Key（不触碰 Value，这是速度快的核心）
			key := iter.Item().Key()
			// 拷贝 Key 到切片（避免引用失效）
			keys = append(keys, string(key))
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("查询所有 Key 失败: %w", err)
	}

	gt.Info("[查询所有 Key 成功] 路径：%s | 共查询到 %d 个 Key", dbPath, len(keys))
	return keys, nil
}
