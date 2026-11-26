package auth

import (
	"context"
)

type Authenticator interface {
	Authenticate(ctx context.Context, credentials any) (*Subject, error)
}
