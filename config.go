package config

import (
	"bytes"
	"encoding/json"
	"golanger.com/log"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sync"
)

var (
	mu          sync.RWMutex
	ignoreFirst byte           = '.'
	regexpNote  *regexp.Regexp = regexp.MustCompile(`#.*`)
)

type Config struct {
	dataType  string
	data      []byte
	directory string
	files     []string
}

func format(configPath string) []byte {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal("<format> error: ", err)
	}

	return regexpNote.ReplaceAll(data, []byte(``))
}

func readFiles(files ...string) []byte {
	lfs := len(files)
	chContent := make(chan []byte, lfs)

	for _, file := range files {
		go func(chContent chan []byte, configPath string) {
			chContent <- format(configPath)
		}(chContent, file)
	}

	buf := bytes.NewBufferString(`{`)

	for i := 1; i <= lfs; i++ {
		content := <-chContent
		if len(content) == 0 {
			continue
		}

		buf.Write(content)
		if i < lfs {
			buf.WriteString(",")
		}
	}

	buf.WriteString(`}`)
	var contentBuf bytes.Buffer
	json.Compact(&contentBuf, buf.Bytes())

	return contentBuf.Bytes()
}

func loadFiles(files ...string) *Config {
	conf := &Config{
		dataType: "files",
	}

	for _, file := range files {
		fileName := filepath.Base(file)
		if fileName[0] == ignoreFirst {
			continue
		}

		conf.files = append(conf.files, file)
	}

	conf.data = readFiles(conf.files...)

	return conf
}

func Data(data string) *Config {
	var buf bytes.Buffer
	json.Compact(&buf, []byte(`{`+data+`}`))
	conf := &Config{
		dataType: "data",
		data:     buf.Bytes(),
	}

	return conf
}

func Files(files ...string) *Config {
	return loadFiles(files...)
}

func Glob(pattern string) *Config {
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatal("<Glob> error: ", err)
	}

	return loadFiles(files...)
}

func Dir(configDir string) *Config {
	fis, err := ioutil.ReadDir(configDir)
	if err != nil {
		log.Fatal("<Dir> error: ", err)
	}

	conf := &Config{
		dataType:  "directory",
		directory: filepath.Clean(configDir),
	}

	var files []string
	for _, fi := range fis {
		fileName := fi.Name()
		if fi.IsDir() || fileName[0] == ignoreFirst {
			continue
		}

		files = append(files, filepath.Join(conf.directory, fileName))
	}

	return loadFiles(files...)
}

func (c *Config) Load(i interface{}) *Config {
	err := json.Unmarshal(c.data, i)
	if err != nil {
		log.Debug("<Config.Load> jsonData: ", c.String())
		log.Fatal("<Config.Load> error: ", err)
	}

	return c
}

func (c *Config) Bytes() []byte {
	return c.data
}

func (c *Config) String() string {
	return string(c.data)
}

func (c *Config) Target() string {
	return c.directory
}
