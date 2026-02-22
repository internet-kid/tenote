package fs

import "time"

// ---------------------------------------------------------------------------
// Section
// ---------------------------------------------------------------------------

type Section string

const (
	SectionNotes Section = "notes"
	SectionTrash Section = "trash"
)

// ---------------------------------------------------------------------------
// Note
// ---------------------------------------------------------------------------

type Note struct {
	ID        string
	Title     string
	Path      string
	Section   Section
	UpdatedAt time.Time
}
