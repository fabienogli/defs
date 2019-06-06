package main

import "fmt"

type HashInvalid error
type HashTooLarge error
type HashTooShort error
type HashNotFound error
type BadRequest error
type InternalError error
type FileTooLarge error
type CannotCreateFile error

func NewHashInvalid() HashInvalid {
	return HashInvalid(fmt.Errorf("hash is not valid"))
}

func NewHashTooLarge() HashTooLarge {
	return HashTooLarge(fmt.Errorf("hash is too large"))
}

func NewHashNotFound() HashNotFound {
	return HashNotFound(fmt.Errorf("hash must be specified"))
}

func NewHashTooShort() HashTooShort {
	return HashTooShort(fmt.Errorf("hash is too short"))
}

func NewBadRequest(msg string) BadRequest {
	return BadRequest(fmt.Errorf(msg))
}

func NewInternalError(msg string) InternalError {
	return InternalError(fmt.Errorf(msg))
}

func NewCannotCreateFile(err error) CannotCreateFile {
	return CannotCreateFile(err)
}

func NewFileTooLarge(err error) FileTooLarge {
	return FileTooLarge(err)
}