package osminokal

import (
	"github.com/google/uuid"
	"log/slog"
)

type Deps struct {
	Logger        *slog.Logger
	NamespaceUUID uuid.UUID
}
