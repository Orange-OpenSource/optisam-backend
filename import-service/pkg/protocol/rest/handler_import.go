// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package rest

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"optisam-backend/common/optisam/logger"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

var testHookCheckZipDir func() = func() {}

type handler struct {
	dir string
}

func (h *handler) uploadHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {

	// parse request
	// const _24K = (1 << 20) * 24
	if err := req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Error("parse multi past form ", zap.Error(err))
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}

	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			//currentDirectory and childDiretoryInfo
			scope := strings.TrimSuffix(hdr.Filename, ".zip")
			dirs, err := getChildDirectories(h.dir, scope)
			if err != nil {
				logger.Log.Error("cannot get child directories", zap.Error(err))
				http.Error(res, "cannot store files", http.StatusInternalServerError)
				return
			}
			newDir, err := getNewDirName(dirs)
			if err != nil {
				logger.Log.Error("cannot make new directory", zap.Error(err))
				http.Error(res, "cannot store files", http.StatusInternalServerError)
				return
			}
			fmt.Println(newDir)
			destDir := filepath.Join(h.dir, scope, newDir)
			fmt.Println(destDir)
			// open uploaded
			infile, err := hdr.Open()
			if err != nil {
				logger.Log.Error("cannot make directory", zap.Error(err))
				http.Error(res, "cannot store files", http.StatusInternalServerError)
				return
			}
			fmt.Println(hdr.Filename)
			// open destination
			var outfile *os.File
			fn := filepath.Join(h.dir, hdr.Filename)

			if outfile, err = os.Create(fn); nil != err {
				logger.Log.Error("cannot create file", zap.Error(err))
				http.Error(res, "cannot store files", http.StatusInternalServerError)
				return
			}
			var written int64
			if written, err = io.Copy(outfile, infile); nil != err {
				logger.Log.Error("cannot copy content of files", zap.Error(err))
				// if all contents are not copied remove the files
				if err := os.Remove(fn); err != nil {
					logger.Log.Error("cannot remove", zap.Error(err))
					http.Error(res, "cannot store files", http.StatusInternalServerError)
					return
				}
				http.Error(res, "cannot store files", http.StatusInternalServerError)
				outfile.Close()
				return
			}

			outfile.Close()
			testHookCheckZipDir()
			_, err = unzip(fn, destDir)
			if err != nil {
				logger.Log.Error("cannot copy content of files", zap.Error(err))
				http.Error(res, "cannot store files", http.StatusInternalServerError)
				return
			}

			if err := os.Remove(fn); err != nil {
				logger.Log.Error("cannot remove", zap.Error(err))
				http.Error(res, "cannot store files", http.StatusInternalServerError)
				return
			}
			res.Write([]byte("uploaded file:" + hdr.Filename + ";length:" + strconv.Itoa(int(written))))
		}
	}
}

// Unzip ...
func unzip(src string, dest string) (fnames []string, retErr error) {
	fmt.Println(src, dest)
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer func() {
		fmt.Println(r.Close())
		if retErr != nil {

		}
	}()

	for _, f := range r.File {
		// fmt.Println(f.Name)
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, filepath.Base(f.Name))

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			// TODO: we do not expect a sub folder we will handle thi case later
			continue
		}

		if err := func() error {

			// Make File
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			defer outFile.Close()

			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
			return nil
			// Close the file without defer to close before next iteration of loop
		}(); err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func getChildDirectories(dir string, scope string) ([]string, error) {
	currentDir := filepath.Join(dir, scope)
	if err := os.MkdirAll(currentDir, os.ModePerm); err != nil {
		logger.Log.Error("cannot make directory", zap.Error(err))
		return nil, err
	}
	fis, err := ioutil.ReadDir(currentDir)

	if err != nil {
		return nil, err
	}
	dirs := make([]string, len(fis))
	for i, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		dirs[i] = fi.Name()
	}
	return dirs, nil
}

func getNewDirName(fileInfo []string) (string, error) {
	for i := range fileInfo {
		fileInfo[i] = strings.TrimPrefix(string(fileInfo[i]), "v")
	}
	dirnum := make([]int, len(fileInfo))
	var err error
	for i := range fileInfo {
		dirnum[i], err = strconv.Atoi(string(fileInfo[i]))
		if err != nil {
			logger.Log.Error("cannot find nextDirname", zap.Error(err))
			return "", err
		}
	}

	if len(dirnum) == 0 {
		return "v1", nil
	}

	sort.Sort(sort.Reverse(sort.IntSlice(dirnum)))
	newDirName := "v" + strconv.Itoa(dirnum[0]+1)
	return newDirName, nil
}
