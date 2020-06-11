package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

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

type Configuration struct {
	Locations []string
}

func main() {
	// Gather information

	configuration, err := loadFromConfiguration()
	if err != nil {
		log.Fatalf("Could not load from configuration: %v", err)
	}

	baseDirs := getBaseDirectories(configuration.Locations)

	allDirectories, allFiles, err := indexFiles(baseDirs)
	if err != nil {
		log.Fatalf("Error while indexing directory: %v", err)
	}
	compared, err := compareFiles(baseDirs, allFiles)
	if err != nil {
		log.Fatalf("Could not compare files: %v", err)
	}

	// Take action

	err = careAboutDirectories(baseDirs, allDirectories)
	if err != nil {
		log.Fatalf("Error while caring about directories: %v", err)
	}
	err = actAccording(compared)
	if err != nil {
		log.Fatalf("Could not perform actions: %v", err)
	}
}

func getBaseDirectories(locations []string) []BaseDirectory {
	var baseDirectories []BaseDirectory
	for _, location := range locations {
		baseDirectories = append(baseDirectories, BaseDirectory{path: location})
	}
	return baseDirectories
}

func loadFromConfiguration() (Configuration, error) {
	opened, err := os.Open("configuration.json")
	if err != nil {
		return Configuration{}, err
	}
	decoder := json.NewDecoder(opened)
	var returnable Configuration
	err = decoder.Decode(&returnable)
	if err != nil {
		return Configuration{}, err
	}
	return returnable, err
}

func actAccording(compared []CheckedFile) error {
	for _, checkedFile := range compared {
		checkedFilePath := fmt.Sprintf("%v%v", checkedFile.file.baseDir.path, checkedFile.file.path)
		for _, place := range checkedFile.actOn {
			placeFilePath := fmt.Sprintf("%v%v", place.path, checkedFile.file.path)
			err := copyFile(checkedFilePath, placeFilePath)
			log.Printf("Copied file %v to %v\n", checkedFilePath, placeFilePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(from string, to string) error {
	fromExists, err := fileExists(from)
	if err != nil {
		return err
	}
	toExists, err := fileExists(to)
	if err != nil {
		return err
	}

	if !fromExists {
		return errors.New(fmt.Sprintf("Origin file (%v) does not exist.", from))
	}
	if toExists {
		err := os.Remove(to)
		if err != nil {
			return err
		}
	}

	fromOpen, err := os.Open(from)
	defer fromOpen.Close()
	if err != nil {
		return err
	}
	stat, err := fromOpen.Stat()
	if err != nil {
		return err
	}

	toOpen, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, stat.Mode())
	defer toOpen.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(toOpen, fromOpen)
	if err != nil {
		return err
	}

	return nil
}

func fileExists(path string) (bool, error) {
	test, err := os.Open(path)
	defer test.Close()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

				if mainFileInfo.ModTime().After(baseFileInfo.ModTime()) {
					actOn = append(actOn, baseDir)
				}
			}
		}
		checkedFiles = append(checkedFiles, CheckedFile{
			actOn:    actOn,
			file:     file,
			checksum: mainFileChecksum,
		})
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
