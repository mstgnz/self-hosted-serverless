package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/mstgnz/self-hosted-serverless/internal/function"
	pb "github.com/mstgnz/self-hosted-serverless/internal/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Service represents the gRPC service
type Service struct {
	pb.UnimplementedFunctionServiceServer
	server   *grpc.Server
	registry *function.Registry
}

// NewService creates a new gRPC service
func NewService(registry *function.Registry) *Service {
	return &Service{
		registry: registry,
	}
}

// Start starts the gRPC server
func (s *Service) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer()
	pb.RegisterFunctionServiceServer(s.server, s)
	reflection.Register(s.server)

	log.Printf("Starting gRPC server on port %d...\n", port)
	return s.server.Serve(lis)
}

// Stop stops the gRPC server
func (s *Service) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// ExecuteFunction executes a function
func (s *Service) ExecuteFunction(ctx context.Context, req *pb.ExecuteFunctionRequest) (*pb.ExecuteFunctionResponse, error) {
	// Convert request to map
	input := make(map[string]any)
	for k, v := range req.Input {
		input[k] = v
	}

	// Execute the function
	result, err := s.registry.Execute(req.Name, input)
	if err != nil {
		return nil, fmt.Errorf("failed to execute function: %w", err)
	}

	// Convert result to response
	response := &pb.ExecuteFunctionResponse{
		Success: true,
	}

	// Handle different result types
	switch r := result.(type) {
	case map[string]any:
		response.Result = make(map[string]string)
		for k, v := range r {
			response.Result[k] = fmt.Sprintf("%v", v)
		}
	default:
		response.Result = map[string]string{
			"result": fmt.Sprintf("%v", result),
		}
	}

	return response, nil
}

// ListFunctions lists all available functions
func (s *Service) ListFunctions(ctx context.Context, req *pb.ListFunctionsRequest) (*pb.ListFunctionsResponse, error) {
	functions := s.registry.ListFunctions()

	response := &pb.ListFunctionsResponse{
		Functions: make([]*pb.FunctionInfo, 0, len(functions)),
	}

	for _, f := range functions {
		response.Functions = append(response.Functions, &pb.FunctionInfo{
			Name:        f.Name,
			Description: f.Description,
			Runtime:     f.Runtime,
		})
	}

	return response, nil
}
