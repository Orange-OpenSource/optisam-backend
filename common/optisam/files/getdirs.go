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
