package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

/* TagStore主要是用于管理存储镜像的仓库列表（repository list）。 */
/*
TagStore 与 Graph：
- TagStore 使用 Graph 提供的功能来解析和管理镜像层次结构。标签（tags）通过 TagStore 映射到具体的镜像 ID，而这些镜像 ID 代表的镜像层次结构由 Graph 管理。
- 例如，当用户拉取一个镜像（例如 nginx:latest）时，TagStore 会将 nginx:latest 映射到具体的镜像 ID，然后 Graph 使用这个镜像 ID 来获取相应的镜像层次结构。
*/

type TagStore struct {
	path         string                // TagStore中记录镜像仓库的文件所在路径，如aufs类型的TagStore path的值为"/var/lib/docker/repositories-aufs"。
	graph        *Graph                // Graph实例对象
	Repositories map[string]Repository // 记录镜像仓库的映射数据结构

	sync.RWMutex                          // TagStore的互斥锁
	pullingPool  map[string]chan struct{} // 记录有哪些镜像正在被下载，若某一个镜像正在被下载，则驳回其他Docker Client发起下载该镜像的请求
	pushingPool  map[string]chan struct{} // 记录有哪些镜像正在被上传，若某一个镜像正在被上传，则驳回其他Docker Client发起上传该镜像的请求。
}

type Repository map[string]string

// NewTagStore 创建 TagStore 实例
func NewTagStore(path string, graph *Graph) (*TagStore, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	tagStore := TagStore{
		path:         abspath,
		graph:        graph,
		Repositories: make(map[string]Repository),
		pullingPool:  make(map[string]chan struct{}),
		pushingPool:  make(map[string]chan struct{}),
	}

	// 如果存在 json 文件则加载，否则创建
	if err := tagStore.reload(); os.IsNotExist(err) { // json 文件不存在
		if err := tagStore.save(); err != nil {
			return nil, err
		}
	} else if err != nil { // 其他加载错误
		return nil, err
	}
}

// 将 TagStore.path 指定的文件内容结构化到 store 实例中
func (store *TagStore) reload() error {
	// 读取 json 文件内容
	jsonData, err := os.ReadFile(store.path)
	if err != nil {
		return err
	}
	// 将 json 字符川 转成 store 结构体
	err = json.Unmarshal(jsonData, store)
	if err != nil {
		return err
	}
	return nil
}

// 将 TagStore 实例 store 中的内容，写入到 json 文件
func (store *TagStore) save() error {
	// 将 store 示例内容转成 json 字符串
	jsonStr, err := json.Marshal(store)
	if err != nil {
		return err
	}

	// 写入 json 文件
	if err := os.WriteFile(store.path, jsonStr, 0600); err != nil {
		return err
	}
	return nil
}
