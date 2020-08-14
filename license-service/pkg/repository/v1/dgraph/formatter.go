// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import "strings"

func addTab(str string, num int) string {
	tab := "\t"
	for j := 0; j < num; j++ {
		str = tab + str
	}
	return str

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func formatter(uf string) string {
	balanceParenthesisCount := 0
	fileContent := uf

	fileContent = strings.TrimSpace(fileContent)

	splittedArray := strings.Split(fileContent, "\n")

	for i := 0; i < len(splittedArray); i++ {
		currentString := splittedArray[i]
		trimmedString := strings.TrimSpace(currentString)
		if strings.HasSuffix(trimmedString, "{") {
			trimmedString = addTab(trimmedString, balanceParenthesisCount)
			balanceParenthesisCount++
		} else if strings.HasPrefix(trimmedString, "}") {
			balanceParenthesisCount--
			trimmedString = addTab(trimmedString, balanceParenthesisCount)
		} else {
			trimmedString = addTab(trimmedString, balanceParenthesisCount)
		}
		splittedArray[i] = trimmedString
	}

	return strings.Join(splittedArray, "\n")

}
