package database

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
	"os"
)

const (
	FilePrefix    = "file:"
	StoragePrefix = "storage:"
)

type File struct {
	Hash string `json:"hash"`
	DNS  uint   `json:"dns"`
	Size uint   `json:"size"`
}

type Storage struct {
	ID    uint   `json:"id"`
	DNS   string `json:"dns"`
	Used  uint   `json:"used"`
	Total uint   `json:"total"`
}

type Cud interface {
	Create(con redis.Conn) error
	Update(con redis.Conn) error
	Save(conn redis.Conn) error
	Delete(conn redis.Conn) error
}

//Negative int that could have produced integer overflow
type ErrorConversion struct {
	NegativeInt int
}

func (err ErrorConversion) Error() string {
	return fmt.Sprintf("Negative number: %d", err.NegativeInt)
}

func GetDatabase() *redis.Pool{
	addr := os.Getenv("ROUTING_DB_ADDR")
	sport := os.Getenv("ROUTING_DB_PORT")
	port,_  := strconv.Atoi(sport)
	return NewPool(addr, port)
}

func NewPool(address string, port int) *redis.Pool {
	address += ":" + strconv.Itoa(port)
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
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