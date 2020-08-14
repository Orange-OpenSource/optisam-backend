// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package addcolumn

import (
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/config"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	//CmdAddDate informs about the command
	CmdAddColumn *config.Command
)

func init() {
	CmdAddColumn = &config.Command{
		Cmd: &cobra.Command{
			Use:   "addcolumn",
			Short: "add column in the dgraph files",
			Long: `print is for printing anything back to the screen.
		For many years people have printed back to the screen.`,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := addDates(); err != nil {
					return err
				}
				return nil
			},
		},
	}
	CmdAddColumn.Cmd.Flags().String("dir", "", "dataloader addcolumn --dir uploaded")
	CmdAddColumn.Cmd.Flags().StringToString("header_add", make(map[string]string), "dataloader addcolumn --header_add updated,2019-08-28T09:58:56.0260078Z --header_add created,2019-08-28T09:58:56.0260078Z")
	CmdAddColumn.Cmd.Flags().StringSlice("header_remove", []string{}, "dataloader  addcolumn --header_remove header_name")
	//	CmdAddDate.Cmd.Flags().StringSlice("value", []string{"updated", "created"}, "dataloader adddate --value 2019-08-28T09:58:56.0260078Z --value 2019-08-28T09:58:56.0260078Z")

}

func addDates() error {
	dir := CmdAddColumn.Conf.GetString("dir")
	kv := CmdAddColumn.Conf.GetStringMapString("header_add")
	rmHeaders := CmdAddColumn.Conf.GetStringSlice("header_remove")
	col := make([]string, 0, len(kv))
	val := make([]string, 0, len(kv))
	for k, v := range kv {
		col = append(col, k)
		val = append(val, v)
	}

	//fmt.Println(kv, col, val, dir)
	scopeDirs, err := getAllScopeDirs(dir)
	files, err := getAllDirFiles(scopeDirs, ".csv")

	if err != nil {
		return err
	}

	//fmt.Println(files)

	// //fmt.Println(filepath.Dir("updated/scope1/products.csv"))
	// col := []string{"updated", "created"}
	// val := []string{"2019-08-28T09:58:56.0260078Z", "2019-08-28T09:58:56.0260078Z"}

	if err := readCsvFile(files, col, val, rmHeaders); err != nil {
		return err
	}
	return nil
}

func readCsvFile(filePaths []string, cols []string, vals []string, rmHeaders []string) error {

	newFiles := make([]string, len(filePaths))

	for j, file := range filePaths {

		f, _ := os.Open(file)
		r := csv.NewReader(f)
		r.Comma = ';'
		dir := filepath.Dir(file)
		newName := dir + "/" + "newfile_" + strconv.Itoa(j+1)
		newFiles[j] = newName
		csvOut, err := os.Create(newName)
		if err != nil {
			log.Fatal("Unable to open output")
		}
		w := csv.NewWriter(csvOut)
		w.Comma = ';'

		record, err := r.Read()
		//fmt.Println(record)
		if err == io.EOF {
			return nil
		}

		headers := make(map[string]int)
		for i := range record {
			headers[record[i]] = i
		}
		// column val
		column := make([]string, 0, len(cols))
		val := make([]string, 0, len(cols))
		replaceCols := make(map[int]string)
		rmvInd := make([]int, 0, len(rmHeaders))

		for i := range rmHeaders {
			index, ok := headers[rmHeaders[i]]
			if ok {
				rmvInd = append(rmvInd, index)
			}
		}
		sort.Ints(rmvInd)

		for i := range cols {
			index, ok := headers[cols[i]]
			if ok {
				replaceCols[index] = vals[i]
				continue
			}
			column = append(column, cols[i])
			val = append(val, vals[i])
		}

		for i, ind := range rmvInd {
			record = append(record[:ind-i], record[ind-i+1:]...)
		}

		for i := range column {
			record = append(record, column[i])
		}
		//fmt.Println(record)
		if err = w.Write(record); err != nil {
			log.Fatal(err)
		}

		for {
			record, err = r.Read()

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			for i, ind := range rmvInd {
				record = append(record[:ind-i], record[ind-i+1:]...)
			}
			for i := range val {
				record = append(record, val[i])
			}

			for k, v := range replaceCols {
				record[k] = v
			}

			if err = w.Write(record); err != nil {
				log.Fatal(err)
			}

		}
		w.Flush()
		f.Close()
		csvOut.Close()

	}
	for i, file := range filePaths {
		os.Remove(file)
		os.Rename(newFiles[i], filePaths[i])
	}
	return nil
}

func getAllFilesWithSuffixFullPath(dir, suffix string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	for _, f := range files {
		name := filepath.Base(f.Name())
		//fmt.Println(name, f.Name())
		if !f.IsDir() && strings.HasSuffix(name, suffix) {
			fileNames = append(fileNames, dir+"/"+f.Name())
		}
	}
	return fileNames, nil
}

func getAllDirFiles(dirs []string, suffix string) ([]string, error) {
	var filenames []string
	for _, dir := range dirs {
		files, err := getAllFilesWithSuffixFullPath(dir, suffix)
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, files...)

	}
	return filenames, nil
}

func getAllScopeDirs(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fileScopes []string
	for _, f := range files {
		if f.IsDir() {
			fileScopes = append(fileScopes, dir+"/"+f.Name())
		}
	}
	return fileScopes, nil
}
