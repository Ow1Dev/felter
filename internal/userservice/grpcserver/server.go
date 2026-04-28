// Package grpcserver implements the userservice gRPC interface.
package grpcserver

import (
	"context"
	"fmt"
	"time"

	"github.com/Ow1Dev/felter/internal/userservice/pb"
	"github.com/Ow1Dev/felter/internal/userservice/store"
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
