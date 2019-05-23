package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

const (
	FilePrefix = "file:"
	StoragePrefix = "storage:"
)

type File struct {
	Hash string `json:"hash"`
	DNS  string `json:"dns"`
	Size uint `json:"size"`
}

type Storage struct {
	ID		uint  `json:"id"`
	DNS   string `json:"dns"`
	Used  uint `json:"used"`
	Total uint `json:"total"`
}

type Cud interface {
	Create(con redis.Conn) error
	Update(con redis.Conn) error
	Save(conn redis.Conn) error
	Delete(conn redis.Conn) error
}

func newPool() *redis.Pool {
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "routing_table:6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func getKeys(pattern string, conn redis.Conn) ([]string, error) {
	tmp, err := redis.Strings(conn.Do("KEYS", pattern+"*"))
	if err != nil {
		return []string{}, err
	}
	keys := make([]string, len(tmp))
	for index, tm := range tmp {
		key := strings.Split(tm, ":")
		if len(key) < 1 {
			return []string{}, fmt.Errorf("Key with the pattern %s not found\n", pattern)
		}
		keys[index] = key[1]
	}
	return keys, nil
}

func GetStorages(conn redis.Conn) ([]Storage, error) {
	keys, err := getKeys(StoragePrefix, conn)
	if err != nil {
		return []Storage{}, err	
	}
	storages := make([]Storage, len(keys))
	for index, key := range keys {
		ukey, err := strconv.ParseUint(key, 10, 32)
		if err != nil {
			return []Storage{}, err
		}
		storages[index], err = GetStorage(uint(ukey), conn)
	}
	return storages, nil
}

func GetStorage(key uint, conn redis.Conn) (Storage, error) {
	sKey := strconv.Itoa(int(key))
	s, err := redis.String(conn.Do("GET", StoragePrefix + mediate(sKey )))
	if err == redis.ErrNil {
		return Storage{}, err
	} else if err != nil {
		return Storage{}, err
	}
	storage := Storage{}
	err = json.Unmarshal([]byte(s), &storage)
	if err != nil {
		return Storage{}, err
	}
	return storage, nil
}

func GetFile(hash string, conn redis.Conn) (File, error) {
	s, err := redis.String(conn.Do("GET", FilePrefix + mediate(hash)))
	if err == redis.ErrNil {
		return File{}, err
	} else if err != nil {
		return File{}, err
	}
	file := File{}
	err = json.Unmarshal([]byte(s), &file)
	if err != nil {
		return File{}, err
	}
	return file, nil
}

//should throw error if there is already a value cause redis overwrite set
func (storage Storage) Create(conn redis.Conn) error {
	_, err := GetStorage(storage.ID, conn)
	if err == redis.ErrNil {
		err = storage.Save(conn)
		return err
	}
	return fmt.Errorf("Storage already Exist\n")
}

func (file File) Create(conn redis.Conn) error {
	_, err := GetFile(file.Hash, conn)
	if err == redis.ErrNil {
		err = file.Save(conn)
		return err
	}
	return fmt.Errorf("File already Exist\n")
}

func (file File) check(conn redis.Conn) error {
	if file.Hash != "" &&
		file.DNS != "" &&
		file.Size != 0 {
		return nil
	}
	return fmt.Errorf("Malform filed\n")
}

func (storage Storage) check() error {
	if storage.DNS != "" &&
		storage.ID != 0 &&
		storage.Total != 0 &&
		storage.Used != 0 {
		return nil
	}
	return fmt.Errorf("Malform filed\n")
}

func (storage Storage) Save(conn redis.Conn) error {
	json, err := json.Marshal(storage)
	if err != nil {
		return err
	}
	sKey := strconv.Itoa(int(storage.ID))
	_, err = conn.Do("SET", StoragePrefix + mediate(sKey), json)
	if err != nil {
		return err
	}
	return nil
}

func (file File) Save(conn redis.Conn) error {
	err := file.check(conn)
	if err != nil {
		return err
	}
	json, err := json.Marshal(file)
	if err != nil {
		return err
	}
	_, err = conn.Do("SET", FilePrefix + mediate(file.Hash), json)
	if err != nil {
		return err
	}
	return nil
	return errors.New("Not implemented")
}

func mediate(lama string) string {
	return strings.Replace(lama, " ", "", -1)
}

func (file File) Update(conn redis.Conn) error {
	_, err := GetFile(file.Hash, conn)
	if err == nil {
		return file.Save(conn)
	}
	return err
}

func (storage Storage) Update(conn redis.Conn) error {
	_, err := GetStorage(storage.ID, conn)
	if err == nil {
		return storage.Save(conn)
	}
	return err
}

func (file File) Delete(redis.Conn) error {

	return errors.New("Not implemented")
}

func (storage Storage) Delete(redis.Conn) error {
	return errors.New("Not implemented")
}
