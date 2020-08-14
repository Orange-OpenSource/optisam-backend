// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package helper

import "regexp"

//RemoveElements removes all elements from originalSlice
// that are in removeElementSlice
func RemoveElements(originalSlice []string, removeElementSlice []string) []string {

	for _, elem := range removeElementSlice {
		//search is linear but can be improved
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
		//search is linear but can be improved
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

//RegexContains Check if passed string matches any of the regex in the slice
func RegexContains(reslice []string, val string) bool {
	for _, item := range reslice {
		re := regexp.MustCompile(item)
		if re.MatchString(val) {
			return true
		}
	}
	return false
}

func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func MakeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}
