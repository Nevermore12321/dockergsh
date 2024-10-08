package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxLoopCounter = 100 // 最大寻找 100 层

// FollowSymlinkInScope 递归解析符号链接，并确保解析的路径不会超出 root 根目录范围
// 即在 root 目录范围内，寻找link软链接的源路径
func FollowSymlinkInScope(link, root string) (string, error) {
	// root 绝对路径
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	// link 绝对路径
	linkAbs, err := filepath.Abs(link)
	if err != nil {
		return "", err
	}

	if linkAbs == rootAbs {
		return link, nil
	}

	// 如果 link 绝对路径不在 root 范围内
	if !strings.HasPrefix(linkAbs, rootAbs) {
		return "", fmt.Errorf("symlink %s is not within %s", linkAbs, rootAbs)
	}

	// 将 link 的路径拆分为多个部分，逐层遍历路径，并在每次循环中将路径组件加入到 prev，形成新的路径
	prev := "/"
	for _, p := range strings.Split(linkAbs, "/") {
		// 在每次循环中将路径组件加入到 prev，形成新的路径
		prev = filepath.Join(prev, p)
		prev = filepath.Clean(prev) // 去掉路径中的 .. . 等相对路径

		// 循环寻找软链接的源地址，判断是否在 root 下
		loopCounter := 0
		for {
			loopCounter++

			if loopCounter > maxLoopCounter { // 递归层数超过 100 层
				return "", fmt.Errorf("loopCounter reached MAX: %v", loopCounter)
			}

			if !strings.HasPrefix(prev, root) {
				break
			}

			// 判断 prev 路径是否存在，如果不存在直接跳出循环
			stat, err := os.Lstat(prev)
			if err != nil {
				if os.IsNotExist(err) {
					break
				}
				return "", err
			}

			// 如果 prev 路径是一个软链接
			if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
				dest, err := os.Readlink(prev)
				if err != nil {
					return "", err
				}

				if filepath.IsAbs(dest) {
					prev = filepath.Join(root, dest)
				} else {
					prev, _ = filepath.Abs(prev)

					if prev = filepath.Clean(filepath.Join(filepath.Dir(prev), dest)); len(prev) < len(root) {
						prev = filepath.Join(root, filepath.Base(prev))
					}

				}
			} else { // 不是软链接
				break
			}
		}
	}
	return prev, nil
}
