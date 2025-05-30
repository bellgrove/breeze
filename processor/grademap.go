package processor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Grademap struct {
	Name            string
	Characteristics map[int]GmCh
	Passes          []GmPa
	Grades          []Grade
}

type Grade struct {
	Name     string
	Grade    string
	Criteria []GmGrCs
}

type GmCh struct {
	Name                      string
	Category                  string
	Defect_grading_pass_index int
	Display_order             int
	Index                     int
	Mode                      string
	Scale                     []int
	Min_value                 []int
	Power                     int
}

type GmPa struct {
	Name  string
	Index int
}

type GmGr struct {
	Name          string
	Criteria_sets map[string]GmGrCs
}

type GmGrCs struct {
	Name        string
	Code        string
	Criteria    []GmGrCsCr
	Not_a_fruit bool
}

type GmGrCsCr struct {
	Characteristic_index int
	Characteristic       string
	Category             string
	Thresholds           []struct {
		Max float64
		Min float64
	}
}

func (c *GmGrCsCr) Check(f *Fruit) bool {
	var val any
	switch c.Category {
	case "Classified Blobs":
		val = f.MClassifiedBlob[c.Characteristic]
	case "Fruit Color":
		val = f.MColour[c.Characteristic]
	case "Special":
	case "Shape":
		// Slippage vs Slippage Factor
		val = f.MFeatures[c.Characteristic]
	case "Defect Color":
		val = f.MColourBlob[c.Characteristic]
	case "Color Combination":
		var total float64
		matched := false
		parts := strings.Split(c.Characteristic, " + ")
		for _, v := range parts {
			if v2 := f.MColour[v]; v2 != nil {
				total += v2.(float64)
				matched = true
			}
		}
		if matched {
			val = total
		}
	case "Size Criteria":
		val = f.MDiameters[c.Characteristic]
	case "Color Function":
		// Softness 2 (v2) (Dark Red +) vs Softness 2 (v2)
		val = f.MFunction[c.Characteristic]
	default:
		slog.Debug("unknown", "cat", c.Category, "cr", c)
		return false
	}

	if val == nil {
		slog.Debug("not detected", "cat", c.Category, "cr", c)
		return false
	}

	// i_val := int(math.Round(val.(float64)))
	i_val := val.(float64)
	slog.Debug("matched", "value", val, "crit", c)

	for _, v := range c.Thresholds {
		if i_val >= v.Min && i_val <= v.Max {
			slog.Debug("triggered", "value", val, "crit", c)
			return true
		}
	}

	return false
}

func (g *Grademap) Grade(f *Fruit) {
	// A -> 3
	// B -> 2
	// len(grades) = 5
	idx := -1
	if len(f.VisionGrade) >= 1 {
		idx = len(g.Grades) - int(f.VisionGrade[0]-'A') - 2
	}
	if idx < 0 || idx >= len(g.Grades)-1 {
		slog.Error("invalid fruit grade", "vision_grade", f.VisionGrade)
		return
	}

	codes := make([]string, 0, 8)
	reasons := make([]string, 0, 8)

	for _, v := range g.Grades[idx].Criteria {
		for _, v2 := range v.Criteria {
			if v2.Check(f) {
				reasons = append(reasons, v.Name)
				if !slices.Contains(codes, v.Code) {
					codes = append(codes, v.Code)
				}
				break
			}
		}
	}
	if len(codes) == 0 {
		return
	}
	slog.Debug("graded", "codes", codes, "reasons", reasons)

	f.PrimaryDefect = codes[0]
	f.OtherDefects = codes[1:]
	f.PrimaryReason = reasons[0]
	f.OtherReason = reasons[1:]
}

func (g *Grademap) Update(b []byte) error {
	var gm struct {
		Characteristics       map[string]GmCh
		Defect_grading_passes map[string]GmPa
		Grades                map[string]GmGr
		Id                    int
		Is_primary            bool
		Name                  string
		Node_id               int
	}

	if err := json.Unmarshal(b, &gm); err != nil {
		return fmt.Errorf("unable to parse grademap: %w", err)
	}

	g.Characteristics = make(map[int]GmCh, len(gm.Characteristics))
	for k1, v1 := range gm.Characteristics {
		v1.Name = k1
		g.Characteristics[v1.Index] = v1
	}
	g.Passes = make([]GmPa, len(gm.Defect_grading_passes))
	for k1, v1 := range gm.Defect_grading_passes {
		v1.Name = k1
		g.Passes[v1.Index] = v1
		// gm.Defect_grading_passes[k1] = v1
	}

	g.Grades = make([]Grade, len(gm.Grades))
	grade_names := slices.Sorted(maps.Keys(gm.Grades))
	re := regexp.MustCompile(`\[([^\]]*)\]`)
	for i1, v1 := range grade_names {
		grade, name, _ := strings.Cut(v1, ". ")
		g.Grades[i1] = Grade{name, grade, make([]GmGrCs, len(gm.Grades[v1].Criteria_sets))}

		criteria_names := slices.SortedFunc(maps.Keys(gm.Grades[v1].Criteria_sets), cmp_crit)

		for i2, v2 := range criteria_names {
			g.Grades[i1].Criteria[i2] = gm.Grades[v1].Criteria_sets[v2]
			g.Grades[i1].Criteria[i2].Name = v2
			code_re := re.FindStringSubmatch(v2)
			if len(code_re) > 1 {
				g.Grades[i1].Criteria[i2].Code = code_re[1]
			} else {
				g.Grades[i1].Criteria[i2].Code = "UNK"
			}

			for i3, v3 := range g.Grades[i1].Criteria[i2].Criteria {
				g.Grades[i1].Criteria[i2].Criteria[i3].Characteristic = g.Characteristics[v3.Characteristic_index].Name
				g.Grades[i1].Criteria[i2].Criteria[i3].Category = g.Characteristics[v3.Characteristic_index].Category
			}
		}
	}

	g.Name = gm.Name

	return nil
}

// cmp(a, b) should return a negative number when a < b, a positive number when
// a > b and zero when a == b or a and b are incomparable in the sense of
// a strict weak ordering.
func cmp_crit(a string, b string) int {
	a_i, err := crit_to_i(a)
	if err != nil {
		slog.Error("unable to compare", "err", err)
		return 0
	}
	b_i, err := crit_to_i(b)
	if err != nil {
		slog.Error("unable to compare", "err", err)
		return 0
	}

	return a_i - b_i
}

func crit_to_i(crit string) (int, error) {
	spl := strings.Split(crit, ".")
	if len(spl) < 3 {
		return 0, fmt.Errorf("invalid criteria name: %s", crit)
	}
	return strconv.Atoi(spl[1])
}
