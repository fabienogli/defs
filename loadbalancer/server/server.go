package server

import "fmt"

//UDP Side

const (
	WhereTo Query = "0"
	WhereIs Query = "1"

	HashAlreadyExisting Response = "1"
	NoStorageLeft       Response = "2"
	HashNotFound        Response = "3"
	MalformRequest      Response = "4"
)
// BOTH
const (
	OK					Response = "0"
	InternalError      Response = "666"
)

//TCP side
const (
	SubscribeNew      Query = "0"
	SubscribeExisting Query = "1"
	Unsub             Query = "2"
	Store             Query = "3"
	Delete            Query = "4"
	DoneStoring       Query = "5"
	TTL               int   = 60

	StorageNonExistent Response = "1"
	NotSameUsedSpace   Response = "2"
	UnknownStorage     Response = "3"
	UnknownFile        Response = "4"

	CmdDelimiter  byte   = '\n'
	ArgsDelimiter string = " "
)

type Query string
type Response string
type ConversionError string
type NotEnoughArgument uint

func (n NotEnoughArgument) Error() string {
	return fmt.Sprintf("Not Enough Argument : %d", n)
}
func (c ConversionError) Error() string {
	return fmt.Sprintf("Error converting %s", c)
}

func (query Query) String() string {
	return string(query)
}

func (response Response) String() string {
	return string(response)
}