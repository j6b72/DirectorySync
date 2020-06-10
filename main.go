package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// TODO: Add skipping a dir / file when error occurs

type File struct {
	baseDir BaseDirectory
	path    string
	mode    os.FileMode
}

type CheckedFile struct {
	actOn    []BaseDirectory
	file     File
	checksum string
}

type Directory struct {
	baseDir BaseDirectory
	path    string
	mode    os.FileMode
}

type BaseDirectory struct {
	path string
}

func main() {

	baseDirs := []BaseDirectory{{path: "./left"}, {path: "./right"}}

	allDirectories, allFiles, err := indexFiles(baseDirs)
	if err != nil {
		log.Fatalf("Error while indexing directory: %v", err)
	}
	err = careAboutDirectories(baseDirs, allDirectories)
	if err != nil {
		log.Fatalf("Error while caring about directories: %v", err)
	}
	_, err = compareFiles(baseDirs, allFiles)
	if err != nil {
		log.Fatalf("Could not compare files: %v", err)
	}

	//actions, err := actAccording(compared)
	//if err != nil { log.Fatalf("Could not perform actions: %v", err) }
}

func compareFiles(baseDirs []BaseDirectory, files []File) ([]CheckedFile, error) {
	var checkedFiles []CheckedFile
	for _, file := range files {
		mainFilePath := fmt.Sprintf("%v%v", file.baseDir.path, file.path)
		var actOn []BaseDirectory
		mainFileChecksum, err := checksumFile(mainFilePath)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not generate main file checksum [%v]: %v", mainFilePath, err))
		}
		mainFileInfo, err := os.Stat(mainFilePath)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not get main file info [%v]: %v", mainFilePath, err))
		}
		for _, baseDir := range baseDirs {
			if baseDir != file.baseDir {
				baseFilePath := fmt.Sprintf("%v%v", baseDir.path, file.path)
				_, err := os.Open(baseFilePath)
				if err != nil {
					if os.IsNotExist(err) {
						actOn = append(actOn, baseDir)
						continue
					} else {
						return nil, errors.New(fmt.Sprintf("Could not open baseDir-specific file [%v]: %v", baseFilePath, err))
					}
				}
				baseFileChecksum, err := checksumFile(baseFilePath)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Could not generate checksum for baseDir-specific file [%v]: %v", baseFileChecksum, err))
				}
				if baseFileChecksum == mainFileChecksum {
					continue
				}

				baseFileInfo, err := os.Stat(baseFilePath)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Could not get main file info [%v]: %v", mainFilePath, err))
				}

				if mainFileInfo.ModTime().Nanosecond() < baseFileInfo.ModTime().Nanosecond() {
					actOn = append(actOn, baseDir)
				}
			}
		}
		checkedFiles = append(checkedFiles, CheckedFile{
			actOn:    actOn,
			file:     file,
			checksum: mainFileChecksum,
		})
		// checksumFile()
	}
	return checkedFiles, nil
}

func checksumFile(path string) (string, error) {
	f, err := os.Open(path)
	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func careAboutDirectories(baseDirs []BaseDirectory, directories []Directory) error {
	for _, dir := range directories {
		for _, baseDir := range baseDirs {
			if baseDir != dir.baseDir {
				dirPath := fmt.Sprintf("%v/%v", baseDir.path, dir.path)
				if _, err := os.Stat(dirPath); os.IsNotExist(err) {
					err = os.Mkdir(dirPath, dir.mode)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func indexDirectory(baseDir BaseDirectory, path string) ([]Directory, []File, error) {
	var allDirectories []Directory
	var allFiles []File
	dirContent, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, nil, err
	}

	for _, dirContentItem := range dirContent {
		if dirContentItem.IsDir() {
			directory, files, err := indexDirectory(baseDir, fmt.Sprintf("%v/%v", path, dirContentItem.Name()))
			if err != nil {
				return nil, nil, err
			}
			allDirectories = append(allDirectories, directory...)
			allFiles = append(allFiles, files...)
			allDirectories = append(allDirectories, Directory{
				baseDir: baseDir,
				path:    fmt.Sprintf("%v/%v", path, dirContentItem.Name())[len(baseDir.path):],
				mode:    dirContentItem.Mode(),
			})
		} else {
			allFiles = append(allFiles, File{
				baseDir: baseDir,
				path:    fmt.Sprintf("%v/%v", path, dirContentItem.Name())[len(baseDir.path):],
				mode:    dirContentItem.Mode(),
			})
		}
	}
	return allDirectories, allFiles, nil
}

func indexFiles(baseDirectories []BaseDirectory) ([]Directory, []File, error) {
	var allDirectories []Directory
	var allFiles []File
	for _, baseDir := range baseDirectories {
		directory, files, err := indexDirectory(baseDir, baseDir.path)
		if err != nil {
			return nil, nil, err
		}
		allDirectories = append(allDirectories, directory...)
		allFiles = append(allFiles, files...)
	}
	return allDirectories, allFiles, nil
}
