package main

import (
	"log"
	"loadbalancer/database"
	"errors"
)

func main() {
	log.Println("load balancer running")
}

//get the location of the file
func whereIs(salt string) (database.Storage, error) {
	return database.Storage{}, errors.New("Not implemented")
}

//Get the best Storage for file
func whereTo(salt string, size int) (database.Storage, error) {
	return database.Storage{}, errors.New("Not implemented")
}


func delete(salt string) error {
	return errors.New("Not implemented")
}

func store(salt string) error {
	return errors.New("Not implemented")
}

//send id back
func subscribe(dns string, used, total uint) (uint, error) {
	return 0, errors.New("Not implemented")
}

func unsubscribe(id uint) error {
	return errors.New("Not implemented")
}