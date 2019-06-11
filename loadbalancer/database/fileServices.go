package database

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strings"
)

func NewFile(hash string, dns uint, size int) (File, error) {
	if size < 0 {
		return File{}, &ErrorConversion{NegativeInt: size}
	}
	return File{
		Hash: hash,
		DNS: dns,
		Size: uint(size),
	}, nil
}

func (file File) SetExp(duration int, conn redis.Conn) error {
	_, err := conn.Do("EXPIRE", FilePrefix + file.Hash, duration)
	return err
}

func (file File) Persist(conn redis.Conn) error {
	_, err := conn.Do("PERSIST", FilePrefix + file.Hash)
	return err
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

func (file File) Delete(conn redis.Conn) error {
	_, err :=conn.Do("DEL", FilePrefix + file.Hash)
	return err
}


func GetFile(hash string, conn redis.Conn) (File, error) {
	s, err := redis.String(conn.Do("GET", FilePrefix+mediate(hash)))
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

func (file File) Create(conn redis.Conn) error {
	_, err := GetFile(file.Hash, conn)
	if err == redis.ErrNil {
		return file.Save(conn)
	}
	return fmt.Errorf("File already Exist\n")
}

func (file File) check(conn redis.Conn) error {
	if file.Hash != "" &&
		file.DNS != 0 &&
		file.Size != 0 {
		return nil
	}
	return fmt.Errorf("Malform filed %v\n", file)
}

func (file File) updateStorage(conn redis.Conn) error {
	storage, err := GetStorage(file.DNS, conn)
	if err == nil {
		storage.Used += file.Size
		return storage.Update(conn)
	}
	return err
}

func (file File) Save(conn redis.Conn) error {
	err := file.check(conn)
	if err != nil {
		return err
	}
	err = file.updateStorage(conn)
	if err != nil {
		return err
	}
	json, err := json.Marshal(file)
	if err != nil {
		return err
	}
	_, err = conn.Do("SET", FilePrefix+mediate(file.Hash), json)
	if err != nil {
		return err
	}
	return nil
}

