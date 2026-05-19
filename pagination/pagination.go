package pagination

type Cursor string

type PageRequest struct {
	Limit int
	After *Cursor
}

type PageResult[T any] struct {
	Items      []T
	NextCursor *Cursor
	HasMore    bool
}

type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

type Sort struct {
	Field     string
	Direction SortDirection
}
