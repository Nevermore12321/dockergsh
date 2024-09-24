package graphdb

import "sort"

type pathSorter struct {
	paths []string               // 需要排序的路径
	by    func(i, j string) bool // 排序函数
}

func (s *pathSorter) Len() int {
	return len(s.paths)
}

func (s *pathSorter) Swap(i, j int) {
	s.paths[i], s.paths[j] = s.paths[j], s.paths[i]
}

func (s *pathSorter) Less(i, j int) bool {
	return s.by(s.paths[i], s.paths[j])
}

func sortByDepth(paths []string) {
	s := &pathSorter{
		paths: paths,
		by: func(i, j string) bool {
			return pathDepth(i) > pathDepth(j)
		},
	}

	sort.Sort(s)
}
