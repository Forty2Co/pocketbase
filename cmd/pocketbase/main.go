// Package main provides the PocketBase server executable.
//
// This is the main entry point for running a PocketBase server instance
// with custom migrations and configurations.
package main

import (
	"github.com/pocketbase/pocketbase"

	_ "github.com/Forty2Co/pocketbase/migrations"
)

func main() {
	app := pocketbase.New()

	if err := app.Start(); err != nil {
		panic(err)
	}
}
