package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
)

func uploadFile(w http.ResponseWriter, r* http.Request) {
	maxSize := int64(3700 << 20) //3000MB

	//Prevent DoS by setting a limit to the body reading
	limit := maxSize +  1024
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// parse file field
	p, err := reader.NextPart()
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if p.FormName() != "file" {
		http.Error(w, "file is expected", http.StatusBadRequest)
		return
	}

	tmpFile, err := os.Create("/tmp/qqksdjqkjdh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tmpFile.Close()



	buf := bufio.NewReader(p)
//	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize - 511))
	written, err := io.Copy(tmpFile, buf)

	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if written > maxSize {
		os.Remove(tmpFile.Name())
		http.Error(w, "file size over limit", http.StatusBadRequest)
		return
	}
	// schedule for other stuffs (s3, scanning, etc.)
}

func main() {
	http.HandleFunc("/upload", uploadFile)

	port := os.Getenv("STORAGE_PORT")
	addr := ":" + port
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}