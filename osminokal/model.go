package osminokal

// Data model

import (
	"github.com/google/uuid"
	"time"
)

type FreeEnergySession struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	ID    uuid.UUID
}
