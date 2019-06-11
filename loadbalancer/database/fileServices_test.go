package database

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"log"
	"testing"
	"time"
)

func (file File) preCreate(c redis.Conn) uint {
	storage := getGoodStorage(file.DNS)
	storage.DNS = "lol"
	err := storage.Create(c)
	if err != nil {
		log.Printf("Error in precreate %v\n", err)
	}
	return file.DNS
}

func TestCreateFile(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	key := "file"
	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	s, err := redis.String(conn.Do("GET", FilePrefix + key))
	if err == redis.ErrNil {
		t.Errorf("File doesn't exists\n")
	} else if err != nil {
		t.Errorf("There was an error %v\n", err)
	}
	dbFile := File{}
	err = json.Unmarshal([]byte(s), &dbFile)
	if (dbFile != file) {
		t.Errorf("File from db is not the same %v", dbFile)
	}
}

func TestCreateBigFile(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	key := "file"
	file := getGoodFile(key)
	file.preCreate(conn)
	file.Size =100
	err := file.Create(conn)
	if err == nil {
		t.Errorf("Should be an error, file too big")
	}
}

func TestUpdateFile(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")
	key := "salt"

	file := getGoodFile(key)
	storageKey := file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	file.DNS = 2
	// We need to update the database in order to save the file
	storage,err := GetStorage(storageKey, conn)
	check(err, t)
	storage.ID = 2
	storage.Create(conn)
	err = file.Update(conn)
	check(err, t)
	dbFile, err := GetFile(key, conn)
	check(err, t)
	if dbFile != file {
		t.Errorf("File retrieved is not the same %v\n", dbFile)
	}
}

func TestCreateAlreadyExistingFile(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")
	key := "salt"

	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	err = file.Create(conn)
	if err == nil {
		t.Errorf("Should be an error, the file already exists, error %v\n", err)
	}
}

func TestSaveFileWithMalformDNS(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	key := "salt"
	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	file.DNS = 3
	err = file.Update(conn)
	if err == nil {
		t.Errorf("The dns doesn't exist")
	}
}

func TestSaveMalformFile(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	f1 := File {
		Hash: "nul",
		DNS: 0,
	}
	f2 := File {
		Size: 1,
		DNS: 0,
	}
	f3 := File {
		Hash: "nul",
		Size: 1,
	}
	err := f1.Save(conn)
	if err == nil {
		t.Errorf("No size, file should not be able to be saved")
	}
	err = f2.Save(conn)
	if err == nil {
		t.Errorf("No hash here")
	}
	err = f3.Save(conn)
	if err == nil {
		t.Errorf("No dns here")
	}
}

func TestDeleteFile(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	keyd := "file1"
	keyf := "file2"
	file1 := getGoodFile(keyd)
	file2 := getGoodFile(keyf)
	file1.preCreate(conn)
	file1.Create(conn)
	file2.Create(conn)


	err := file1.Delete(conn)
	check(err, t)

	_, err = GetFile(file2.Hash, conn)
	check(err, t)
}

func TestExpFile(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	key := "file"
	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	ttl := 1 //seconds
	err = file.SetExp(ttl, conn)
	check(err, t)
	duration := ttl + 1
	time.Sleep(time.Duration(duration) * time.Second)
	_, err = redis.String(conn.Do("GET", FilePrefix + key))
	if err != redis.ErrNil {
		t.Error("File should be erased")
	}
}

func TestGetFile(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")
	key := "salt"
	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	dbFile, err := GetFile(key, conn)
	check(err, t)
	if dbFile != file {
		t.Errorf("File retrieved is not the same %v\n", dbFile)
	}
}

func getGoodFile(key string) File {
	file := File{
		Hash: key,
		DNS: 1,
		Size: 3,
	}
	return file
}