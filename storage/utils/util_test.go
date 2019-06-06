package utils

import (
	"bytes"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestIsDevelopmentMode(t *testing.T) {
	err := os.Setenv("STORAGE_MODE", "development")
	if err != nil {
		t.Errorf("Setting env STORAGE_MODE failed : %s\n", err)
	}
	shouldBeTrue := IsDevelopmentMode()

	if shouldBeTrue != true {
		t.Errorf("storage mode failed, expected %t, got %t", true, shouldBeTrue)
	}

	err = os.Setenv("STORAGE_MODE", "production")
	if err != nil {
		t.Errorf("setting env STORAGE_MODE to production failed : %s", err)
	}

	shouldBeFalse := IsDevelopmentMode()
	if shouldBeFalse != false {
		t.Errorf("storage mode failed : expected %t, got %t", false, shouldBeFalse)
	}
}

func TestRespondWithJSON(t *testing.T) {
	//Request with empty data
	recorder := httptest.NewRecorder()
	RespondWithJSON(recorder, 200, nil)
	if recorder.Code != 200 {
		t.Errorf("status code must be 200, it was %d", recorder.Code)
	}
	contentType := recorder.Header().Get("Content-Type")
	if contentType == "" {
		t.Errorf("response must have a Content-Type header !")
	}
	if strings.Compare(contentType, "application/json") != 0 {
		t.Errorf("response must have Content-Type set to application/json, got %s", contentType)
	}

	responseSize := len([]byte("null"))
	buf := make([]byte, responseSize)
	n, err := recorder.Body.Read(buf)
	if err != nil {
		t.Errorf("error reading bytes from Body : %s\n", err)
	}

	if n != responseSize {
		t.Errorf("should be able to read %d bytes, but read %d", responseSize, n)
	}

	//Try a new request with data
	recorder = httptest.NewRecorder()
	mock := make(map[string]string)
	mock["testKey"] = "testValue"
	RespondWithJSON(recorder, 200, mock)

	//No need to recheck headers, we just check the data itself
	expected := []byte(`{"testKey":"testValue"`)
	buf = make([]byte, len(expected))
	n, err = recorder.Body.Read(buf)
	if err != nil {
		t.Errorf("error reading Body : %s ", err)
	}
	if n != len(expected) {
		t.Errorf("shoud be able to read %d bytes, but read %d\n", len(expected), n)
	}

	if !bytes.Equal(expected, buf) {
		t.Errorf("the data is not equal !")
	}
}