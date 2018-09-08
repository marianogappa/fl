package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func serve(server *http.Server) {
	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	fmt.Printf("server: serving on %v\n", server.Addr)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	fmt.Printf("server: shutting down\n")
}
