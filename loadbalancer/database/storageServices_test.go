package database

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"testing"
)

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
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	key := uint(1)
	storage := getGoodStorage(key)
	err := storage.Create(conn)
	check(err, t)
	sKey := strconv.Itoa(int(key))

	s, err := redis.String(conn.Do("GET", StoragePrefix + sKey))
	if err == redis.ErrNil {
		t.Errorf("Storage doesn't exists\n")
	} else if err != nil {
		t.Errorf("There was an error %v\n", err)
	}
	dbStorage := Storage{}
	err = json.Unmarshal([]byte(s), &dbStorage)
	if dbStorage != storage {
		t.Errorf("Storage from db is not the same %v", dbStorage)
	}
}

func TestGetStorage(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	key := uint(1)
	storage := getGoodStorage(key)
	err := storage.Create(conn)
	check(err, t)
	dbStorage, err := GetStorage(key, conn)
	check(err, t)
	if dbStorage != storage {
		t.Errorf("Storage retrieved is not the same %v\n", dbStorage)
	}
}


func TestUpdateStorage(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")
	// Create storage
	key := uint(1)
	storage := getGoodStorage(key)
	err := storage.Create(conn)
	check(err, t)
	// Change dns name
	storage.DNS ="new_dns"
	err = storage.Update(conn)
	check(err, t)
	dbStorage, err := GetStorage(key, conn)
	check(err, t)
	// check it's indeed the good value
	if dbStorage != storage {
		t.Errorf("Storage retrieved is not the same %v\n", dbStorage)
	}
}

func TestCreateAlreadyExistingStorage(t *testing.T) {
	// Initiate connection
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

	key := uint(1)
	storage := getGoodStorage(key)
	err := storage.Create(conn)
	check(err, t)

	err = storage.Create(conn)
	if err == nil {
		t.Errorf("Should be an error, the storage already exists")
	}
}

func TestSaveMalformStorage(t *testing.T) {
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")

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
}

func TestDeleteStorage(t *testing.T) {
	pool := getPool()
	conn := pool.Get()
	defer conn.Close()
	defer conn.Do("FLUSHALL")
	storage1 := getGoodStorage(uint(1))
	storage1.Create(conn)
	storage2 := getGoodStorage(uint(2))
	storage2.Create(conn)
	err := storage1.Delete(conn)
	check(err, t)

	_, err = GetStorage(storage2.ID, conn)
	check(err, t)
}

