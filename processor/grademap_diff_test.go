package processor

import "testing"

// TestDiffGrademaps is a table-driven test for DiffGrademaps.
//
// The stub implementation returns nil, so all cases that expect changes will
// fail RED.  This is the intended state — Plan 02 implements the real logic.
func TestDiffGrademaps(t *testing.T) {
	cases := []struct {
		name        string
		prev        map[string]any
		next        map[string]any
		wantMinLen  int
		wantField   string
		wantOld     string
		wantNew     string
		wantZeroLen bool
	}{
		{
			name:       "nil prev — all fields in next are new",
			prev:       nil,
			next:       map[string]any{"Name": "GM1"},
			wantMinLen: 1,
			wantField:  "Name",
			wantOld:    "",
			wantNew:    "GM1",
		},
		{
			name:       "Name changed — one change returned",
			prev:       map[string]any{"Name": "GM1"},
			next:       map[string]any{"Name": "GM2"},
			wantMinLen: 1,
			wantField:  "Name",
			wantOld:    "GM1",
			wantNew:    "GM2",
		},
		{
			name:        "nothing changed — empty slice",
			prev:        map[string]any{"Name": "GM1"},
			next:        map[string]any{"Name": "GM1"},
			wantZeroLen: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DiffGrademaps(tc.prev, tc.next)

			if tc.wantZeroLen {
				if len(got) != 0 {
					t.Errorf("DiffGrademaps: got %d changes, want 0", len(got))
				}
				return
			}

			if len(got) < tc.wantMinLen {
				t.Fatalf("DiffGrademaps: got %d changes, want at least %d", len(got), tc.wantMinLen)
			}

			// Find the expected change in the result.
			found := false
			for _, c := range got {
				if c.Field == tc.wantField && c.OldValue == tc.wantOld && c.NewValue == tc.wantNew {
					if c.EntityType == "" {
						t.Errorf("change for Field=%q has empty EntityType", tc.wantField)
					}
					found = true
					break
				}
			}
			if !found {
				t.Errorf("DiffGrademaps: no change with Field=%q OldValue=%q NewValue=%q in %+v",
					tc.wantField, tc.wantOld, tc.wantNew, got)
			}
		})
	}
}
