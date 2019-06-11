package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	u "storage/utils"
	"strconv"
	"strings"
)


var downloadDir string

func getAbsDirectory() string {
	path := os.Getenv("STORAGE_DIR")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.MkdirAll(path, 0600)
	}

	return path
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	limitInMbStr := os.Getenv("STORAGE_LIMIT")
	limitInMb, _ := strconv.Atoi(limitInMbStr)
	maxSizeInByte := int64(limitInMb * 1024 * 1024)

	//Limit DoS by setting a limit to the body reading
	//The 1024 added are for the content of the metadata, may be augmented if fitted but
	//should be high enough.
	limit := maxSizeInByte +  1024
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	r.ParseMultipartForm(maxSizeInByte)
	filename, ttl, err := parseMultiPartForm(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch err.(type) {
		case HashTooLarge, HashTooShort, HashInvalid:
			statusCode = http.StatusUnprocessableEntity
		case BadRequest:
			statusCode = http.StatusBadRequest
		default:
			u.RespondWithError(w, statusCode, err)
			return
		}
		u.RespondWithMsg(w, statusCode, msg)
		return
	}

	//Creating file on hard drive
	file, _, err := r.FormFile("file")
	if err != nil {
		u.RespondWithError(w, http.StatusBadRequest, err)
	}
	defer file.Close()
	err = writeFileToDisk(filename, file, limit)
	if err != nil {
		switch err.(type) {
		case CannotCreateFile:
			u.RespondWithError(w, http.StatusInternalServerError, err)
			return
		}
	}

	err = setTTL(ttl, filename)
	if err != nil {
		DeleteFile(filename)
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	u.RespondWithMsg(w, 200, "file uploaded successfully")
}

func DeleteFile(filename string) {
	err := os.RemoveAll(filename)
	if err != nil {
		log.Panicf("could not delete file %s : %s", filename, err)
	}
}

func setTTL(ttl string, hash string) error{
	echo := exec.Command("echo", fmt.Sprintf("rm %s%s", getAbsDirectory(), hash))
	at := exec.Command("at", fmt.Sprintf("now + %s", ttl))
	r, w := io.Pipe()

	echo.Stdout = w
	at.Stdin = r

	err := echo.Start()
	if err != nil {
		return err
	}
	err = at.Start()
	if err != nil {
		return err
	}
	err = echo.Wait()
	if err != nil {
		return err
	}

	w.Close()
	err = at.Wait()
	r.Close()
	if err != nil {
		return err
	}

	return nil
}

func parseMultiPartForm(r *http.Request) (filename string, ttl string,  err error){

	hash := r.FormValue("hash")

	if hash == "" {
		return "", "", NewHashNotFound()
	}

	filename, err = parseHash(strings.NewReader(hash))
	if err != nil {
		return "", "", err
	}

	ttl = r.FormValue("ttl")
	if ttl == "" {
		//default ttl to one day
		ttl = "1 day"
	}

	return filename, ttl, nil
}

func parseHash(r io.Reader) (string, error) {
	//Hash is 256-bit long, encoded in 64 bit for readability)
	hashSize := 64 //bytes
	hash := make([]byte, hashSize)

	n, err := r.Read(hash)
	if err != nil && err != io.EOF {
		return "", NewHashInvalid()
	}
	if n < hashSize {
		return "", NewHashTooShort()
	}

	fileName := string(hash)
	sanitarized := sanitarizeString(fileName)

	if len(sanitarized) != len(fileName) {
		return "", NewHashInvalid()
	}

	return fileName, nil
}

func sanitarizeString(toSanitarize string) string {
	replacer := strings.NewReplacer("/", "", ".", "", " ", "")
	return replacer.Replace(toSanitarize)
}

func writeFileToDisk(filename string, r io.Reader, sizeLimit int64) error{
	dir := getAbsDirectory()
	path := dir + filename

	tmpFile, err := os.Create(path)
	if err != nil {
		return NewCannotCreateFile(err)
	}
	defer tmpFile.Close()

	reader := bufio.NewReader(r)
	//Prevent from reading too much
	lmt := io.MultiReader(reader, io.LimitReader(r, sizeLimit))
	written, err := io.Copy(tmpFile, lmt)

	if err != nil && err != io.EOF {
		return NewInternalError(err.Error())
	}

	//Somehow the file was bigger than expected
	if written > sizeLimit {
		log.Printf("file was removed : size (%d) too big (limit = %d)", written, sizeLimit)
		_ = os.Remove(tmpFile.Name())
		return NewFileTooLarge(fmt.Errorf("expected max size %d, got %d", sizeLimit, written))
	}

	log.Println("succesfully created file in ", path)
	return nil
}

func download(w http.ResponseWriter, r *http.Request) {
	downloadDir = getAbsDirectory()
	vars := mux.Vars(r)
	fileName := vars["file"]


	if fileName == "" {
		u.RespondWithMsg(w, http.StatusBadRequest, "File was not specified")
		return
	}

	fileName = sanitarizeString(fileName)
	log.Printf("client requested %s\n", fileName)

	path := downloadDir + fileName

	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			u.RespondWithMsg(w, http.StatusNotFound, fmt.Sprintf("file %s not found", fileName))
			return
		}
		u.RespondWithError(w, http.StatusInternalServerError, err)

	}


	fileHeader := make([]byte, 512)

	//Copy the headers into the FileHeader buffer
	_, _ = file.Read(fileHeader)

	//Get content type of file
	contentType := http.DetectContentType(fileHeader)

	//Get the file size
	stat, _ := file.Stat()                         //Get info from file
	fileSize := strconv.FormatInt(stat.Size(), 10) //Get file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fileSize)

	//Send the file
	_, err = file.Seek(0, 0)
	if err != nil {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return

	}

	_, err = io.Copy(w, file)
	if err != nil {
		u.RespondWithError(w, http.StatusBadGateway, err)
		return
	}
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/upload", uploadFile)
	r.HandleFunc("/download/{file}", download).Methods("GET")
	http.Handle("/", r)

	port := os.Getenv("STORAGE_PORT")
	addr := ":" + port
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
