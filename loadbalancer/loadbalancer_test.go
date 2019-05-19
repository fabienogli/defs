package main

import (
	"errors"
	"testing"

	"github.com/gomodule/redigo/redis"
)

//@TODO

//Sauvegarder un hash qui n'existe pas et le get
//Sauvegarder un hash qui existe et erreur
//Erreur full
//get un hash sauvegarder
//Delete un fichier

func TestClient(t *testing.T) {
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()

	pong, err := conn.Do("PING")
	if err != nil {
		t.Errorf("Error getting answer")
	}
	s, err := redis.String(pong, err)
	if err != nil {
		t.Errorf("Error converting answer to string")
	}
	if s != "PONG" {
		t.Errorf("Wrong answer %s\n", s)
	}
}

func TestRegister(t *testing.T) {
	return nil
}

//get the location of the file
func TestWhereIs(t *testing.T) {
	return Storage{}, nil
}

//Get the best Storage for file
func TestWhereTo(t *testing.T) {
	return Storage{}, nil
}

func TestSave(t *testing.T) {
	return errors.New("Not implemented")
}

func TestDelete(t *testing.T) {
	return errors.New("Not implemented")
}
