package graph

import (
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/graphdriver"
	"github.com/Nevermore12321/dockergsh/pkg/truncindex"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

/*
	graph 是镜像文件管理，目录在 /var/lib/dockergsh/graph
	通过graphdriver实例，管理 graph 目录的内容
*/

type Graph struct {
	Root    string                 // graph 的工作根目录，"/var/lib/docker/graph
	idIndex *truncindex.TruncIndex // 检索字符串标识符，通过字符串前缀表示镜像 ID，全局唯一
	driver  graphdriver.Driver     // graphdriver 类型
}

// NewGraph 实例化 Graph。如果 `root`（/var/lib/dockergsh/graph） 目录不存在，则会创建它。
// root : 表示存放 graph 相关信息的根目录
// driver: Graph 示例的 Driver
func NewGraph(root string, driver graphdriver.Driver) (*Graph, error) {
	absPath, err := filepath.Abs(root) // 获取绝对路径
	if err != nil {
		return nil, err
	}

	// 创建 graph 根目录
	if err := os.MkdirAll(absPath, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}

	// Graph 实例化
	graph := &Graph{
		Root:    absPath,
		idIndex: truncindex.NewTruncIndex([]string{}),
		driver:  driver,
	}

	// 如果 graph 根目录中已经存放了很多容器/镜像 id 信息，需要加载
	if err := graph.restore(); err != nil {
		return nil, err
	}

	return graph, nil
}

// 从 Graph 根目录中加载 id 信息
func (graph *Graph) restore() error {
	// 读取 Graph 根目录，目录下的所有文件名称，都是 id 字符串
	dir, err := os.ReadDir(graph.Root)
	if err != nil {
		return err
	}

	var ids = []string{}

	for _, v := range dir {
		id := v.Name()               // 文件名称就是 id
		if graph.driver.Exists(id) { // 如果存储层存在，加载
			ids = append(ids, id)
		}
	}
	graph.idIndex = truncindex.NewTruncIndex(ids)
	log.Debugf("Restored %d elements", len(dir))
	return nil
}
