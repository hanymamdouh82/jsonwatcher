// Copyright 2023 Hany Mamdouh. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/fsnotify/fsnotify"
)

var (
	ctx        context.Context
	cancel     context.CancelFunc
	filePath   string
	predefFile string
)

func init() {
	// Define the path to the JSON file you want to watch
	predefFile = "your_file.json"
}

func main() {
	// Set file to watch
	flagFile := flag.String("f", predefFile, "file name to watch. Include path if not in current directory.")
	flag.Parse()
	filePath = *flagFile
	fmt.Printf("Watching: %s\n", filePath)

	// Handle Ctrl+C to gracefully exit the application
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	go watchCtrlC(sigCh)

	// Create a new filesystem watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start watching the JSON file for changes
	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Create a context with cancellation
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to receive file change events
	events := make(chan fsnotify.Event)

	// Start a goroutine to listen for file change events
	go fileWatcher(watcher, events)

	// Start listening for keyboard input
	err = keyboard.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()
	go watchForKeyExit(cancel)

	// Loop to handle file change events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Printf("Last File modified: %s\n", time.Now())
				fmt.Println("Press: 'q' to quit")
				// Cancel the context to signal the less goroutine to terminate
				cancel()

				// Start a new less process with a new context
				ctx, cancel = context.WithCancel(context.Background())
				go startLess(ctx, filePath)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

// Watch for ctrl+c to gracefully terminate
func watchCtrlC(sigCh chan os.Signal) {
	<-sigCh
	log.Println("Ctrl+C detected. Exiting...")
	// Cancel the context to signal the less goroutine to terminate
	cancel()
	exec.Command("reset").Run()
	os.Exit(0)
}

// Start new `less` instance for the watched file.
func startLess(ctx context.Context, filePath string) {
	lessCmd := exec.CommandContext(ctx, "less", filePath)
	lessCmd.Stdout = os.Stdout
	lessCmd.Stdin = os.Stdin
	lessCmd.Stderr = os.Stderr
	err := lessCmd.Run()
	if err != nil {
		log.Println("Error running less:", err)
	}
}

// File watcher infinite loop.
func fileWatcher(watcher *fsnotify.Watcher, events chan fsnotify.Event) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Check if the event is a modification of the JSON file
			if event.Op&fsnotify.Write == fsnotify.Write {
				events <- event
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

// Watch for `q` key to clean exit.
func watchForKeyExit(cancel context.CancelFunc) {
	for {
		char, _, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}

		if char == 'q' || char == 'Q' {
			log.Println("Q key detected. Exiting...")
			cancel()
			exec.Command("reset").Run()
			os.Exit(0)
		}
	}
}
