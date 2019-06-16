package main

import (
	"context"
	"flag"
	"fmt"
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
	fs.Usage = func() { fmt.Printf("Usage: up-or-not [OPTIONS] TARGET....") }

	interval := time.Second
	fs.DurationVar(&interval, "interval", interval, "interval between ping attempts")

	address := "127.0.0.1:4444"
	fs.StringVar(&address, "addr", address, "server address")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	if fs.NArg() == 0 {
		fs.Usage()
		os.Exit(1)
	}

	targetIPs := fs.Args()

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

	models := make([]*model, 0, len(targetIPs))
	for _, targetIP := range targetIPs {
		m := &model{
			TargetIP: targetIP,
			Interval: interval,
		}
		models = append(models, m)

		wg.Add(1)
		go func() {
			defer wg.Done()
			logerr(ping(ctx, m), "ping")
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logerr(serveHTTP(ctx, &http.Server{
			Addr:    address,
			Handler: buildHTTPHandler(models),
		}), "http server")
	}()

	wg.Wait()
}

func logerr(err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}
