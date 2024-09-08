package daemongsh

import (
	"encoding/json"
	"github.com/Nevermore12321/dockergsh/pkg/symlink"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// todo add elements
type Container struct {
	sync.Mutex           // 修改属性时，需要上锁
	Root       string    // 当前容器示例的根目录，即 /var/lib/dockergsh/containers/[CONTAINER_ID]
	Created    time.Time // 容器的创建时间

	ID     string // 容器 ID
	State  *State // 容器状态
	Driver string // 容器使用的镜像 graph driver 类型
}

// LoadFromDisk 从 /var/lib/dockergsh/containers/[CONTAINER_ID] 目录下加载已存在的容器
func (container *Container) LoadFromDisk() error {
	// 获取对应容器的 config.json 文件地址
	configPath, err := container.jsonPath()
	if err != nil {
		return err
	}

	// 读取 config.json 文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// 加载容器结构体
	// 容器的 docker.PortMapping 结构会破坏结构化 config.json 内容，跳过
	if err := json.Unmarshal(data, container); err != nil && !strings.Contains(err.Error(), "docker.PortMapping") {
		return err
	}

	// todo label and hostconfig
	return nil
}

// jsonPath 读取具体容器下的 config.json 文件，保存容器的详细信息
func (container *Container) jsonPath() (string, error) {
	return container.getRootResourcePath("config.json")
}

// getRootResourcePath 从 conatiner.Root 中寻找containerID 目录原始路径，因为 containers 目录下存放的是软链接
func (container *Container) getRootResourcePath(resourceFile string) (string, error) {
	resourceCleanPath := filepath.Join("/", resourceFile)
	return symlink.FollowSymlinkInScope(filepath.Join(container.Root, resourceCleanPath), container.Root)
}
