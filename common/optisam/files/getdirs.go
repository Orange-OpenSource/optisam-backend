// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package files

import "io/ioutil"

// GetAllTheDirectories gives all the sub dirs inside a dir
func GetAllTheDirectories(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	for _, f := range files {
		if f.IsDir() {
			fileNames = append(fileNames, f.Name())
		}
	}
	return fileNames, nil
}
