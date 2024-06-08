package utils

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func UuidToString(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}

func StringToUuid(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if s != "" && err != nil {
		log.Warn().Err(err).
			Str("UUID", s).
			Msg("Failed to parse UUID")
		return uuid.Nil
	}
	return id
}
