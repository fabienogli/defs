package database

import (
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

const FilePrefix string = "file:"
const StoragePrefix string = "storage:"

type File struct {
	Salt string
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
	Create() error
	Update() error
	Save() error
	Delete() error
}

/*
func main() {
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	err := set(conn)
	if err != nil {
		panic(err)
	}
	err = get(conn)
	if err != nil {
		panic(err)
	}

	err = setStruct(conn)
	if err != nil {
		panic(err)
	}
	err = getStruct(conn)
	if err != nil {
		panic(err)
	}
}
*/

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



//should throw error if there is already a value cause redis overwrite set
func (storage Storage) Create() error {
	return errors.New("Not implemented")
}

func (file File) Create() error {
	return errors.New("Not implemented")
}

func (storage Storage) Save() error {
	return errors.New("Not implemented")
}

func (file File) Save() error {
	return errors.New("Not implemented")
}

func (file File) Update() error {
	return errors.New("Not implemented")
}

func (storage Storage) Update() error {
	return errors.New("Not implemented")
}

func (file File) Delete() error {
	return errors.New("Not implemented")
}

func (storage Storage) Delete() error {
	return errors.New("Not implemented")
}
