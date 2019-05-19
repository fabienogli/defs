package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
)

const filePrefix string = "file:"

type File struct {
	Salt string
	DNS  string
}

type Storage struct {
	DNS   string
	Used  string
	Total string
}

func setStruct(c redis.Conn) error {
	fl := File{
		Salt: "123",
		DNS:  "ici",
	}

	json, err := json.Marshal(fl)
	if err != nil {
		return err
	}

	_, err = c.Do("SET", filePrefix+fl.Salt, json)
	if err != nil {
		return err
	}
	return nil
}

func getStruct(c redis.Conn) error {
	salt := "123"
	s, err := redis.String(c.Do("GET", filePrefix+salt))
	if err == redis.ErrNil {
		fmt.Println("File does not exist")
	} else if err != nil {
		return err
	}
	fl := File{}
	err = json.Unmarshal([]byte(s), &fl)

	log.Printf("%+v\n", fl)
	return nil
}

func check(err error) {
	if err != nil {
		log.Println(err)
	}
}

func main() {
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	err := set(conn)
	check(err)
	err = get(conn)
	check(err)

	err = setStruct(conn)
	check(err)
	err = getStruct(conn)
	check(err)
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

func set(c redis.Conn) error {
	_, err := c.Do("SET", "Favorite Movie", "Repo Man")
	if err != nil {
		return err
	}
	_, err = c.Do("SET", "Release Year", 1984)
	if err != nil {
		return err
	}
	return nil
}

// get executes the redis GET command
func get(c redis.Conn) error {

	// Simple GET example with String helper
	key := "Favorite Movie"
	s, err := redis.String(c.Do("GET", key))
	if err != nil {
		return (err)
	}
	fmt.Printf("%s = %s\n", key, s)

	// Simple GET example with Int helper
	key = "Release Year"
	i, err := redis.Int(c.Do("GET", key))
	if err != nil {
		return (err)
	}
	fmt.Printf("%s = %d\n", key, i)

	// Example where GET returns no results
	key = "Nonexistent Key"
	s, err = redis.String(c.Do("GET", key))
	if err == redis.ErrNil {
		fmt.Printf("%s does not exist\n", key)
	} else if err != nil {
		return err
	} else {
		fmt.Printf("%s = %s\n", key, s)
	}

	return nil
}

func Register(storage Storage) error {
	return nil
}

//get the location of the file
func WhereIs(hash string) (Storage, error) {
	return Storage{}, nil
}

//Get the best Storage for file
func WhereTo(hash string, size int) (Storage, error) {
	return Storage{}, nil
}

func Save(file File) error {
	return errors.New("Not implemented")
}

func Delete(file File) error {
	return errors.New("Not implemented")
}
