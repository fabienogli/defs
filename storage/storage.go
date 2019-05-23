package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	u "storage/utils"
)

<<<<<<< HEAD
func getAbsDirectory() string {
	// TODO irindul 2019-05-22 : Fetch from ENV/DB for base folder
	// Add this to a volume in docker-compose for persitence ;)

	path := "/storage/tmp/"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0600)
	}

	return path
}

func uploadFile(w http.ResponseWriter, r* http.Request) {
	limitInMbStr := os.Getenv("STORAGE_LIMIT")
	limitInMb, _ := strconv.Atoi(limitInMbStr)
	maxSizeInByte := int64(limitInMb * 1024 * 1024)
	log.Printf("max size allowed : %d bytes", maxSizeInByte)

	//Limit DoS by setting a limit to the body reading
	//The 1024 added are for the content of the metadata, may be augmented if fitted but
	//should be high enough.
	limit := maxSizeInByte +  1024
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	reader, err := r.MultipartReader()
	if err != nil {
		u.RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	p, err := reader.NextPart()
	if err != nil {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if p.FormName() != "hash" {
		u.RespondWithMsg(w, http.StatusBadRequest, "hash of the file is expected")
		return
	}

	//Hash is 256-bit long (which is 32 bytes ;) )
	hashSize := 32 //bytes
	hash := make([]byte, hashSize)
	n, err := p.Read(hash)
	if err != nil  && err != io.EOF {
		u.RespondWithError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if n != hashSize {
		u.RespondWithMsg(w, http.StatusUnprocessableEntity, fmt.Sprintf("hash must be %d bit long, was %d", hashSize*8, n))
		return
	}

	fileName := string(hash)

	// parse file field
	p, err = reader.NextPart()
	if err != nil && err != io.EOF { //Maybe treat EOF (&& err != io.EOF)
 		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if p.FormName() != "file" {
		u.RespondWithMsg(w, http.StatusBadRequest, "file is expected")
		return
	}

	//Creating file on hard drive
	dir := getAbsDirectory()
	path := dir + fileName


	tmpFile, err := os.Create(path)
=======
func main() {
	fmt.Println("Starting http file sever")
	http.HandleFunc("/download/", Download)
	err := http.ListenAndServe(":8080", nil)
>>>>>>> can download now, need to test in docker
	if err != nil {
		// TODO irindul 2019-05-22 : Maybe handle with something else rather than http 500 (allowing the client to debug)
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	log.Println("created file in ", path)
	defer tmpFile.Close()


	buf := bufio.NewReader(p)
	//Prevent from reading too much
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSizeInByte))
	written, err := io.Copy(tmpFile, lmt)

	if err != nil && err != io.EOF {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	//Somehow the file was bigger than expected
	if written > maxSizeInByte {
		os.Remove(tmpFile.Name())
		u.RespondWithMsg(w, http.StatusUnprocessableEntity, "file size over limit")
		return
	}
}

func main() {
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/download", Download)

	port := os.Getenv("STORAGE_PORT")
	addr := ":" + port
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func Download(writer http.ResponseWriter, request *http.Request) {
	//First of check if Get is set in the URL
	Filename := request.URL.Query().Get("file")
	if Filename == "" {
		//Get not set, send a 400 bad request
		http.Error(writer, "Get 'file' not specified in url.", 400)
		return
	}
	fmt.Println("Client requests: " + Filename)

	//Check if file exists and open
	Openfile, err := os.Open("./" + Filename)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(writer, "File not found.", 404)
		log.Printf("ERror: %v", err)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	writer.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	writer.Header().Set("Content-Type", FileContentType)
	writer.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(writer, Openfile) //'Copy' the file to the client
	return
}
