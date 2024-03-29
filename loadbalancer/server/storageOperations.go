package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"loadbalancer/database"
)

func deleteFile(hash string) Response {
	file, err := database.GetFile(hash, conn)
	if err != nil {
		return UnknownFile
	}
	err = file.Delete(conn)
	if err == nil {
		return OK
	}
	return InternalError
}

func store(hash string, c net.Conn) error {
	file, err := database.GetFile(hash, conn)
	if err != nil {
		return err
	}

	err = file.Persist(conn)

	if err != nil {
		return err
	}

	//Read data
	netData, err := bufio.NewReader(c).ReadString(CmdDelimiter)
	args := strings.Split(netData, ArgsDelimiter)
	if err == nil && len(args) >= 2 {
		if Query(args[0]) == DoneStoring && err == nil {
			return nil
		} else {
			log.Printf("Unknown code")
		}
	} else {
		if err != io.EOF {
			log.Printf("Error while reading file : %v\nArguments: %v", err, args)
		}
	}
	err = file.SetExp(TTL, conn)
	return err
}

func subscribeNew(dns string, used, total uint) Response {
	storage := database.Storage{
		DNS:   dns,
		Used:  used,
		Total: total,
	}
	storage.GenerateUid(conn)
	err := storage.Create(conn)
	if err != nil {
		log.Printf("Error while creating storage: %v\nstorage: %v", err, storage)
		return InternalError
	}
	return Response(fmt.Sprintf("%s%s%d", OK, ArgsDelimiter, storage.ID))
}

func subscribeExisting(id uint, dns string, used, total uint) Response {
	storage := database.Storage{
		ID:    id,
		DNS:   dns,
		Used:  used,
		Total: total,
	}
	dbStorage, err := database.GetStorage(id, conn)
	if err != nil {
		return UnknownStorage
	}
	if storage.Used != dbStorage.Used {
		return NotSameUsedSpace
	}
	err = storage.Update(conn)
	if err != nil {
		return InternalError
	}
	return OK
}

func unsubscribe(id uint) Response {
	storage, err := database.GetStorage(id, conn)
	if err != nil {
		return UnknownStorage
	}
	err = storage.Delete(conn)
	if err == nil {
		return OK
	}
	return InternalError
}
