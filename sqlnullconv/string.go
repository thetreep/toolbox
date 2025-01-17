package nullconv

import (
	"database/sql"
)

func ParseInt16(source sql.NullInt16) *int16 {
	var result *int16
	if source.Valid {
		result = &source.Int16
	}

	return result
}

func ParseFloat64(source sql.NullFloat64) *float64 {
	var result *float64
	if source.Valid {
		result = &source.Float64
	}

	return result
}

func FormatString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{
		String: *value,
		Valid:  true,
	}
}

func FormatInt16(value *int16) sql.NullInt16 {
	if value == nil {
		return sql.NullInt16{Valid: false}
	}

	return sql.NullInt16{
		Int16: *value,
		Valid: true,
	}
}
