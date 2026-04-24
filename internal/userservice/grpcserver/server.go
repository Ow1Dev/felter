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
