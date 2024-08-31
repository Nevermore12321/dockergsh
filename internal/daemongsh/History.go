package daemongsh

import "sort"

/*
History 是列出所有容器的便捷类型，按照时间排序
*/
type History []*Container

// Len 所有容器的个数
func (history *History) Len() int {
	return len(*history)
}

func (history *History) Add(container *Container) {
	*history = append(*history, container)
}

// Sort 容器按照时间排序
func (history *History) Sort() {
	sort.Sort(history) // Sort 需要有两个函数，Less 和 Swap，也就是如果前一个比后一个小，交换
}

// Less 容器大小比较方法，按照时间先后
func (history *History) Less(i, j int) bool {
	containers := *history
	return containers[j].Created.Before(containers[i].Created) // 后一个比前一个早，则交换
}

// Swap 容器交换的方法
func (history *History) Swap(i, j int) {
	containers := *history
	tmp := containers[i]
	containers[i] = containers[j]
	containers[j] = tmp
}
