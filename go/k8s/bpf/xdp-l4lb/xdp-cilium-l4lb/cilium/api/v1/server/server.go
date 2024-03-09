package server

import (
	"context"
)

var (
	// ServerCtx and ServerCancel
	ServerCtx, serverCancel = context.WithCancel(context.Background())
)
