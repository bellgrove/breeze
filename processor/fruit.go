package processor

import (
	"encoding/json"
	"log/slog"
	"math"
	"time"
)

type Fruit struct {
	SizerTime        time.Time
	CarrierId        string
	SchemaVer        int
	Status           string
	Lane             int
	Frame            int
	Rod              int
	Cup              int
	CupWeight        float32
	CupWAI           float32
	Side             string
	BatchId          int
	BatchName        string
	BatchGuid        string
	VarietyName      string
	VarietyGuid      string
	ProductName      string
	ProductGuid      string
	ProductPack      string
	OutletName       string
	OutletId         int
	OutletTotalled   bool
	SizeName         string
	SizeId           int
	GradeName        string
	GradeId          int
	IsSampled        bool
	Area             float32
	Weight           float32
	CartonEquivalent float32
	Density          float32
	LeftOffset       float32
	MajorDim         float32
	MinorDim         float32
	Volume           float32
	VisionGrade      string
	VisionValue      float32

	RotationTotal      float32
	RotationProcessed  float32
	SubgradeIndex      int
	SkinImages         int
	StemDetectionError int

	CenterOffsets   []byte
	ClassifiedBlob  []byte
	Colour          []byte
	ColourBlob      []byte
	Diameters       []byte
	Features        []byte
	Function        []byte
	Timer           []byte
	MCenterOffsets  map[string]any
	MClassifiedBlob map[string]any
	MColour         map[string]any
	MColourBlob     map[string]any
	MDiameters      map[string]any
	MFeatures       map[string]any
	MFunction       map[string]any
	MTimer          map[string]any

	ProcessingTime time.Duration
	PrimaryDefect  string
	OtherDefects   []string
	PrimaryReason  string
	OtherReason    []string
}

func Columns() []string {
	return []string{
		"time",
		"carrier_id",
		"schema_ver",
		"status",
		"lane",
		"frame",
		"rod",
		"cup",
		"cup_weight",
		"cup_wai",
		"side",
		"batch_id",
		"batch_name",
		"batch_guid",
		"variety_name",
		"variety_guid",
		"product_name",
		"product_guid",
		"product_pack",
		"outlet_name",
		"outlet_id",
		"outlet_totalled",
		"size_name",
		"size_id",
		"grade_name",
		"grade_id",
		"is_sampled",
		"area",
		"weight",
		"carton_equivalent",
		"density",
		"left_offset",
		"major_dim",
		"minor_dim",
		"volume",
		"vision_grade",
		"vision_value",
		"rotation_total",
		"rotation_processed",
		"subgrade_index",
		"skin_images",
		"stem_detection_error",
		"center_offsets",
		"classified_blob",
		"colour",
		"colour_blob",
		"diameters",
		"features",
		"function",
		"timer",
		"processing_time",
		"primary_defect",
		"other_defects",
		"primary_reason",
		"other_reasons",
	}
}

func (a *Fruit) AsRow() ([]any, error) {
	return []any{
		a.SizerTime,
		a.CarrierId,
		a.SchemaVer,
		a.Status,
		a.Lane,
		a.Frame,
		a.Rod,
		a.Cup,
		a.CupWeight,
		a.CupWAI,
		a.Side,
		a.BatchId,
		a.BatchName,
		a.BatchGuid,
		a.VarietyName,
		a.VarietyGuid,
		a.ProductName,
		a.ProductGuid,
		a.ProductPack,
		a.OutletName,
		a.OutletId,
		a.OutletTotalled,
		a.SizeName,
		a.SizeId,
		a.GradeName,
		a.GradeId,
		a.IsSampled,
		a.Area,
		a.Weight,
		a.CartonEquivalent,
		a.Density,
		a.LeftOffset,
		a.MajorDim,
		a.MinorDim,
		a.Volume,
		a.VisionGrade,
		a.VisionValue,
		a.RotationTotal,
		a.RotationProcessed,
		a.SubgradeIndex,
		a.SkinImages,
		a.StemDetectionError,
		a.CenterOffsets,
		a.ClassifiedBlob,
		a.Colour,
		a.ColourBlob,
		a.Diameters,
		a.Features,
		a.Function,
		a.Timer,
		a.ProcessingTime,
		a.PrimaryDefect,
		a.OtherDefects,
		a.PrimaryReason,
		a.OtherReason,
	}, nil
}

func (a *Fruit) UnmarshalJSON(b []byte) error {
	var breeze struct {
		CarrierId string
		Sizer     struct {
			Schema struct {
				Version int
			}
			Status  string
			Lane    int
			Frame   int
			Carrier struct {
				Rod      int
				Cup      int
				Weight_g float32
				Wai_g    float32
				Side     string
			}
			Batch struct {
				Id   int
				Guid string
				Name string
			}
			Variety struct {
				Id   string
				Name string
			}
			Product struct {
				Id       string
				Name     string
				Packname string
			}
			Outlet struct {
				Id          int
				Name        string
				Is_totalled bool
			}
			Size struct {
				Id   int
				Name string
			}
			Grade struct {
				Id   int
				Name string
			}
			Is_sampled        bool
			Weight_g          float32
			Carton_equivalent float32
			Density_kg_m3     float32
			Left_offset_g     float32
			Major_mm          float32
			Minor_mm          float32
			Vision_grade      string
			Timestamp         time.Time
		}
		Spectrim struct {
			Total_rotation_degrees     float32
			Processed_rotation_degrees float32
			Subgrade_index_1based      int
			Skin_images_count          int
			Stem_detection_error       int
			Surface_area_mm2           float32
			Volume_ml                  float32
			Vision_grade_value         float32

			Center_offsets  map[string]any
			Classified_blob map[string]any
			Color           map[string]any
			Color_blob      map[string]any
			Diameters       map[string]any
			Features        map[string]any
			Function        map[string]any
			Timer           map[string]any
		}
	}

	// var f any
	json.Unmarshal(b, &breeze)
	a.CarrierId = breeze.CarrierId
	a.SchemaVer = breeze.Sizer.Schema.Version
	a.Status = breeze.Sizer.Status
	a.Lane = breeze.Sizer.Lane
	a.Frame = breeze.Sizer.Frame
	a.Rod = breeze.Sizer.Carrier.Rod
	a.Cup = breeze.Sizer.Carrier.Cup
	a.CupWeight = breeze.Sizer.Carrier.Weight_g
	a.CupWAI = breeze.Sizer.Carrier.Wai_g
	a.Side = breeze.Sizer.Carrier.Side
	a.BatchId = breeze.Sizer.Batch.Id
	a.BatchName = breeze.Sizer.Batch.Name
	a.BatchGuid = breeze.Sizer.Batch.Guid
	a.VarietyName = breeze.Sizer.Variety.Name
	a.VarietyGuid = breeze.Sizer.Variety.Id
	a.ProductName = breeze.Sizer.Product.Name
	a.ProductGuid = breeze.Sizer.Product.Id
	a.ProductPack = breeze.Sizer.Product.Packname
	a.OutletName = breeze.Sizer.Outlet.Name
	a.OutletId = breeze.Sizer.Outlet.Id
	a.OutletTotalled = breeze.Sizer.Outlet.Is_totalled
	a.SizeName = breeze.Sizer.Size.Name
	a.SizeId = breeze.Sizer.Size.Id
	a.GradeName = breeze.Sizer.Grade.Name
	a.GradeId = breeze.Sizer.Grade.Id
	a.IsSampled = breeze.Sizer.Is_sampled
	a.Area = breeze.Spectrim.Surface_area_mm2
	a.Weight = breeze.Sizer.Weight_g
	a.CartonEquivalent = breeze.Sizer.Carton_equivalent
	a.Density = breeze.Sizer.Density_kg_m3
	a.LeftOffset = breeze.Sizer.Left_offset_g
	a.MajorDim = breeze.Sizer.Major_mm
	a.MinorDim = breeze.Sizer.Minor_mm
	a.Volume = breeze.Spectrim.Volume_ml
	a.VisionGrade = breeze.Sizer.Vision_grade
	a.VisionValue = breeze.Spectrim.Vision_grade_value
	a.SizerTime = breeze.Sizer.Timestamp

	a.RotationTotal = breeze.Spectrim.Total_rotation_degrees
	a.RotationProcessed = breeze.Spectrim.Processed_rotation_degrees
	a.SubgradeIndex = breeze.Spectrim.Subgrade_index_1based
	a.SkinImages = breeze.Spectrim.Skin_images_count
	a.StemDetectionError = breeze.Spectrim.Stem_detection_error

	if val, err := json.Marshal(breeze.Spectrim.Center_offsets); err == nil {
		a.CenterOffsets = val
		a.MCenterOffsets = breeze.Spectrim.Center_offsets
	}
	if val, err := json.Marshal(breeze.Spectrim.Classified_blob); err == nil {
		a.ClassifiedBlob = val
		a.MClassifiedBlob = breeze.Spectrim.Classified_blob
	}
	if val, err := json.Marshal(breeze.Spectrim.Color); err == nil {
		a.Colour = val
		a.MColour = breeze.Spectrim.Color
	}
	if val, err := json.Marshal(breeze.Spectrim.Color_blob); err == nil {
		a.ColourBlob = val
		a.MColourBlob = breeze.Spectrim.Color_blob
	}
	if val, err := json.Marshal(breeze.Spectrim.Diameters); err == nil {
		a.Diameters = val
		a.MDiameters = breeze.Spectrim.Diameters
	}
	if val, err := json.Marshal(breeze.Spectrim.Features); err == nil {
		a.Features = val
		a.MFeatures = breeze.Spectrim.Features
	}
	if val, err := json.Marshal(breeze.Spectrim.Function); err == nil {
		a.Function = val
		a.MFunction = breeze.Spectrim.Function
	}
	if val, err := json.Marshal(breeze.Spectrim.Timer); err == nil {
		a.Timer = val
		a.MTimer = breeze.Spectrim.Timer
	}

	if breeze.Spectrim.Timer["Total time for fruit processing"] != nil {
		a.ProcessingTime = time.Duration(int(math.Round(1000000 * breeze.Spectrim.Timer["Total time for fruit processing"].(float64))))
	} else {
		slog.Warn("Invalid fruit processing time", "raw", b)
	}
	// a.OtherDefects = []string{"ABCD"}
	return nil
}
