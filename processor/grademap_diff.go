package processor

// GrademapChange describes a single field-level change between two grademap
// snapshots.  EntityType is always "grademap" for top-level fields; EntityName
// identifies which grademap was affected.
type GrademapChange struct {
	EntityType string
	EntityName string
	Field      string
	OldValue   string
	NewValue   string
}

// DiffGrademaps compares two grademap snapshots represented as
// map[string]any (keyed by field name, values converted to strings) and
// returns a slice of GrademapChange describing every field that was added,
// removed, or modified.  When prev is nil every field in next is treated as
// newly added.
//
// Stub — returns nil.  Full implementation added by Plan 02.
func DiffGrademaps(prev, next map[string]any) []GrademapChange {
	return nil
}
