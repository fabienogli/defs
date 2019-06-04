package routers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	s "supervisor/storage"
	u "supervisor/utils"
	"time"
)

type HashErr uint8

func (HashErr) Error() string {
	return "Hash already exists"
}

func hashSHA256(fileName string) string {
	t := time.Now().String()
	hash := sha256.Sum256([]byte(fileName + t))
	return hex.EncodeToString(hash[:])
}

func upload(w http.ResponseWriter, r *http.Request) {
	limitInMbStr := os.Getenv("STORAGE_LIMIT")
	limitInMb, _ := strconv.Atoi(limitInMbStr)
	maxSizeInByte := int64(limitInMb * 1024 * 1024)
	r.ParseMultipartForm(maxSizeInByte)

	filename := r.FormValue("filename")
	if filename == "" {
		u.RespondWithMsg(w, http.StatusUnprocessableEntity, "filename must be provided")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		u.RespondWithMsg(w, http.StatusUnprocessableEntity, "file must be provided")
		log.Println(err)
		return
	}
	defer file.Close()

	lb, err := s.NewLoadBalancerClient()
	defer lb.Close()

	if err != nil {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	var baseUrl = ""
	var hash = hashSHA256(filename)
	for true {
		response, err := lb.WhereTo(hash, int(fileHeader.Size))
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				u.RespondWithError(w, http.StatusGatewayTimeout, err)
				return
			}
			u.RespondWithError(w, http.StatusInternalServerError, err)
			return
		}
		baseUrl, err = getBaseUrl(response, w)
		if err != nil {
			if _, ok := err.(HashErr); ok {
				hash = hashSHA256(filename)
				continue
			} else {
				//Handled in getBaseUrl()
				return
			}
		}
		break
	}

	url := baseUrl + "upload"
	body := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(body)

	bodyWriter.WriteField("hash", hash)

	// add a form file to the body
	fileWriter, err := bodyWriter.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	_, err = io.Copy(fileWriter, file)
	bodyWriter.Close()

	// send request
	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest(http.MethodGet, url, body)

	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		u.RespondWithError(w, http.StatusBadGateway, err)
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		infos := map[string]string{
			"hash": hash,
		}

		u.RespondWithJSON(w, http.StatusOK, infos)
		return
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	lb, err := s.NewLoadBalancerClient()
	defer lb.Close()
	if err != nil {
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	response, err := lb.WhereIs(hash)
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			u.RespondWithError(w, http.StatusGatewayTimeout, err)
			return
		}
		u.RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	baseUrl, err := getBaseUrl(response, w)
	if err != nil {
		//Error handled in getBaseUrl()
		return
	}
	url := baseUrl + "download/" + hash

	proxyRequest, err := http.NewRequest(http.MethodGet, url, r.Body)
	hostName := os.Getenv("SUPERVISOR_HTTP_HOST")
	proxyRequest.Header.Set("HOST", hostName)

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyRequest.Header = make(http.Header)
	for h, val := range r.Header {
		proxyRequest.Header[h] = val
	}

	client := &http.Client{}
	resp, err := client.Do(proxyRequest)
	if err != nil {
		u.RespondWithError(w, http.StatusBadGateway, err)
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func getBaseUrl(response string, w http.ResponseWriter) (string, error) {
	respPart := strings.Split(response, " ")
	if respPart[0] != s.Ok.String() {
		switch respPart[0] {
		case s.HashAlreadyExisting.String():
			//Todo rehash file and loop...
			return "", HashErr(uint8(s.HashAlreadyExisting))
		case s.HashNotFound.String():
			err := fmt.Errorf("hash not found")
			u.RespondWithError(w, http.StatusNotFound, err)
			return "", err
		default:
			err := fmt.Errorf("response not implemented : %s", respPart[0])
			u.RespondWithError(w, http.StatusNotImplemented, err)
			return "", err
		}
	}

	storeDns := respPart[1]
	storagePortStr := os.Getenv("STORAGE_PORT")
	protocol := os.Getenv("STORAGE_PROTOCOL")

	url := fmt.Sprintf("%s://%s:%s/", protocol, storeDns, storagePortStr)
	return url, nil
}

func SetStoreRoute(r *mux.Router) *mux.Router {
	r.HandleFunc("/file", upload).Methods("POST")
	r.HandleFunc("/file/{hash}", download).Methods("GET")
	return r
}