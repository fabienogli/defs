package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	u "storage/utils"
	"strconv"
	"strings"
)

var downloadDir string

func getAbsDirectory() string {
	path := os.Getenv("STORAGE_DIR")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0600)
	}

	return path
}

type httpUpload struct {
	w         http.ResponseWriter
	r         *http.Request
	sizeLimit int64
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Header)
	limitInMbStr := os.Getenv("STORAGE_LIMIT")
	limitInMb, _ := strconv.Atoi(limitInMbStr)
	maxSizeInByte := int64(limitInMb * 1024 * 1024)

	//Limit DoS by setting a limit to the body reading
	//The 1024 added are for the content of the metadata, may be augmented if fitted but
	//should be high enough.
	limit := maxSizeInByte +  1024
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	upload := httpUpload{
		w: w,
		r: r,
		sizeLimit: maxSizeInByte,
	}

	err := upload.parseMultiPartForm()
	if err != nil {
		//Already handled in parseMultiPartForm()
		return
	}

	u.RespondWithMsg(w, 200, "file uploaded successfully")
}

func (up httpUpload) parseMultiPartForm() error {
	reader, err := up.r.MultipartReader()
	if err != nil {
		u.RespondWithError(up.w, http.StatusBadRequest, err)
		return err
	}

	p, err := reader.NextPart()
	if err != nil {
		u.RespondWithError(up.w, http.StatusInternalServerError, err)
		return err

	}

	fileName, err := up.parseHashToFileName(p)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// parse file field
	p, err = reader.NextPart()
	if err != nil && err != io.EOF {
		u.RespondWithError(up.w, http.StatusInternalServerError, err)
		return err
	}

	if p.FormName() != "file" {
		u.RespondWithMsg(up.w, http.StatusBadRequest, "file is expected")
		return err
	}

	//Creating file on hard drive
	dir := getAbsDirectory()
	path := dir + fileName
	up.writeFileToDisk(path, p)
	return nil
}

func (up httpUpload) parseHashToFileName(p *multipart.Part) (string, error) {
	if p.FormName() != "hash" {
		u.RespondWithMsg(up.w, http.StatusBadRequest, "hash of the file is expected")
		return "", fmt.Errorf("hash file not present")
	}

	//Hash is 256-bit long, encoded in 64 bit for readability)
	hashSize := 64 //bytes
	hash := make([]byte, hashSize)
	n, err := p.Read(hash)
	if err != nil  && err != io.EOF {
		u.RespondWithError(up.w, http.StatusUnprocessableEntity, err)
		return "", err
	}
	if n != hashSize {
		u.RespondWithMsg(up.w, http.StatusUnprocessableEntity, fmt.Sprintf("hash must be %d bit long, was %d", hashSize, n))
		return "", err
	}

	fileName := string(hash)
	tmp := sanitarizeString(fileName)

	if len(tmp) != len(fileName) {
		msg := "hash is not a proper hash"
		u.RespondWithMsg(up.w, http.StatusUnprocessableEntity, msg)
		return "", fmt.Errorf(msg)
	}

	return fileName, nil
}

func sanitarizeString(toSanitarize string) string {
	replacer := strings.NewReplacer("/", "", ".", "")
	return replacer.Replace(toSanitarize)
}

func (up httpUpload) writeFileToDisk(path string, p *multipart.Part) {
	tmpFile, err := os.Create(path)
	if err != nil {
		u.RespondWithError(up.w, http.StatusInternalServerError, err)
		return
	}
	defer tmpFile.Close()

	buf := bufio.NewReader(p)
	//Prevent from reading too much
	lmt := io.MultiReader(buf, io.LimitReader(p, up.sizeLimit))
	written, err := io.Copy(tmpFile, lmt)

	if err != nil && err != io.EOF {
		u.RespondWithError(up.w, http.StatusInternalServerError, err)
		return
	}

	//Somehow the file was bigger than expected
	if written > up.sizeLimit {
		log.Printf("file was removed : size (%d) too big (limit = %d)", written, up.sizeLimit)
		os.Remove(tmpFile.Name())
		u.RespondWithMsg(up.w, http.StatusUnprocessableEntity, "file size over limit")
		return
	}

	log.Println("succesfully created file in ", path)
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

	_, _ = io.Copy(w, file)
	// TODO irindul 2019-05-26 : Handle errors !
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
