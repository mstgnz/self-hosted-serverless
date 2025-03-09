package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mstgnz/self-hosted-serverless/internal/cli"
	"github.com/mstgnz/self-hosted-serverless/internal/function"
	"github.com/mstgnz/self-hosted-serverless/internal/grpc"
	"github.com/mstgnz/self-hosted-serverless/internal/server"
)

func main() {
	// Parse command line arguments
	var port int
	var grpcPort int
	var command string

	flag.IntVar(&port, "port", 8080, "Port to run the HTTP server on")
	flag.IntVar(&grpcPort, "grpc-port", 9090, "Port to run the gRPC server on")
	flag.Parse()

	// Get the command from arguments
	args := flag.Args()
	if len(args) > 0 {
		command = args[0]
	}

	// Handle CLI commands
	switch command {
	case "create":
		if len(args) < 3 {
			fmt.Println("Usage: go-serverless create function <function-name>")
			os.Exit(1)
		}
		if args[1] == "function" {
			cli.CreateFunction(args[2])
		}
	case "run":
		if len(args) < 2 {
			fmt.Println("Usage: go-serverless run <function-name>")
			os.Exit(1)
		}
		cli.RunFunction(args[1])
	case "list":
		cli.ListFunctions()
	case "metrics":
		if len(args) > 1 {
			cli.GetFunctionMetrics(args[1])
		} else {
			cli.GetMetrics()
		}
	case "":
		// Start the server if no command is provided
		registry := function.NewRegistry()

		// Start HTTP server
		srv := server.NewServer(port, registry)
		go func() {
			log.Printf("Starting HTTP server on port %d...\n", port)
			if err := srv.Start(); err != nil {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}()

		// Start gRPC server
		grpcSrv := grpc.NewService(registry)
		go func() {
			log.Printf("Starting gRPC server on port %d...\n", grpcPort)
			if err := grpcSrv.Start(grpcPort); err != nil {
				log.Fatalf("Failed to start gRPC server: %v", err)
			}
		}()

		// Wait for interrupt signal
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		// Gracefully shutdown servers
		log.Println("Shutting down servers...")
		srv.Stop()
		grpcSrv.Stop()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: create, run, list, metrics")
		os.Exit(1)
	}
}
