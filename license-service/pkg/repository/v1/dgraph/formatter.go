package dgraph

import "strings"

func addTab(str string, num int) string {
	tab := "\t"
	for j := 0; j < num; j++ {
		str = tab + str
	}
	return str

}

// func check(e error) {
// 	if e != nil {
// 		panic(e)
// 	}
// }

func formatter(uf string) string {
	balanceParenthesisCount := 0
	fileContent := uf

	fileContent = strings.TrimSpace(fileContent)

	splittedArray := strings.Split(fileContent, "\n")

	for i := 0; i < len(splittedArray); i++ {
		currentString := splittedArray[i]
		trimmedString := strings.TrimSpace(currentString)
		if strings.HasSuffix(trimmedString, "{") { // nolint: gocritic
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
