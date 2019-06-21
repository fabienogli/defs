package database

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

func NewStorage(dns string, used, total int) (Storage, error) {
	if used < 0 {
		return Storage{}, &ErrorConversion{NegativeInt: used}
	}
	if total < 0 {
		return Storage{}, &ErrorConversion{NegativeInt: total}
	}
	return Storage{
		DNS: dns,
		Used: uint(used),
		Total: uint(total),
	}, nil
}

func (storage *Storage)  GenerateUid(conn redis.Conn) {
	tmp, _ := redis.Strings(conn.Do("KEYS", StoragePrefix+"*"))
	storage.ID = uint(len(tmp) + 1)
}

func (storage Storage) GetAvailableSpace() uint {
	return storage.Total - storage.Used
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
	s, err := redis.String(conn.Do("GET", StoragePrefix+mediate(sKey)))
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

//should throw error if there is already a value cause redis overwrite set
func (storage Storage) Create(conn redis.Conn) error {
	_, err := GetStorage(storage.ID, conn)
	if err == redis.ErrNil {
		return storage.Save(conn)
	}
	return fmt.Errorf("Storage already Exist\n")
}


func (storage Storage) check() error {
	if storage.DNS != "" &&
		storage.ID != 0 &&
		storage.Total != 0 &&
		storage.Used < storage.Total {
		return nil
	}
	return fmt.Errorf("Malform storage\n")
}

func (storage Storage) Save(conn redis.Conn) error {
	err := storage.check()
	if err != nil {
		return err
	}
	json, err := json.Marshal(storage)
	if err != nil {
		return err
	}
	sKey := strconv.Itoa(int(storage.ID))
	_, err = conn.Do("SET", StoragePrefix+mediate(sKey), json)
	if err != nil {
		return err
	}
	return nil
}

func (storage Storage) Update(conn redis.Conn) error {
	_, err := GetStorage(storage.ID, conn)
	if err == nil {
		return storage.Save(conn)
	}
	return err
}

func (storage Storage) Delete(conn redis.Conn) error {
	sKey := strconv.Itoa(int(storage.ID))
	_, err :=conn.Do("DEL", StoragePrefix + sKey)
	return err
}