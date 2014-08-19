package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Collection struct {
	Name       string
	BasePath   string
	DataPath   string
	ConfigPath string
}

func (c *Collection) GetConfig(key string) (int64, error) {
	var data []byte
	configFile := filepath.Join(c.ConfigPath, key)
	fh, err := os.Open(configFile)
	defer fh.Close()
	if err != nil {
		return 0, err
	}
	stat, err := fh.Stat()
	if err != nil {
		return 0, err
	}
	data = make([]byte, stat.Size())
	if _, err = fh.Read(data); err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(data), 10, 64)
}

func (c *Collection) SetConfig(key string, value int64) error {
	configFile := filepath.Join(c.ConfigPath, key)
	fh, err := os.Create(configFile)
	defer fh.Close()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(fh, "%d", value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Collection) GetDataFiles() []string {
	var files []string
	var walker func(path string, info os.FileInfo, err error) error
	var maxAge int64
	var ageThreshold int64

	maxAge, _ = c.GetConfig("maxage")
	ageThreshold = time.Now().UnixNano() - (maxAge * 1000000)

	walker = func(path string, info os.FileInfo, err error) error {
		var fileAge int64
		fileAge, err = strconv.ParseInt(filepath.Base(path), 10, 64)
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			// enforce max age, remove old items
			if maxAge <= 0 {
				files = append(files, path)
			} else {
				if fileAge >= ageThreshold {
					files = append(files, path)
				} else {
					os.Remove(path)
				}
			}
		}
		return nil
	}
	filepath.Walk(c.DataPath, walker)
	return files
}

func (c *Collection) EnforceCapacity() {
	maxItems, _ := c.GetConfig("maxitems")
	maxAge, _ := c.GetConfig("maxage")
	files := c.GetDataFiles()
	fileCount := int64(len(files))
	indexThreshold := fileCount - maxItems
	ageThreshold := time.Now().UnixNano() - (maxAge * 1000000)
	for i, file := range files {
		fileAge, _ := strconv.ParseInt(filepath.Base(file), 10, 64)
		if (maxItems > 0 && int64(i) < indexThreshold) || (maxAge > 0 && fileAge < ageThreshold) {
			os.Remove(file)
		}
	}
}

func CreateCollection(name string, dataPath string) *Collection {
	c := new(Collection)
	c.Name = name
	c.BasePath = filepath.Join(dataPath, name)
	c.DataPath = filepath.Join(c.BasePath, "data")
	c.ConfigPath = filepath.Join(c.BasePath, "config")
	return c
}
