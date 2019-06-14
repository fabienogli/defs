package database

import (
	"github.com/gomodule/redigo/redis"
	"testing"
)

func check(err error, t *testing.T) {
	if err != nil {
		t.Errorf("There was an error %v\n", err)
	}
}

// Test client connection
func TestClient(t *testing.T) {
	pool := getPool()
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

func getPool() *redis.Pool {
	return NewPool("test_routing_table", 6379)
}