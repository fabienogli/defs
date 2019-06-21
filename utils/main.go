package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	envFiles := findFile(".", ".env.example" )
	for _, env := range envFiles {
		var name string
		split := strings.Split(env, "/")
		if len(split) > 1 {
			name = split[0] + "/.env"
		} else {
			name = ".env"
		}
		content, err := ioutil.ReadFile(env)
		check(err)
		f, err := os.Create(name)
		check(err)
		defer f.Close()
		_, err = f.Write(content)
		check(err)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func findFile(directory string, ext string) []string {
	var files []string
	err := filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == ext {
				files = append(files, path)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return files
}
