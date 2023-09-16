package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/eiannone/keyboard"
	"github.com/fsnotify/fsnotify"
)

var (
	ctx      context.Context
	cancel   context.CancelFunc
	filePath string
)

func init() {
	// Define the path to the JSON file you want to watch
	filePath = "your_file.json"
}

func main() {
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
				fmt.Println("File modified:")
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

func watchCtrlC(sigCh chan os.Signal) {
	<-sigCh
	log.Println("Ctrl+C detected. Exiting...")
	// Cancel the context to signal the less goroutine to terminate
	cancel()
	exec.Command("stty", "sane").Run()
	os.Exit(0)
}

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
