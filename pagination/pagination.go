package pagination

type (
	// Cursor is an opaque token used to continue pagination from a specific position.
	Cursor string

	// CursorCodec encodes and decodes pagination cursors for a given type.
	CursorCodec[T any] interface {
		Encode(T) (Cursor, error)
		Decode(Cursor) (T, error)
	}

	// PageResult represents a paginated collection of items.
	PageResult[T any] struct {
		Items      []T
		NextCursor Cursor
		HasMore    bool
	}

	PageRequest struct {
		Limit int
		After Cursor
	}

	SortDirection string

	Sort struct {
		Field     string
		Direction SortDirection
	}
)

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)
