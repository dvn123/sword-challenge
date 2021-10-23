package main

import (
	_ "github.com/go-sql-driver/mysql"
	"sword-challenge/internal"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//todo move os get env here
func main() {
	internal.StartServer()
}
