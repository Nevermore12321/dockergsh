package driver

// StorageDriver 定义了存储引擎（文件类型）必须实现的方法。
// 存储驱动实现自动注册机制，并且自动创建工厂接口 StorageDriverFactory
type StorageDriver interface {
	// Name 返回人类可读的存储驱动名称
	Name() string
}
