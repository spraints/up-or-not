package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	fs := flag.NewFlagSet("upornot", flag.ExitOnError)

	targetIP := "127.0.0.1"
	fs.StringVar(&targetIP, "target", targetIP, "IP address to ping")

	interval := time.Second
	fs.DurationVar(&interval, "interval", interval, "interval between ping attempts")

	address := "127.0.0.1:4444"
	fs.StringVar(&address, "addr", address, "server address")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for sig := range sigs {
			log.Printf("received %v, exiting", sig)
			cancel()
		}
	}()

	var wg sync.WaitGroup
	m := &model{
		TargetIP: targetIP,
		Interval: interval,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logerr(serveHTTP(ctx, &http.Server{
			Addr:    address,
			Handler: buildHTTPHandler(m),
		}), "http server")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logerr(ping(ctx, m), "ping")
	}()

	wg.Wait()
}

func logerr(err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}
