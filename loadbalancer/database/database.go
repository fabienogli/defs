package database

import (
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
	Hash string
	DNS  string
	Size uint
}

type Storage struct {
	ID		uint
	DNS   string
	Used  uint
	Total uint
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

	return Storage{}, errors.New("Not implemented")
}

func GetFile(hash string, conn redis.Conn) (File, error) {
	return File{}, errors.New("Not implemented")
}

//should throw error if there is already a value cause redis overwrite set
func (storage Storage) Create(redis.Conn) error {
	return errors.New("Not implemented")
}

func (file File) Create(redis.Conn) error {
	return errors.New("Not implemented")
}

func (storage Storage) Save(redis.Conn) error {
	return errors.New("Not implemented")
}

func (file File) Save(redis.Conn) error {
	return errors.New("Not implemented")
}

func (file File) Update(redis.Conn) error {
	return errors.New("Not implemented")
}

func (storage Storage) Update(redis.Conn) error {
	return errors.New("Not implemented")
}

func (file File) Delete(redis.Conn) error {
	return errors.New("Not implemented")
}

func (storage Storage) Delete(redis.Conn) error {
	return errors.New("Not implemented")
}
