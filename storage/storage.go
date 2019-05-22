package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	u "storage/utils"
)

func uploadFile(w http.ResponseWriter, r* http.Request) {
	limitInMbStr := os.Getenv("STORAGE_LIMIT")
	limitInMb, _ := strconv.Atoi(limitInMbStr)
	maxSizeInByte := int64(limitInMb * 1024 * 1024)

	//Limit DoS by setting a limit to the body reading
	//The 1024 added are fot the content of the metadata, may be augmented if fitted but
	//should be high enough.
	limit := maxSizeInByte +  1024
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	reader, err := r.MultipartReader()
	if err != nil {
		u.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// parse file field
	p, err := reader.NextPart()
	if err != nil { //Maybe treat EOF (&& err != EOF)
 		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if p.FormName() != "file" {
		u.RespondWithMsg(w, http.StatusBadRequest, "file is expected")
		return
	}

	
	// TODO irindul 2019-05-22 : Get file name from metadata
	//Creating file on hard drive
	tmpFile, err := os.Create("/tmp/qqksdjqkjdh")
	if err != nil {
		// TODO irindul 2019-05-22 : Maybe handle with something else rather than http 500 (allowing the client to debug)
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	defer tmpFile.Close()


	buf := bufio.NewReader(p)
	//Prevent from reading too much
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSizeInByte))
	written, err := io.Copy(tmpFile, lmt)

	if err != nil && err != io.EOF {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	if written > maxSizeInByte {
		os.Remove(tmpFile.Name())
		u.RespondWithMsg(w, http.StatusUnprocessableEntity, "file size over limit")
		return
	}
}

func main() {
	http.HandleFunc("/upload", uploadFile)

	port := os.Getenv("STORAGE_PORT")
	addr := ":" + port
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}