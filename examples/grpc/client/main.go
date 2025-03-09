package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/mstgnz/self-hosted-serverless/internal/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Parse command line arguments
	functionName := flag.String("function", "", "Function name to execute")
	inputJSON := flag.String("input", "{}", "Input JSON")
	serverAddr := flag.String("server", "localhost:9090", "Server address")
	flag.Parse()

	if *functionName == "" {
		log.Fatal("Function name is required")
	}

	// Connect to the gRPC server
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client
	client := pb.NewFunctionServiceClient(conn)

	// Parse input JSON
	var inputMap map[string]interface{}
	if err := json.Unmarshal([]byte(*inputJSON), &inputMap); err != nil {
		log.Fatalf("Failed to parse input JSON: %v", err)
	}

	// Convert input to string map for gRPC
	input := make(map[string]string)
	for k, v := range inputMap {
		// Convert values to strings
		valueBytes, err := json.Marshal(v)
		if err != nil {
			log.Fatalf("Failed to marshal value: %v", err)
		}
		input[k] = string(valueBytes)
	}

	// Execute the function
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Printf("Executing function %s with input %v\n", *functionName, input)

	resp, err := client.ExecuteFunction(ctx, &pb.ExecuteFunctionRequest{
		Name:  *functionName,
		Input: input,
	})

	if err != nil {
		log.Fatalf("Error executing function: %v", err)
	}

	// Print the result
	fmt.Println("Result:")
	for k, v := range resp.Result {
		fmt.Printf("  %s: %s\n", k, v)
	}

	// List available functions
	listResp, err := client.ListFunctions(ctx, &pb.ListFunctionsRequest{})
	if err != nil {
		log.Fatalf("Error listing functions: %v", err)
	}

	fmt.Println("\nAvailable Functions:")
	for _, f := range listResp.Functions {
		fmt.Printf("  %s: %s (%s)\n", f.Name, f.Description, f.Runtime)
	}
}
