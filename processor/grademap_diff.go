package processor

import "fmt"

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

// entityTypeForKey maps known top-level container keys to their entity type.
// Any unknown top-level map key gets entity_type "unknown".
var containerEntityTypes = map[string]string{
	"Characteristics":       "characteristic",
	"Grades":                "grade",
	"Defect_grading_passes": "pass",
}

// DiffGrademaps compares two grademap snapshots represented as
// map[string]any (keyed by field name, values converted to strings) and
// returns a slice of GrademapChange describing every field that was added,
// removed, or modified.  When prev is nil every field in next is treated as
// newly added.
//
// Leaf values are stringified with fmt.Sprintf("%v", v).
// Arrays/slices are stringified as a whole (not recursed into).
// Returns nil (not empty slice) when no changes exist.
func DiffGrademaps(prev, next map[string]any) []GrademapChange {
	var changes []GrademapChange

	for key, nextVal := range next {
		nextMap, nextIsMap := nextVal.(map[string]any)

		if nextIsMap {
			// Nested container: Characteristics, Grades, Defect_grading_passes, or unknown.
			entityType, known := containerEntityTypes[key]
			if !known {
				entityType = "unknown"
			}

			// Get the prev container map (may be nil or absent).
			var prevMap map[string]any
			if prev != nil {
				if pv, ok := prev[key]; ok {
					prevMap, _ = pv.(map[string]any)
				}
			}

			// Walk each entity within the container.
			for entityName, entityVal := range nextMap {
				entityMap, entityIsMap := entityVal.(map[string]any)
				if !entityIsMap {
					// Scalar inside container — treat as a single field change.
					oldStr := ""
					if prevMap != nil {
						if oldVal, ok := prevMap[entityName]; ok {
							oldStr = fmt.Sprintf("%v", oldVal)
						}
					}
					newStr := fmt.Sprintf("%v", entityVal)
					if oldStr != newStr {
						changes = append(changes, GrademapChange{
							EntityType: entityType,
							EntityName: entityName,
							Field:      entityName,
							OldValue:   oldStr,
							NewValue:   newStr,
						})
					}
					continue
				}

				// Entity is itself a map: walk its leaf fields.
				var prevEntityMap map[string]any
				if prevMap != nil {
					if pv, ok := prevMap[entityName]; ok {
						prevEntityMap, _ = pv.(map[string]any)
					}
				}

				for field, newLeaf := range entityMap {
					// Do not recurse into nested maps within an entity — stringify them.
					oldStr := ""
					if prevEntityMap != nil {
						if oldLeaf, ok := prevEntityMap[field]; ok {
							oldStr = fmt.Sprintf("%v", oldLeaf)
						}
					}
					newStr := fmt.Sprintf("%v", newLeaf)
					if oldStr != newStr {
						changes = append(changes, GrademapChange{
							EntityType: entityType,
							EntityName: entityName,
							Field:      field,
							OldValue:   oldStr,
							NewValue:   newStr,
						})
					}
				}
			}
		} else {
			// Top-level scalar (or array/slice — stringified as a whole).
			oldStr := ""
			if prev != nil {
				if oldVal, ok := prev[key]; ok {
					oldStr = fmt.Sprintf("%v", oldVal)
				}
			}
			newStr := fmt.Sprintf("%v", nextVal)
			if oldStr != newStr {
				changes = append(changes, GrademapChange{
					EntityType: "grademap",
					EntityName: "",
					Field:      key,
					OldValue:   oldStr,
					NewValue:   newStr,
				})
			}
		}
	}

	return changes
}
