package main

import (
	"os"
	"path/filepath"
	"strconv"
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

func (c *Collection) SetConfig(key, value string) error {
	configFile := filepath.Join(c.ConfigPath, key)
	fh, err := os.Create(configFile)
	defer fh.Close()
	if err != nil {
		return err
	}
	_, err = fh.Write([]byte(value))
	if err != nil {
		return err
	}
	return nil
}

func (c *Collection) GetDataFiles() []string {
	var files []string
	var walker func(path string, info os.FileInfo, err error) error

	walker = func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}
	filepath.Walk(c.DataPath, walker)
	return files
}

func (c *Collection) EnforceMaxItems() {
	max, _ := c.GetConfig("maxitems")
	if max < 1 {
		return
	}
	files := c.GetDataFiles()
	fileCount := int64(len(files))
	if fileCount > max {
		oldFiles := files[0 : fileCount-max]
		for _, file := range oldFiles {
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
