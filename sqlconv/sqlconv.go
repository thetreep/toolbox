package sqlconv

import (
	"database/sql"
)

func ParseInt16(source sql.NullInt16) *int16 {
	if !source.Valid {
		return nil
	}

	return &source.Int16
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

func ParseFloat64(source sql.NullFloat64) *float64 {
	if !source.Valid {
		return nil
	}

	return &source.Float64
}

func FormatFloat64(value *float64) sql.NullFloat64 {
	if value == nil {
		return sql.NullFloat64{Valid: false}
	}

	return sql.NullFloat64{
		Float64: *value,
		Valid:   true,
	}
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

func ParseString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}
