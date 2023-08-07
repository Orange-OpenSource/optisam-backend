package helper

import (
	"regexp"
	"sort"
)

// RemoveElements removes all elements from originalSlice
// that are in removeElementSlice
func RemoveElements(originalSlice []string, removeElementSlice []string) []string {

	for _, elem := range removeElementSlice {
		// search is linear but can be improved
		for i := 0; i < len(originalSlice); i++ {
			if originalSlice[i] == elem {
				originalSlice = append(originalSlice[:i], originalSlice[i+1:]...)
				i--
			}
		}
	}
	return originalSlice
}

func AppendElementsIfNotExists(originalSlice []string, addElementSlice []string) []string {
	for _, addelem := range addElementSlice {
		found := false
		// search is linear but can be improved
		for _, elem := range originalSlice {
			if elem == addelem {
				found = true
				break
			}
		}
		if !found {
			originalSlice = append(originalSlice, addelem)
		}
	}
	return originalSlice
}

// RegexContains Check if passed string matches any of the regex in the slice
func RegexContains(reslice []string, val string) bool {
	for _, item := range reslice {
		re := regexp.MustCompile(item)
		if re.MatchString(val) {
			return true
		}
	}
	return false
}

func Contains(slice []string, vals ...string) bool {
outer:
	for _, val := range vals {
		for _, item := range slice {
			if item == val {
				continue outer
			}
		}
		return false
	}
	return true
}

func ContainsInts(slice []int32, vals ...int32) bool {
outer:
	for _, val := range vals {
		for _, item := range slice {
			if item == val {
				continue outer
			}
		}
		return false
	}
	return true
}

func MakeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func CompareSlices(a1, a2 []string) bool {
	sort.Strings(a1)
	sort.Strings(a2)
	if len(a1) == len(a2) {
		for i, v := range a1 {
			if v != a2[i] {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func ExactCompareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
