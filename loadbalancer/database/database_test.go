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
	err := storage.Create()
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
	storage.Create()
	return key
}

func getGoodFile(key string) File {
	file := File{
		Salt: key,
		DNS: "dns",
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
	err := file.Create()
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
	err := storage.Create()
	check(err, t)
	dbStorage, err := GetStorage(key)
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
	err := file.Create()
	check(err, t)
	dbFile, err := GetFile(key)
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
	err := storage.Create()
	check(err, t)
	// Change dns name
	storage.DNS ="new_dns"
	err = storage.Update()
	check(err, t)
	dbStorage, err := GetStorage(key)
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
	err := file.Create()
	check(err, t)
	file.DNS ="new_dns"
	// We need to update the database in order to save the file
	storage,err := GetStorage(storageKey)
	check(err, t)
	storage.DNS = "new_dns"
	storage.Update()
	err = file.Update()
	check(err, t)
	dbFile, err := GetFile(key)
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
	err := file.Create()
	check(err, t)

	err = file.Create()
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
	err := storage.Create()
	check(err, t)

	err = storage.Create()
	if err == nil {
		t.Errorf("Should be an error, the storage already exists")
	}
}

func TestUpdateNonExistentFile(t *testing.T) {
	log.Println("Not yet implemented")

}

func TestSaveFileWithoutStorage(t *testing.T) {
	log.Println("Not yet implemented")

}

func TestSaveFileWithMalformDNS(t *testing.T) {
	// Initiate connection
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	key := "salt"
	file := getGoodFile(key)
	file.preCreate(conn)
	err := file.Save()
	check(err, t)

	err = file.Create()
	if err == nil {
		t.Errorf("Should be an error, the file already exists")
	}
}

func TestSaveMalformFile(t *testing.T) {
	log.Println("Not yet implemented")

}

func TestSaveMalformStorage(t *testing.T) {
	log.Println("Not yet implemented")
}

//@Not necessary at the moment
func TestChangeLocationOfFile(t *testing.T) {
	log.Println("Not yet implemented")
}

//get the location of the file
func TestWhereIs(t *testing.T) {
}

//Get the best Storage for file
func TestWhereTo(t *testing.T) {
}

func TestSave(t *testing.T) {
}

func TestDelete(t *testing.T) {
}