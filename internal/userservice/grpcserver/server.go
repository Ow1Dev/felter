// Package grpcserver implements the userservice gRPC interface.
package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ow1Dev/felter/internal/log"
	"github.com/Ow1Dev/felter/internal/userservice/pb"
	"github.com/Ow1Dev/felter/internal/userservice/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Server implements the generated gRPC interface.
type Server struct {
	pb.UnimplementedUserServiceServer
	store store.Store
}

// NewServer returns a gRPC server backed by store.
func NewServer(s store.Store) *Server {
	return &Server{store: s}
}

// UnaryInterceptor extracts correlation ID from incoming metadata, injects it into context,
// and logs the request with method, correlation ID, and duration.
func UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if id := md.Get("x-correlation-id"); len(id) > 0 {
			ctx = log.SetCorrelationID(ctx, id[0])
		}
	}

	corr := log.CorrelationID(ctx)
	resp, err := handler(ctx, req)

	dur := time.Since(start)
	if err != nil {
		slog.Default().Error("grpc request",
			slog.String("corr", corr),
			slog.String("method", info.FullMethod),
			slog.Int64("dur_ms", dur.Milliseconds()),
			slog.String("err", err.Error()),
		)
	} else {
		slog.Default().Info("grpc request",
			slog.String("corr", corr),
			slog.String("method", info.FullMethod),
			slog.Int64("dur_ms", dur.Milliseconds()),
		)
	}

	return resp, err
}

// CreateUser handles the CreateUser RPC.
func (srv *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	if req.Email == "" || req.Username == "" {
		return nil, fmt.Errorf("email and username are required")
	}
	u, err := srv.store.CreateUser(ctx, req.Email, req.Username, req.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("store create: %w", err)
	}
	return toProto(u), nil
}

// GetUser handles the GetUser RPC.
func (srv *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	if req.Id == 0 {
		return nil, fmt.Errorf("id is required")
	}
	u, err := srv.store.GetUser(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("store get user: %w", err)
	}
	return toProto(u), nil
}

// GetUserFromProvider handles the GetUserFromProvider RPC.
func (srv *Server) GetUserFromProvider(ctx context.Context, req *pb.GetUserFromProviderRequest) (*pb.User, error) {
	if req.Provider == "" || req.ProviderId == "" {
		return nil, fmt.Errorf("provider and provider_id are required")
	}
	u, err := srv.store.GetUserFromProvider(ctx, req.Provider, req.ProviderId)
	if err != nil {
		return nil, fmt.Errorf("store get user from provider: %w", err)
	}
	return toProto(u), nil
}

// CreateUserFromProvider handles the CreateUserFromProvider RPC.
func (srv *Server) CreateUserFromProvider(ctx context.Context, req *pb.CreateUserFromProviderRequest) (*pb.User, error) {
	if req.Provider == "" || req.ProviderId == "" || req.Email == "" || req.Username == "" {
		return nil, fmt.Errorf("provider, provider_id, email, and username are required")
	}
	u, err := srv.store.CreateUserFromProvider(ctx, req.Provider, req.ProviderId, req.Email, req.Username)
	if err != nil {
		return nil, fmt.Errorf("store create user from provider: %w", err)
	}
	return toProto(u), nil
}

func toProto(u *store.User) *pb.User {
	p := &pb.User{
		Id:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
	if u.DisplayName != nil {
		p.DisplayName = *u.DisplayName
	}
	return p
}
