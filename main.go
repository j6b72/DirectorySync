package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// TODO: Add skipping a dir / file when error occurs

type File struct {
	baseDir baseDirectory
	path    string
	mode    os.FileMode
}

type Directory struct {
	baseDir baseDirectory
	path    string
	mode    os.FileMode
}

type baseDirectory struct {
	path string
}

func main() {

	baseDirs := []baseDirectory{{path: "./left"}, {path: "./right"}}

	allDirectories, _, err := indexFiles(baseDirs)
	if err != nil {
		log.Fatalf("Error while indexing directory: %v", err)
	}
	err = careAboutDirectories(baseDirs, allDirectories)
	if err != nil {
		log.Fatalf("Error while caring about directories: %v", err)
	}
	//compared, err := compareFiles(allFiles)
	//if err != nil { log.Fatalf(err) }
	//actions, err := performActions(compared)
	//if err != nil { log.Fatalf(err) }
}

func careAboutDirectories(baseDirs []baseDirectory, directories []Directory) error {
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

func indexDirectory(baseDir baseDirectory, path string) ([]Directory, []File, error) {
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

func indexFiles(baseDirectories []baseDirectory) ([]Directory, []File, error) {
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
