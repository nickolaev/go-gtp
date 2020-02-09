// Copyright 2019-2020 go-gtp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

// Command enb works as pseudo eNB that forwards packets through GTPv1 tunnel.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var configPath = flag.String("config", "./enb.yml", "Path to the configuration file.")
	flag.Parse()
	log.SetPrefix("[eNB] ")

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	enb, err := newENB(cfg)
	if err != nil {
		log.Printf("failed to initialize eNB: %s", err)
	}
	defer enb.close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGHUP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fatalCh := make(chan error, 1)
	go func() {
		if err := enb.run(ctx); err != nil {
			fatalCh <- err
		}
	}()

	for {
		select {
		case sig := <-sigCh:
			switch sig {
			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				fallthrough
			// kill -SIGTERM XXXX
			case syscall.SIGTERM:
				fallthrough
			// kill -SIGQUIT XXXX
			case syscall.SIGQUIT:
				return
			case syscall.SIGHUP:
				// reload config and attach/detach subscribers again
				newCfg, err := loadConfig(*configPath)
				if err != nil {
					log.Printf("Error reloading config %s", err)
				}

				if err := enb.reload(newCfg); err != nil {
					log.Printf("Error applying reloaded config %s", err)
				}
			default:
				fmt.Println("Unknown signal.")
				return
			}
		case err := <-enb.errCh:
			log.Printf("WARN: %s", err)
		case err := <-fatalCh:
			log.Printf("FATAL: %s", err)
			return
		}
	}
}
