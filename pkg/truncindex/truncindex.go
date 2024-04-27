package truncindex

import (
	"errors"
	"fmt"
	"github.com/tchap/go-patricia/v2/patricia"
	"strings"
	"sync"
)

var (
	ErrNoId              = errors.New("prefix can't be empty")
	ErrIdHasSpaceIllegal = errors.New("illegal character: ' '")
)

func init() {
	// 更改每个节点长度的 patricia max，len(ID) 始终为 64
	patricia.MaxPrefixPerNode(64)
}

/*
TruncIndex 允许通过任何唯一前缀检索字符串标识符。这用于通过更方便的速记前缀检索图像和容器 ID。
*/
type TruncIndex struct {
	sync.RWMutex                     // 读写锁
	trie         *patricia.Trie      // Trie 字典树、前缀树
	ids          map[string]struct{} // 已经被添加到 Trie 树中的 id 列表（64位）
}

// NewTruncIndex 创建 TruncIndex 实例
func NewTruncIndex(ids []string) *TruncIndex {
	idx := &TruncIndex{
		trie: patricia.NewTrie(),
		ids:  make(map[string]struct{}),
	}

	// 添加 id 列表到 trie 树
	for _, id := range ids {
		_ = idx.addId(id)
	}
	return idx
}

// 向 Trie 前缀树中添加 id 元素
func (idx *TruncIndex) addId(id string) error {
	// id 字符串中不能有空字符
	if strings.Contains(id, " ") {
		return ErrIdHasSpaceIllegal
	}

	// id 字符串不能为空
	if id == "" {
		return ErrNoId
	}

	// 如果 id 已经被添加到 trie 前缀树中，报错
	if _, exits := idx.ids[id]; exits {
		return fmt.Errorf("id already exists: '%s'", id)
	}

	idx.ids[id] = struct{}{}

	// 向 trie 前缀树添加 id 元素
	if inserted := idx.trie.Insert(patricia.Prefix(id), struct{}{}); !inserted {
		return fmt.Errorf("failed to insert id: %s", id)
	}

	return nil
}

// Add 线程安全的向 Trie 前缀树中添加 id 元素
func (idx *TruncIndex) Add(id string) error {
	idx.Lock()
	defer idx.Unlock()

	if err := idx.addId(id); err != nil {
		return err
	}
	return nil
}

// Delete 线程安全的从 Trie 前缀树中删除 id 元素
func (idx *TruncIndex) Delete(id string) error {
	idx.Lock()
	defer idx.Unlock()

	// 判断 id 元素是否存在
	if _, exits := idx.ids[id]; !exits || id == "" {
		return fmt.Errorf("no such id: '%s'", id)
	}

	// 从 Trie 前缀树中删除 id 元素
	if deleted := idx.trie.Delete(patricia.Prefix(id)); !deleted {
		return fmt.Errorf("can't delete, no such id: '%s'", id)
	}

	return nil
}

// Get 线程安全的从 Trie 前缀树中读取 s 元素
func (idx *TruncIndex) Get(s string) (string, error) {
	idx.RLock()
	defer idx.RUnlock()

	var id string
	if s == "" {
		return "", ErrNoId
	}

	// callback，在 trie 找到元素，就赋值给 id
	subTreeVisitFunc := func(prefix patricia.Prefix, item patricia.Item) error {
		if id != "" { // 说明已经找到了一个与 s 相同的
			id = ""
			return fmt.Errorf("trie found two entries")
		}
		id = string(prefix)
		return nil
	}

	// 从 Trie 前缀树中寻找 s 元素，使用 subTreeVisitFunc 作 callback 处理
	if err := idx.trie.VisitSubtree(patricia.Prefix(s), subTreeVisitFunc); err != nil {
		return "", err
	}

	// 如果找到 id 元素，返回
	if id != "" {
		return id, nil
	}

	// 否则没有找到，返回错误
	return "", fmt.Errorf("no such id: %s", s)
}
