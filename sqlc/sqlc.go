package sqlc

import "github.com/sqlc-dev/pqtype"

func FormatRaw(value []byte) pqtype.NullRawMessage {
	if value == nil {
		return pqtype.NullRawMessage{Valid: false}
	}

	return pqtype.NullRawMessage{
		RawMessage: value,
		Valid:      true,
	}
}

func ParseRaw(value pqtype.NullRawMessage) []byte {
	if !value.Valid {
		return nil
	}

	return value.RawMessage
}
