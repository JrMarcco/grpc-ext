package interceptor

import (
	"context"
	"net"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// Builder provides methods to extract peer information from gRPC context
type Builder struct {
	md metadata.MD
}

// PeerInfo contains information about the peer connection
type PeerInfo struct {
	Name string
	IP   string
}

// PeerInfo returns peer information extracted from the context
func (b *Builder) PeerInfo(ctx context.Context) *PeerInfo {
	if ctx == nil {
		return &PeerInfo{}
	}

	return &PeerInfo{
		Name: b.PeerName(ctx),
		IP:   b.PeerIP(ctx),
	}
}

// PeerName extracts the peer name from the context headers
func (b *Builder) PeerName(ctx context.Context) string {
	return b.grpcHeaderVal(ctx, "x-peer-name")
}

// PeerIP extracts the peer IP from the context headers or connection
func (b *Builder) PeerIP(ctx context.Context) string {
	// First try to get IP from custom header
	if clientIP := b.grpcHeaderVal(ctx, "x-client-ip"); clientIP != "" {
		if net.ParseIP(clientIP) != nil {
			return clientIP
		}
	}

	// Fallback to peer information
	pr, ok := peer.FromContext(ctx)
	if !ok || pr == nil || pr.Addr == nil {
		return ""
	}

	// Extract IP from address
	addr := pr.Addr.String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// If splitting fails, try to parse the entire string as IP
		if net.ParseIP(addr) != nil {
			return addr
		}
		return ""
	}

	return host
}

// grpcHeaderVal extracts a value from gRPC metadata headers
func (b *Builder) grpcHeaderVal(ctx context.Context, key string) string {
	if b.md == nil {
		// Cache metadata once
		var ok bool
		if b.md, ok = metadata.FromIncomingContext(ctx); !ok {
			b.md = metadata.New(map[string]string{})
		}
	}

	if ctx == nil || key == "" {
		return ""
	}

	values := b.md.Get(key)
	if len(values) == 0 {
		return ""
	}

	// Return first value instead of joining all values
	return values[0]
}
