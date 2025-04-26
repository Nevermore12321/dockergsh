package version

import (
	"strconv"
	"strings"
)

type Version string

// 相等返回 0，大于返回 1，小于返回 -1
func (v Version) compareTo(other Version) int {
	var (
		meTab    = strings.Split(string(v), ".")
		otherTab = strings.Split(string(other), ".")
	)
	maxLen := len(meTab)
	if len(otherTab) > maxLen {
		maxLen = len(otherTab)
	}

	for i := 0; i < maxLen; i++ {
		var meInt, otherInt int

		if len(meTab) > i {
			meInt, _ = strconv.Atoi(meTab[i])
		}
		if len(otherTab) > i {
			otherInt, _ = strconv.Atoi(otherTab[i])
		}

		if meInt > otherInt {
			return 1
		}
		if otherInt > meInt {
			return -1
		}
	}
	return 0
}

func (v Version) Equal(other Version) bool {
	return v.compareTo(other) == 0
}

func (v Version) LessThan(other Version) bool {
	return v.compareTo(other) < 0
}

func (v Version) LessThanOrEqual(other Version) bool {
	return v.compareTo(other) <= 0
}

func (v Version) GreaterThan(other Version) bool {
	return v.compareTo(other) > 0
}

func (v Version) GreaterThanOrEqual(other Version) bool {
	return v.compareTo(other) > 0
}
