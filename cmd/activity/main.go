package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"activityregistration/internal/clock"
	"activityregistration/internal/httpapi"
	"activityregistration/internal/service"
	"activityregistration/internal/store"
)

func main() {
	databasePath := flag.String("db", "activity.db", "embedded database path")
	address := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()
	if err := run(*databasePath, *address); err != nil {
		panic(err)
	}
}

func run(databasePath, address string) error {
	database, err := store.Open(databasePath)
	if err != nil {
		return err
	}
	defer database.Close()
	current := clock.NewFixed(time.Date(2025, time.January, 1, 9, 0, 0, 0, time.UTC))
	application := service.New(database, current)
	server := &http.Server{Addr: address, Handler: httpapi.NewHandler(application), ReadHeaderTimeout: 5 * time.Second}
	fmt.Printf("activity registration listening on %s\n", address)
	return server.ListenAndServe()
}
