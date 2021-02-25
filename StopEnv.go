package main

import (
	"bufio"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var basePath, function, temp, logfilePath string
	var openFlags int
	var files []string
	var searchStrings = [...]string{"replicaCount: ","replicas: "}
	flag.StringVar(&function, "function", "", "stop or start")
	flag.StringVar(&basePath, "path", "", "basepath to habitat")
	flag.Parse()
	logfilePath = basePath + "\\replicas.log"
	if function == "stop" {
		openFlags = os.O_RDWR|os.O_CREATE|os.O_TRUNC
	} else {
		openFlags = os.O_RDWR
	}
	logfile, err := os.OpenFile(logfilePath, openFlags, 0666)
	if err != nil {
		log.Println("error opening file 'replicas.log'. Error: ", err.Error())
		os.Exit(0)
	}
	defer logfile.Close()
	writer := io.MultiWriter(os.Stdout, logfile)
	log.SetOutput(writer)

	err = filepath.Walk(basePath, listFiles(&files))
	if err != nil {
		panic(err)
	}
	if function == "stop" {
		for _, searchString := range searchStrings {
			for _, file := range files {
				temp = readContent(file, searchString)
				if temp != "" {
					log.Println(file, "|", temp)
					changeContent(file, temp, searchString + "0")
				}
			}
		}
	} else if function == "start" {
		for _, file := range files {
			temp = readContent(logfilePath, file)
			if temp != "" {
				temp = temp[strings.Index(temp, "|")+2:]
				changeContent(file, temp[:strings.Index(temp, ":")] + ": 0", temp)
			}
		}
	}
}

func readContent(file string, search string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()

	lineScanner := bufio.NewScanner(f)
	line := 1
	temp := ""
	for lineScanner.Scan() {
		if strings.Contains(lineScanner.Text(), search) {
			temp = strings.TrimSpace(lineScanner.Text())
			break
		}
		line++
	}
	return temp
}

func changeContent(file string, old string, new string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	newContents := strings.Replace(string(content), old, new, 1)
	err = ioutil.WriteFile(file, []byte(newContents), 0)
	if err != nil {
		panic(err)
	}
}

func listFiles(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if !info.IsDir() {
			if filepath.Ext(path) == ".yaml" {
				*files = append(*files, path)
			}
		}
		return nil
	}
}
