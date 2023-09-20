package kernel

import (
	"bytes"
	"errors"
	"fmt"
)

// 获取到运行 dockergsh 所在服务器的 kernel 版本
type KernelVersionInfo struct {
	Kernel int    // kernel 版本号
	Major  int    // kernel 的 major 版本号
	Minor  int    // kernel 的 minor 版本号
	Flavor string // Kernel 版本信息
}

// 打印格式： Kernel.Major.Minor Flavor
func (kvf *KernelVersionInfo) String() string {
	return fmt.Sprintf("%d.%d.%d %s", kvf.Kernel, kvf.Major, kvf.Minor, kvf.Flavor)
}

func GetKernelVersion() (*KernelVersionInfo, error) {
	var err error

	// 获取 uname 返回结果
	utsname, err := uname()
	if err != nil {
		return nil, err
	}

	// 将 uname 返回结果的 syscall.Utsname 结构体中 Release 信息提取
	release := make([]byte, len(utsname.Release))
	i := 0
	for _, c := range utsname.Release {
		release[i] = byte(c)
		i += 1
	}

	// 从 release 中删除 \x00 以便 Atoi 正确解析，
	release = release[:bytes.IndexByte(release, 0)]

	// 解析 release 信息到 KernelVersionInfo 结构体
	return ParseRelease(string(release))
}

// ParseRelease 解析 uname 返回结果的 Release 信息到 KernelVersionInfo 结构体
func ParseRelease(release string) (*KernelVersionInfo, error) {
	var (
		parsed               int
		kernel, major, minor int
		flavor, partial      string
	)

	// 从 release 信息中读取出 kernel major 以及系统信息 partial
	parsed, _ = fmt.Sscanf(release, "%d.%d%s", &kernel, &major, &partial)
	if parsed < 2 {
		// 没有完全解析
		return nil, errors.New("Can't parse kernel version " + release)
	}

	// 再从 系统信息 partial 中解析出 minor 和 flavor
	parsed, _ = fmt.Sscanf(partial, ".%d%s", &minor, &flavor)
	if parsed < 1 {
		flavor = partial
	}

	return &KernelVersionInfo{
		Kernel: kernel,
		Major:  major,
		Minor:  minor,
		Flavor: flavor,
	}, nil
}

// CompareKernelVersion 获取到系统版本信息后，对版本号进行比较
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func CompareKernelVersion(a, b *KernelVersionInfo) int {
	// 先比较kernel主版本
	if a.Kernel < b.Kernel {
		return -1
	} else if a.Kernel > b.Kernel {
		return 1
	}

	// 如果 kernel 主版本相同，比较 kernel 的 major 版本
	if a.Minor < b.Major {
		return -1
	} else if a.Major > b.Major {
		return 1
	}

	// 如果 kernel 的主版本，major 版本都相同，比较 minor 版本
	if a.Minor < b.Minor {
		return -1
	} else if a.Minor > a.Minor {
		return 1
	}

	// 都相同
	return 0
}
