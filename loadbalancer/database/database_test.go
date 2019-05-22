package database

import (
	"testing"
	"encoding/json"
	"log"

	"github.com/gomodule/redigo/redis"
)

//@TODO

//Sauvegarder un hash qui n'existe pas et le get
//Sauvegarder un hash qui existe et erreur
//Erreur full
//get un hash sauvegarder
//Delete un fichier

func check(err error, t *testing.T) {
	if err != nil {
		t.Errorf("There was an error %v\n", err)
	}
}

// Test client connection
func TestClient(t *testing.T) {
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	pong, err := conn.Do("PING")
	check(err, t)
	s, err := redis.String(pong, err)
	check(err, t)
	if s != "PONG" {
		t.Errorf("Wrong answer %s\n", s)
	}
}

func getGoodStorage(key uint) Storage {
	storage := Storage{
		ID: key,
		DNS: "dns",
		Used: 0,
		Total: 10,
	}
	return storage
}

func TestCreateStorage(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	key := uint(1)
	storage := getGoodStorage(key)
	err := storage.Create(conn)
	check(err, t)

	s, err := redis.String(conn.Do("GET", StoragePrefix + string(key)))
	if err == redis.ErrNil {
		t.Errorf("Storage doesn't exists\n")
	} else if err != nil {
		t.Errorf("There was an error %v\n", err)
	}
	dbStorage := Storage{}
	err = json.Unmarshal([]byte(s), &dbStorage)
	if (dbStorage != storage) {
		t.Errorf("Storage from db is not the same %v", dbStorage)
	}
}

func (file File) preCreate(c redis.Conn) uint {
	key := uint(1)
	storage := getGoodStorage(key)
	storage.DNS = file.DNS
	storage.Create(c)
	return key
}

func getGoodFile(key string) File {
	file := File{
		Hash: key,
		DNS: "dns",
		Size: 3,
	}
	return file
}

func TestCreateFile(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	key := "file"
	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	s, err := redis.String(conn.Do("GET", FilePrefix + key))
	if err == redis.ErrNil {
		t.Errorf("Storage doesn't exists\n")
	} else if err != nil {
		t.Errorf("There was an error %v\n", err)
	}
	dbFile := File{}
	err = json.Unmarshal([]byte(s), &dbFile)
	if (dbFile != file) {
		t.Errorf("Storage from db is not the same %v", dbFile)
	}
}

func TestGetStorage(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()
	key := uint(1)
	storage := getGoodStorage(key)
	err := storage.Create(conn)
	check(err, t)
	dbStorage, err := GetStorage(key, nil)
	check(err, t)
	if dbStorage != storage {
		t.Errorf("Storage retrieved is not the same %v\n", dbStorage)
	}
}

func TestGetFile(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()
	key := "salt"
	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	dbFile, err := GetFile(key, nil)
	check(err, t)
	if dbFile != file {
		t.Errorf("File retrieved is not the same %v\n", dbFile)
	}
}
func TestUpdateStorage(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()
	// Create storage
	key := uint(1)
	storage := getGoodStorage(key)
	err := storage.Create(conn)
	check(err, t)
	// Change dns name
	storage.DNS ="new_dns"
	err = storage.Update(conn)
	check(err, t)
	dbStorage, err := GetStorage(key, nil)
	check(err, t)
	// check it's indeed the good value
	if dbStorage != storage {
		t.Errorf("File retrieved is not the same %v\n", dbStorage)
	}
}

func TestUpdateFile(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()
	key := "salt"

	file := getGoodFile(key)
	storageKey := file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)
	file.DNS ="new_dns"
	// We need to update the database in order to save the file
	storage,err := GetStorage(storageKey, nil)
	check(err, t)
	storage.DNS = "new_dns"
	storage.Update(conn)
	err = file.Update(conn)
	check(err, t)
	dbFile, err := GetFile(key, nil)
	check(err, t)
	if dbFile != file {
		t.Errorf("File retrieved is not the same %v\n", dbFile)
	}
}

func TestCreateAlreadyExistingFile(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()
	key := "salt"
	
	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Create(conn)
	check(err, t)

	err = file.Create(conn)
	if err == nil {
		t.Errorf("Should be an error, the file already exists")
	}
}

func TestCreateAlreadyExistingStorage(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	key := uint(1)
	storage := getGoodStorage(key)
	err := storage.Create(conn)
	check(err, t)

	err = storage.Create(conn)
	if err == nil {
		t.Errorf("Should be an error, the storage already exists")
	}
}

func TestSaveFileWithMalformDNS(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	key := "salt"
	file := getGoodFile(key)
	file.preCreate(conn)
	file.Hash = "bouh"
	err := file.Save(conn)
	check(err, t)

	err = file.Create(conn)
	if err == nil {
		t.Errorf("Should be an error, the file already exists")
	}
}

func TestSaveMalformFile(t *testing.T) {
		// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	f1 := File {
		Hash: "nul",
		DNS: "dns",
	}
	f2 := File {
		Size: 1,
		DNS: "dns",
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

func TestSaveMalformStorage(t *testing.T) {
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	s1 := Storage {
		ID	:1,
		DNS   :"dns",
		Used  :11,
		Total :10,
	}
	err := s1.Save(conn)

	if err == nil {
		t.Errorf("More used space than total space")
	}

	s2 := Storage {
		ID	:1,
		Used  :1,
		Total :10,
	}
	err = s2.Save(conn)

	if err == nil {
		t.Errorf("No DNS")
	}

	s3 := Storage {
		ID	:1,
		DNS   :"dns",
		Total :10,
	}
	err = s3.Save(conn)
	if err == nil {
		t.Errorf("No Used space")
	}

	s4 := Storage {
		ID	:1,
		DNS   :"dns",
		Used: 1,
		Total :10,
	}
	err = s4.Save(conn)
	if err == nil {
		t.Errorf("No Used space")
	}

	log.Println("Not yet implemented")
}