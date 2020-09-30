package binary

import "time"

// Metadata holds build metadatas.
type Metadata struct {
	Branch     string    `json:"branch"`
	Compiler   string    `json:"compiler"`
	CompiledAt string    `json:"compiledAt"`
	Sha        string    `json:"sha"`
	StartTime  time.Time `json:"started_at"`
}

// ToMap returns a map from given metadata.
//nolint
func (m Metadata) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Branch":     m.Branch,
		"Compiler":   m.Compiler,
		"CompiledAt": m.CompiledAt,
		"Sha":        m.Sha,
		"StartAt":    m.StartTime,
	}
}

var binaryMetadata = Metadata{ // nolint
	StartTime: time.Now(),
}

// BinaryMetadata return metadata on the given binary.
func BinaryMetadata() Metadata { // nolint
	return binaryMetadata
}

// SetBinaryMetadata sets metadata.
func SetBinaryMetadata(m Metadata) { // nolint
	binaryMetadata = m
}
