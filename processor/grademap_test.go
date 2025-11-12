package processor

import (
	"encoding/json"
	"log/slog"
	"testing"
	"time"
)

func TestGrademap_Grade(t *testing.T) {
	type args struct {
		fr_json Fruit
		gm_json string
	}
	tests := []struct {
		name string
		g    Grademap
		args args
		exp  string
	}{
		// {"Pink lady C2", Grademap{}, args{c2_fruit_json, gsm_gm_json}},
		{"Avo class 1", Grademap{}, args{c1_avo, avo2_gm_json}, "SC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.g.Update([]byte(tt.args.gm_json))
			if err != nil {
				t.Errorf("FromJson() error = %v", err)
				return
			}
			// v, _ := json.Marshal(tt.args.fr_json)
			// slog.Info("f", "f", v)
			f := tt.args.fr_json
			json.Unmarshal(f.ClassifiedBlob, &f.MClassifiedBlob)
			json.Unmarshal(f.Function, &f.MFunction)
			json.Unmarshal(f.ColourBlob, &f.MColourBlob)
			json.Unmarshal(f.Colour, &f.MColour)
			// json.Unmarshal([]byte(tt.args.fr_json), &f)
			f.VisionGrade = "B"
			tt.g.Grade(&f)
			slog.Info("Grading result", "PrimaryDefect", f.PrimaryDefect, "OtherDefects", f.OtherDefects, "PrimaryReason", f.PrimaryReason, "OtherReasons", f.OtherReason)

			if f.PrimaryDefect != tt.exp {
				t.Errorf("Grade() = %#v, want %#v", f.PrimaryDefect, tt.exp)
				return
			}
		})
	}
}

func TestGrademap_Update(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		g    Grademap
		args args
	}{
		{"GrannySmith DPC 2024", Grademap{}, args{[]byte(gsm_gm_json)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.g.Update(tt.args.b)
			if err != nil {
				t.Errorf("FromJson() error = %v", err)
				return
			}
			if tt.name != tt.g.Name {
				t.Errorf("Update() = %#v, want %#v", tt.g.Name, tt.name)
				return
			}
		})
	}
}

var c1_time, _ = time.Parse(time.RFC3339, "2025-05-20T05:52:43.615Z")
var c1_avo_str = "{\"SizerTime\":\"2025-05-20T05:52:43.615Z\",\"CarrierId\":\"0301030018B3DA03\",\"SchemaVer\":12,\"Status\":\"delivered\",\"Lane\":3,\"Frame\":1,\"Rod\":181,\"Cup\":685,\"CupWeight\":162.8,\"CupWAI\":0.18,\"Side\":\"down\",\"BatchId\":6846,\"BatchName\":\"1080222\",\"BatchGuid\":\"fa09fb59-117e-4e70-bd76-b316008b4426\",\"VarietyName\":\"Avocado\",\"VarietyGuid\":\"23fbe05f-79c7-4ee7-818d-55dba28da3b0\",\"ProductName\":\"OB BIN\",\"ProductGuid\":\"ac4948a6-d6f9-49dc-85f8-2afbda94ba0d\",\"ProductPack\":\"Bin\",\"OutletName\":\"BF2\",\"OutletId\":20,\"OutletTotalled\":true,\"SizeName\":\"30\",\"SizeId\":5,\"GradeName\":\"Class 1\",\"GradeId\":3,\"IsSampled\":false,\"Area\":14422.28,\"Weight\":165.2,\"CartonEquivalent\":0.0333333,\"Density\":1003,\"LeftOffset\":-75.2,\"MajorDim\":78.5,\"MinorDim\":80.2,\"Volume\":229.03058,\"VisionGrade\":\"C\",\"VisionValue\":66.17358,\"RotationTotal\":-436.04715,\"RotationProcessed\":-360,\"SubgradeIndex\":8,\"SkinImages\":33,\"StemDetectionError\":0,\"CenterOffsets\":\"eyJGcnVpdCBDZW50ZXIgT2Zmc2V0IFggQXZnIChtbSkiOiA0LjI2NzQzODg4ODU0OTgwNSwgIkZydWl0IENlbnRlciBPZmZzZXQgWCBNYXggKG1tKSI6IDE3LjY1NTMxMzQ5MTgyMTI5LCAiRnJ1aXQgQ2VudGVyIE9mZnNldCBYIE1pbiAobW0pIjogLTkuNDYyNTY5MjM2NzU1MzcxLCAiRnJ1aXQgQ2VudGVyIE9mZnNldCBZIEF2ZyAobW0pIjogMy42NDE3ODI1MjIyMDE1MzgsICJGcnVpdCBDZW50ZXIgT2Zmc2V0IFkgTWF4IChtbSkiOiA0Ljk1MTkzOTEwNTk4NzU0OSwgIkZydWl0IENlbnRlciBPZmZzZXQgWSBNaW4gKG1tKSI6IDIuNDMwODQ4ODM2ODk4ODAzN30gfCB7IjEuIFNjYXIiOiA3LjA2MDQxMTkzMDA4NDIyODUsICIzLiBOZXR0aW5nIjogMTI0LjczMjI4NDU0NTg5ODQ0LCAiMS4gU2NhciBCaWciOiA0LCAiMS4gSW5zZWN0IERhbWFnZSI6IDEsICIxLiBTY2FyIDEwMCBQeCA+IjogMTAxOC4yNzIzOTk5MDIzNDM4LCAiMS4gVG90YWwgQmxlbWlzaCI6IDEyNjMuOTUyMjcwNTA3ODEyNSwgIjEuIFNjYXIgU21hbGwgdmVpbiI6IDcsICIxLiBTY2FyIHNpemUgY291bnQiOiAxfQ==\",\"ClassifiedBlob\":\"eyIxLiBTY2FyIjogNy4wNjA0MTE5MzAwODQyMjg1LCAiMy4gTmV0dGluZyI6IDEyNC43MzIyODQ1NDU4OTg0NCwgIjEuIFNjYXIgQmlnIjogNCwgIjEuIEluc2VjdCBEYW1hZ2UiOiAxLCAiMS4gU2NhciAxMDAgUHggPiI6IDEwMTguMjcyMzk5OTAyMzQzOCwgIjEuIFRvdGFsIEJsZW1pc2giOiAxMjYzLjk1MjI3MDUwNzgxMjUsICIxLiBTY2FyIFNtYWxsIHZlaW4iOiA3LCAiMS4gU2NhciBzaXplIGNvdW50IjogMX0=\",\"Colour\":\"eyJCbGFjayI6IDAsICJCcm93biI6IDAuMzA1Mzc5MzYwOTE0MjMwMzUsICJHcmVlbiI6IDgwLjkxODgyMzI0MjE4NzUsICJTdGlja2VyIjogMCwgIkNvbnZleW9yIjogMCwgIkRhcmsgR3JlZW4iOiAxMS4zMjkyMzg4OTE2MDE1NjIsICJMaWdodCBHcmVlbiI6IDQuNTYzOTExNDM3OTg4MjgxLCAiU3VuYnVybiBSZWQiOiAwLCAiRGVmZWN0IENvbG9yIDgiOiA3Ni44NzM4MzI3MDI2MzY3MiwgIlN1bmJ1cm4gWWVsbG93IjogMS45NTc1NjAwNjI0MDg0NDczLCAiMS4gQ29sb3IgRGVmZWN0IjogOTA1LjkyOTc0ODUzNTE1NjIsICJEZWZlY3QgQ29sb3IgMTIiOiAzODMuNDA0MzI3MzkyNTc4MSwgIlB1cnBsZSBSaXBlbmluZyI6IDAuOTE3MjU2NzEyOTEzNTEzMn0=\",\"ColourBlob\":\"eyJUb3RhbCBMYXJnZSBCbG9iIE51bSI6IDQsICJUb3RhbCBTbWFsbCBCbG9iIE51bSI6IDMsICJUb3RhbCBNZWRpdW0gQmxvYiBOdW0iOiA5LCAiTGFyZ2UgMS4gQ29sb3IgRGVmZWN0IEJsb2IgTnVtIjogNCwgIkxhcmdlIERlZmVjdCBDb2xvciAxMiBCbG9iIE51bSI6IDEsICJNZWRpdW0gRGVmZWN0IENvbG9yIDggQmxvYiBOdW0iOiAxLCAiU21hbGwgMS4gQ29sb3IgRGVmZWN0IEJsb2IgTnVtIjogMywgIlNtYWxsIERlZmVjdCBDb2xvciAxMiBCbG9iIE51bSI6IDEsICJNZWRpdW0gMS4gQ29sb3IgRGVmZWN0IEJsb2IgTnVtIjogOSwgIk1lZGl1bSBEZWZlY3QgQ29sb3IgMTIgQmxvYiBOdW0iOiA1fQ==\",\"Diameters\":\"eyJSb3VuZG5lc3MgQXZnIjogNzguNzgxOTkwMDUxMjY5NTMsICJSb3VuZG5lc3MgTWF4IjogODEuMTUyODU0OTE5NDMzNiwgIlJvdW5kbmVzcyBNaW4iOiA3Ni43NzMxMDk0MzYwMzUxNiwgIlN0ZW0gQXJlYSBBdmciOiAwLCAiU3RlbSBBcmVhIE1heCI6IDAsICJTdGVtIEFyZWEgTWluIjogMCwgIlNxdWFyZW5lc3MgQXZnIjogNzcuODc0NzcxMTE4MTY0MDYsICJTcXVhcmVuZXNzIE1heCI6IDgyLjUxMTc3OTc4NTE1NjI1LCAiU3F1YXJlbmVzcyBNaW4iOiA3NC4xMjk3OTg4ODkxNjAxNiwgIk1ham9yIERpYW1ldGVyIChtbSkiOiA4Mi4wOTQ2NTc4OTc5NDkyMiwgIk1pbm9yIERpYW1ldGVyIChtbSkiOiA2NS42Mjk4Mjk0MDY3MzgyOCwgIlN0ZW0gTGVuZ3RoIEF2ZyAobW0pIjogMi4xMjgyNDM5MjMxODcyNTYsICJTdGVtIExlbmd0aCBNYXggKG1tKSI6IDIuMjgzNDczOTY4NTA1ODU5NCwgIlN0ZW0gTGVuZ3RoIE1pbiAobW0pIjogMi4wMjM2NTg3NTI0NDE0MDYyLCAiVHJhcGV6b2lkIEhlaWdodCBBdmciOiA4Mi4wOTQ2NTc4OTc5NDkyMiwgIlRyYXBlem9pZCBIZWlnaHQgTWF4IjogODIuOTc0ODk5MjkxOTkyMTksICJUcmFwZXpvaWQgSGVpZ2h0IE1pbiI6IDgxLjQ5NTQ5ODY1NzIyNjU2LCAiTWF4IERpYW1ldGVyIEF2ZyAobW0pIjogODIuMDk0NjU3ODk3OTQ5MjIsICJNYXggRGlhbWV0ZXIgTWF4IChtbSkiOiA4Mi45NzQ4OTkyOTE5OTIxOSwgIk1heCBEaWFtZXRlciBNaW4gKG1tKSI6IDgxLjQ5NTQ5ODY1NzIyNjU2LCAiTWluIERpYW1ldGVyIEF2ZyAobW0pIjogNjQuMjk4MzQ3NDczMTQ0NTMsICJNaW4gRGlhbWV0ZXIgTWF4IChtbSkiOiA2NS42MzkyMjExOTE0MDYyNSwgIk1pbiBEaWFtZXRlciBNaW4gKG1tKSI6IDYyLjIzOTUzMjQ3MDcwMzEyNSwgIlN0ZW0gRGlhbWV0ZXIgQXZnIChtbSkiOiA4MS4wNTE3MTIwMzYxMzI4MSwgIlN0ZW0gRGlhbWV0ZXIgTWF4IChtbSkiOiA4MS4yMTUyNzA5OTYwOTM3NSwgIlN0ZW0gRGlhbWV0ZXIgTWluIChtbSkiOiA4MC44ODY1OTY2Nzk2ODc1LCAiVHJhcGV6b2lkIFNob3VsZGVyIEF2ZyI6IDY0LjI5NjY4NDI2NTEzNjcyLCAiVHJhcGV6b2lkIFNob3VsZGVyIE1heCI6IDY1LjYyOTgyOTQwNjczODI4LCAiVHJhcGV6b2lkIFNob3VsZGVyIE1pbiI6IDYyLjIzNzg2OTI2MjY5NTMxLCAiSW1hZ2VzIFVzZWQgKERpYW1ldGVycykiOiAwLCAiU3F1YXJlbmVzcyAoS2l3aWZydWl0KSBBdmciOiAyNy44NzQ3Nzg3NDc1NTg1OTQsICJTcXVhcmVuZXNzIChLaXdpZnJ1aXQpIE1heCI6IDMyLjUxMTc3NTk3MDQ1ODk4NCwgIlNxdWFyZW5lc3MgKEtpd2lmcnVpdCkgTWluIjogMjQuMTI5NzkzMTY3MTE0MjU4LCAiU3RlbSBEaXJlY3Rpb24gKDAtMTgwKSBBdmciOiA4OS44NDg0ODc4NTQwMDM5LCAiU3RlbSBEaXJlY3Rpb24gKDAtMTgwKSBNYXgiOiAxMTMsICJTdGVtIERpcmVjdGlvbiAoMC0xODApIE1pbiI6IDcyLCAiVmVydGljYWwgRGlhbWV0ZXIgQXZnIChtbSkiOiA4MS40MzI5NjA1MTAyNTM5LCAiVmVydGljYWwgRGlhbWV0ZXIgTWF4IChtbSkiOiA4Mi42OTQ3NjMxODM1OTM3NSwgIlZlcnRpY2FsIERpYW1ldGVyIE1pbiAobW0pIjogNzkuNzg5NTk2NTU3NjE3MTksICJFeHBlcmltZW50YWwgRGlhbWV0ZXIgMSBBdmciOiA4Mi4wOTQ2NTc4OTc5NDkyMiwgIkV4cGVyaW1lbnRhbCBEaWFtZXRlciAxIE1heCI6IDgyLjA5NDY1Nzg5Nzk0OTIyLCAiRXhwZXJpbWVudGFsIERpYW1ldGVyIDEgTWluIjogODIuMDk0NjU3ODk3OTQ5MjIsICJFeHBlcmltZW50YWwgRGlhbWV0ZXIgMiBBdmciOiA4Mi4wOTQ2NTc4OTc5NDkyMiwgIkV4cGVyaW1lbnRhbCBEaWFtZXRlciAyIE1heCI6IDgyLjA5NDY1Nzg5Nzk0OTIyLCAiRXhwZXJpbWVudGFsIERpYW1ldGVyIDIgTWluIjogODIuMDk0NjU3ODk3OTQ5MjIsICJMb25nZXN0IEhvcml6b250YWwgQXZnIChtbSkiOiA2NC4zOTU2ODMyODg1NzQyMiwgIkxvbmdlc3QgSG9yaXpvbnRhbCBNYXggKG1tKSI6IDY1Ljc5Mjk5MTYzODE4MzYsICJMb25nZXN0IEhvcml6b250YWwgTWluIChtbSkiOiA2Mi4yNTEzNjU2NjE2MjEwOTQsICJQZXJpbWV0ZXIgRGlhbWV0ZXIgQXZnIChtbSkiOiA3Mi42MTgxNDg4MDM3MTA5NCwgIlBlcmltZXRlciBEaWFtZXRlciBNYXggKG1tKSI6IDc0LjA5MTkzNDIwNDEwMTU2LCAiUGVyaW1ldGVyIERpYW1ldGVyIE1pbiAobW0pIjogNzAuODIyNDk0NTA2ODM1OTQsICIyRCBPcmllbnRhdGlvbiBCb3ggQW5nbGUgQXZnIjogODMuNjcwMjM0NjgwMTc1NzgsICIyRCBPcmllbnRhdGlvbiBCb3ggQW5nbGUgTWF4IjogODkuNjgyMjQzMzQ3MTY3OTcsICIyRCBPcmllbnRhdGlvbiBCb3ggQW5nbGUgTWluIjogNzMuNzc3ODAxNTEzNjcxODgsICIyRCBPcmllbnRhdGlvbiBCb3ggV2lkdGggQXZnIjogNjQuNzM2MDMwNTc4NjEzMjgsICIyRCBPcmllbnRhdGlvbiBCb3ggV2lkdGggTWF4IjogNjYuNzEzNTQ2NzUyOTI5NjksICIyRCBPcmllbnRhdGlvbiBCb3ggV2lkdGggTWluIjogNjIuOTgyNDY3NjUxMzY3MTksICJFcXVhdG9yaWFsIERpYW1ldGVyIEF2ZyAobW0pIjogNzMuMTk2NTAyNjg1NTQ2ODgsICJFcXVhdG9yaWFsIERpYW1ldGVyIE1heCAobW0pIjogODIuMDk0NjU3ODk3OTQ5MjIsICJFcXVhdG9yaWFsIERpYW1ldGVyIE1pbiAobW0pIjogNjQuMjk4MzQ3NDczMTQ0NTMsICJIb3Jpem9udGFsIERpYW1ldGVyIEF2ZyAobW0pIjogNjQuMzM3NDMyODYxMzI4MTIsICJIb3Jpem9udGFsIERpYW1ldGVyIE1heCAobW0pIjogNjUuNzAyNDc2NTAxNDY0ODQsICJIb3Jpem9udGFsIERpYW1ldGVyIE1pbiAobW0pIjogNjIuMjQwNTAxNDAzODA4NTk0LCAiMkQgT3JpZW50YXRpb24gQm94IEhlaWdodCBBdmciOiA4MC44NDcxMjk4MjE3NzczNCwgIjJEIE9yaWVudGF0aW9uIEJveCBIZWlnaHQgTWF4IjogODIuODc3NzkyMzU4Mzk4NDQsICIyRCBPcmllbnRhdGlvbiBCb3ggSGVpZ2h0IE1pbiI6IDc4LjM5NjE5NDQ1ODAwNzgxLCAiU3RlbSBEaXJlY3Rpb24gKENoZXJyaWVzKSBBdmciOiA4OCwgIlN0ZW0gRGlyZWN0aW9uIChDaGVycmllcykgTWF4IjogMTA4LCAiU3RlbSBEaXJlY3Rpb24gKENoZXJyaWVzKSBNaW4iOiA2NywgIlNxdWFyZW5lc3MgKEtpd2lmcnVpdCBFT0UpIEF2ZyI6IDI3Ljg3NDc3ODc0NzU1ODU5NCwgIlNxdWFyZW5lc3MgKEtpd2lmcnVpdCBFT0UpIE1heCI6IDMyLjUxMTc3NTk3MDQ1ODk4NCwgIlNxdWFyZW5lc3MgKEtpd2lmcnVpdCBFT0UpIE1pbiI6IDI0LjEyOTc5MzE2NzExNDI1OCwgIlZlcnRpY2FsIERpYW1ldGVyIEF2ZyAocGl4ZWxzKSI6IDE5OS42NDM5ODE5MzM1OTM3NSwgIlZlcnRpY2FsIERpYW1ldGVyIE1heCAocGl4ZWxzKSI6IDIwMi44NTg0NzQ3MzE0NDUzLCAiVmVydGljYWwgRGlhbWV0ZXIgTWluIChwaXhlbHMpIjogMTk1LjIwOTIyODUxNTYyNSwgIlBlcnBlbmRpY3VsYXIgRGlhbWV0ZXIgQXZnIChtbSkiOiA2NC4yOTY2ODQyNjUxMzY3MiwgIlBlcnBlbmRpY3VsYXIgRGlhbWV0ZXIgTWF4IChtbSkiOiA2NS42Mjk4Mjk0MDY3MzgyOCwgIlBlcnBlbmRpY3VsYXIgRGlhbWV0ZXIgTWluIChtbSkiOiA2Mi4yMzc4NjkyNjI2OTUzMSwgIlN0ZW0gQXJlYSAoRWxvbmdhdGVkIFN0ZW1zKSBBdmciOiA2Ny4xNTQ2NDc4MjcxNDg0NCwgIlN0ZW0gQXJlYSAoRWxvbmdhdGVkIFN0ZW1zKSBNYXgiOiA3My4xMTM4NzYzNDI3NzM0NCwgIlN0ZW0gQXJlYSAoRWxvbmdhdGVkIFN0ZW1zKSBNaW4iOiA2Mi41NzUzMDk3NTM0MTc5NywgIlN0ZW0gRGlhbWV0ZXIgKEFwcGxlcykgQXZnIChtbSkiOiA4MC44NDcxMjk4MjE3NzczNCwgIlN0ZW0gRGlhbWV0ZXIgKEFwcGxlcykgTWF4IChtbSkiOiA4Mi44Nzc3OTIzNTgzOTg0NCwgIlN0ZW0gRGlhbWV0ZXIgKEFwcGxlcykgTWluIChtbSkiOiA3OC4zOTYxOTQ0NTgwMDc4MSwgIkhvcml6b250YWwgRGlhbWV0ZXIgQXZnIChwaXhlbHMpIjogMTU5LjEyMDk4NjkzODQ3NjU2LCAiSG9yaXpvbnRhbCBEaWFtZXRlciBNYXggKHBpeGVscykiOiAxNjcuMzI3NzI4MjcxNDg0MzgsICJIb3Jpem9udGFsIERpYW1ldGVyIE1pbiAocGl4ZWxzKSI6IDE1My43NDkwMDgxNzg3MTA5NCwgIkNvbWJpbmVkIEVxdWF0b3JpYWwgRGlhbWV0ZXIgQXZnIChtbSkiOiAwLCAiQ29tYmluZWQgRXF1YXRvcmlhbCBEaWFtZXRlciBNYXggKG1tKSI6IDAsICJDb21iaW5lZCBFcXVhdG9yaWFsIERpYW1ldGVyIE1pbiAobW0pIjogMCwgIkVxdWF0b3JpYWwgRGlhbWV0ZXIgKEFwcGxlcykgQXZnIChtbSkiOiA2NC43MzYwMzA1Nzg2MTMyOCwgIkVxdWF0b3JpYWwgRGlhbWV0ZXIgKEFwcGxlcykgTWF4IChtbSkiOiA2Ni43MTM1NDY3NTI5Mjk2OSwgIkVxdWF0b3JpYWwgRGlhbWV0ZXIgKEFwcGxlcykgTWluIChtbSkiOiA2Mi45ODI0Njc2NTEzNjcxOSwgIlN0ZW0gTGVuZ3RoIChFbG9uZ2F0ZWQgU3RlbXMpIEF2ZyAobW0pIjogMi4xMjgyNDM5MjMxODcyNTYsICJTdGVtIExlbmd0aCAoRWxvbmdhdGVkIFN0ZW1zKSBNYXggKG1tKSI6IDIuMjgzNDczOTY4NTA1ODU5NCwgIlN0ZW0gTGVuZ3RoIChFbG9uZ2F0ZWQgU3RlbXMpIE1pbiAobW0pIjogMi4wMjM2NTg3NTI0NDE0MDYyLCAiU3RlbSBEaWFtZXRlciAoRWxvbmdhdGVkIFN0ZW1zKSBBdmcgKG1tKSI6IDgxLjA1MTcxMjAzNjEzMjgxLCAiU3RlbSBEaWFtZXRlciAoRWxvbmdhdGVkIFN0ZW1zKSBNYXggKG1tKSI6IDgxLjIxNTI3MDk5NjA5Mzc1LCAiU3RlbSBEaWFtZXRlciAoRWxvbmdhdGVkIFN0ZW1zKSBNaW4gKG1tKSI6IDgwLjg4NjU5NjY3OTY4NzUsICIyRCBPcmllbnRhdGlvbiBCb3ggKFN0ZW0gQmFzZWQpIEFuZ2xlIEF2ZyI6IDg5LjkxNDk2Mjc2ODU1NDY5LCAiMkQgT3JpZW50YXRpb24gQm94IChTdGVtIEJhc2VkKSBBbmdsZSBNYXgiOiAxMDYuMjIyMTk4NDg2MzI4MTIsICIyRCBPcmllbnRhdGlvbiBCb3ggKFN0ZW0gQmFzZWQpIEFuZ2xlIE1pbiI6IDc3LjE3MzE0OTEwODg4NjcyLCAiMkQgT3JpZW50YXRpb24gQm94IChTdGVtIEJhc2VkKSBXaWR0aCBBdmciOiA2NC43MzYwMzA1Nzg2MTMyOCwgIjJEIE9yaWVudGF0aW9uIEJveCAoU3RlbSBCYXNlZCkgV2lkdGggTWF4IjogNjYuNzEzNTQ2NzUyOTI5NjksICIyRCBPcmllbnRhdGlvbiBCb3ggKFN0ZW0gQmFzZWQpIFdpZHRoIE1pbiI6IDYyLjk4MjQ2NzY1MTM2NzE5LCAiMkQgT3JpZW50YXRpb24gQm94IChTdGVtIEJhc2VkKSBIZWlnaHQgQXZnIjogODAuODQ3MTI5ODIxNzc3MzQsICIyRCBPcmllbnRhdGlvbiBCb3ggKFN0ZW0gQmFzZWQpIEhlaWdodCBNYXgiOiA4Mi44Nzc3OTIzNTgzOTg0NCwgIjJEIE9yaWVudGF0aW9uIEJveCAoU3RlbSBCYXNlZCkgSGVpZ2h0IE1pbiI6IDc4LjM5NjE5NDQ1ODAwNzgxLCAiRXF1YXRvcmlhbCBEaWFtZXRlciAoRWxvbmdhdGVkIFN0ZW1zKSBBdmcgKG1tKSI6IDY1LjAzODAzMjUzMTczODI4LCAiRXF1YXRvcmlhbCBEaWFtZXRlciAoRWxvbmdhdGVkIFN0ZW1zKSBNYXggKG1tKSI6IDY1LjcwOTY5MzkwODY5MTQsICJFcXVhdG9yaWFsIERpYW1ldGVyIChFbG9uZ2F0ZWQgU3RlbXMpIE1pbiAobW0pIjogNjQuMDEyNDU4ODAxMjY5NTN9\",\"Features\":\"eyJTcG90cyI6IDAsICJGbGF0bmVzcyI6IDAsICJTeW1tZXRyeSI6IDAsICJCb3ggU2hhcGUiOiA2OC4yOTY4OTAyNTg3ODkwNiwgIkN1cnZhdHVyZSI6IDQuMTExNzA2MjU2ODY2NDU1LCAiTHVtcGluZXNzIjogMTIuODcyMDc3OTQxODk0NTMxLCAiUm91bmRuZXNzIjogNzYuNzczMTA5NDM2MDM1MTYsICJTdGVtIEFyZWEiOiAwLCAiU3RlbSBTaXplIjogMTAxOC4yNzIzOTk5MDIzNDM4LCAiQ2FseXggU2l6ZSI6IDAsICJFbG9uZ2F0aW9uIjogMjQuMzEzMDYwNzYwNDk4MDQ3LCAiU21vb3RobmVzcyI6IDAsICJTcXVhcmVuZXNzIjogMCwgIlN0ZW0gQW5nbGUiOiAxODAsICJGbGF0bmVzcyB2MiI6IDAsICJPcmllbnRhdGlvbiI6IDMuMzM3NTAzNDMzMjI3NTM5LCAiU3RlbSBMZW5ndGgiOiAyLjI4MzQ3Mzk2ODUwNTg1OTQsICJEb3VibGUgRmFjdG9yIjogMCwgIlNxdWFyZW5lc3MgdjIiOiAwLCAiU3RlbSBMZW5ndGggJSI6IDIuODE3MzA1MDg4MDQzMjEzLCAiQ29uY2F2aXR5IENvdW50IjogMCwgIkN1cnZlZCBTeW1tZXRyeSI6IDc2LjI0NjM2ODQwODIwMzEyLCAiR3JhZGUgT3B0aW1pemVyIjogNTAsICJPdmVyc2l6ZSBGYWN0b3IiOiAwLCAiU2xpcHBhZ2UgRmFjdG9yIjogMCwgIlRvdWNoaW5nIEZhY3RvciI6IDAsICJTdGVtIEluIFRvcCBWaWV3IjogMTAwLCAiRmxhdG5lc3MgdjIgKEVPRSkiOiAwLCAiTG9vc2UgU2tpbiAoT25pb24pIjogMCwgIlNxdWF0bmVzcyAoQXBwbGVzKSI6IDAsICJFbG9uZ2F0aW9uIChBcHBsZXMpIjogMCwgIlJvdW5kbmVzcyAoT3ZlcmFsbCkiOiA3OC43ODE5NzQ3OTI0ODA0NywgIlNsaXBwYWdlIChDaGVycmllcykiOiAwLCAiU3F1YXJlbmVzcyB2MiAoRU9FKSI6IDAsICJGbGF0bmVzcyAoS2l3aWZydWl0KSI6IDAsICJTbGlwcGFnZSAoS2l3aWZydWl0KSI6IDYzLjYzNjM2Mzk4MzE1NDMsICJTdGVtIERldGVjdGlvbiBFcnJvciI6IDg3LjUsICJGbGF0bmVzcyAoRGVwcmVjYXRlZCkiOiA1LjQxMTc0Njk3ODc1OTc2NiwgIlNxdWFyZW5lc3MgKEtpd2lmcnVpdCkiOiAwLCAiU3RlbSBEaWFtZXRlciAoQXBwbGVzKSI6IDAsICJXb2JibGluZXNzIChLaXdpZnJ1aXQpIjogMCwgIlNxdWFyZW5lc3MgKERlcHJlY2F0ZWQpIjogNTEuMjc4NTQxNTY0OTQxNDA2LCAiRW5kIE92ZXIgRW5kIChLaXdpZnJ1aXQpIjogMTQuNzU0OTQzODQ3NjU2MjUsICJUcmFwZXpvaWQgQW5nbGUgKERlZ3JlZXMpIjogMTgwLCAiU3RlbSBBcmVhIChFbG9uZ2F0ZWQgU3RlbXMpIjogNzMuMTEzODc2MzQyNzczNDQsICJWZXJ0aWNhbCBEaWFtZXRlciBTdGFiaWxpdHkiOiA0LjgzMDEzOTYzNjk5MzQwOCwgIkVxdWF0b3JpYWwgRGlhbWV0ZXIgKEFwcGxlcykiOiAwLCAiU3RlbSBMZW5ndGggKEVsb25nYXRlZCBTdGVtcykiOiAyLjI4MzQ3Mzk2ODUwNTg1OTQsICJTdGVtIExlbmd0aCAlIChFbG9uZ2F0ZWQgU3RlbXMpIjogMi44MTczMDUwODgwNDMyMTMsICJFeHBlcmltZW50YWwgU2hhcGUgQ2hhcmFjdGVyaXN0aWMgMSI6IDYyLjE1NjUyODQ3MjkwMDM5LCAiRXhwZXJpbWVudGFsIFNoYXBlIENoYXJhY3RlcmlzdGljIDIiOiA2Mi4xNTY1Mjg0NzI5MDAzOX0=\",\"Function\":\"eyJHcmVlbiBBdmciOiAxNTkuNjIzNTUwNDE1MDM5MDYsICJGdW5jdGlvbiBWYWx1ZSI6IDY2LjE3MzU3NjM1NDk4MDQ3LCAiU29mdG5lc3MgMSAodjIpIjogMzkuMjcyMDk0NzI2NTYyNSwgIlNvZnRuZXNzIDIgKHYyKSI6IDI3LjIwNjkwOTE3OTY4NzUsICJJUiBWYXJpYXRpb24gKExhcmdlc3QgQ2hhbm5lbCkiOiA0ODA0Ny40ODgyODEyNX0=\",\"Timer\":\"eyJGaWx0ZXIgZGlhbWV0ZXJzIjogMCwgIkNvcHkgQ2FwdHVyZWQgSW1hZ2UiOiAwLCAiQ2hhaW4gY29kZSB0aW1lIChtcykiOiAxMC45OTAzMDIwODU4NzY0NjUsICJEZXRlY3RpbmcgdGV4dHVyZSAgKG1zKSI6IDU4LjA2NjA5MzQ0NDgyNDIyLCAiU2hhcGUgYW5kIFNpemUgdGltZSAobXMpIjogNDkuOTAyNDkyNTIzMTkzMzYsICJBdmVyYWdpbmcgcmFkaWkgdGltZSAobXMpIjogMCwgIlVuaWZvcm1pdHkgY29ycmVjdGlvbiAobXMpIjogMS45MjQ5OTk1OTQ2ODg0MTU1LCAiQmxvYiBEZXRlY3Rpb24gLSBUb3RhbCAobXMpIjogNDM4Ljg2MzcwODQ5NjA5Mzc1LCAiRG91YmxlcyBkZXRlY3Rpb24gdGltZSAobXMpIjogMCwgIkZpbmRpbmcgRGlhbWV0ZXJzIHRpbWUgKG1zKSI6IDAsICJDaGFpbmNvZGUgdG8gcmFkaWkgdGltZSAobXMpIjogMCwgIlZvbHVtZSBjYWxjdWxhdGlvbiB0aW1lIChtcykiOiAwLCAiQmFuZCBjb2xvciBjb3VudGluZyB0aW1lIChtcykiOiAwLCAiVG90YWwgdGltZSBmb3IgZnJ1aXQgcHJvY2Vzc2luZyI6IDU2OS4xNTQyOTY4NzUsICJCbG9iIERldGVjdGlvbiAtIEZpbmQgaW5pdGlhbCBibG9icyAobXMpIjogNDM4Ljc3ODgzOTExMTMyODF9\",\"MCenterOffsets\":null,\"MClassifiedBlob\":null,\"MColour\":null,\"MColourBlob\":null,\"MDiameters\":null,\"MFeatures\":null,\"MFunction\":null,\"MTimer\":null,\"ProcessingTime\":903936218,\"PrimaryDefect\":\"\",\"OtherDefects\":null,\"PrimaryReason\":\"\",\"OtherReason\":null}"
var c1_avo = Fruit{
	SizerTime:        c1_time,
	CarrierId:        "0301030018B3DA03",
	SchemaVer:        12,
	Status:           "delivered",
	Lane:             3,
	Frame:            1,
	Rod:              181,
	Cup:              685,
	CupWeight:        162.8,
	CupWAI:           0.18,
	Side:             "down",
	BatchId:          6846,
	BatchName:        "1080222",
	BatchGuid:        "fa09fb59-117e-4e70-bd76-b316008b4426",
	VarietyName:      "Avocado",
	VarietyGuid:      "23fbe05f-79c7-4ee7-818d-55dba28da3b0",
	ProductName:      "OB BIN",
	ProductGuid:      "ac4948a6-d6f9-49dc-85f8-2afbda94ba0d",
	ProductPack:      "Bin",
	OutletName:       "BF2",
	OutletId:         20,
	OutletTotalled:   true,
	SizeName:         "30",
	SizeId:           5,
	GradeName:        "Class 1",
	GradeId:          3,
	IsSampled:        false,
	Area:             14422.28,
	Weight:           165.2,
	CartonEquivalent: 0.0333333,
	Density:          1003,
	LeftOffset:       -75.2,
	MajorDim:         78.5,
	MinorDim:         80.2,
	Volume:           229.03057861328125,
	VisionGrade:      "C",
	VisionValue:      66.17358,

	RotationTotal:      -436.04715,
	RotationProcessed:  -360,
	SubgradeIndex:      8,
	SkinImages:         33,
	StemDetectionError: 0,

	CenterOffsets:  []byte(`{"Fruit Center Offset X Avg (mm)": 4.267438888549805, "Fruit Center Offset X Max (mm)": 17.65531349182129, "Fruit Center Offset X Min (mm)": -9.462569236755371, "Fruit Center Offset Y Avg (mm)": 3.641782522201538, "Fruit Center Offset Y Max (mm)": 4.951939105987549, "Fruit Center Offset Y Min (mm)": 2.4308488368988037} | {"1. Scar": 7.0604119300842285, "3. Netting": 124.73228454589844, "1. Scar Big": 4, "1. Insect Damage": 1, "1. Scar 100 Px >": 1018.2723999023438, "1. Total Blemish": 1263.9522705078125, "1. Scar Small vein": 7, "1. Scar size count": 1}`),
	ClassifiedBlob: []byte(`{"1. Scar": 7.0604119300842285, "3. Netting": 124.73228454589844, "1. Scar Big": 4, "1. Insect Damage": 1, "1. Scar 100 Px >": 1018.2723999023438, "1. Total Blemish": 1263.9522705078125, "1. Scar Small vein": 7, "1. Scar size count": 1}`),
	Colour:         []byte(`{"Black": 0, "Brown": 0.30537936091423035, "Green": 80.9188232421875, "Sticker": 0, "Conveyor": 0, "Dark Green": 11.329238891601562, "Light Green": 4.563911437988281, "Sunburn Red": 0, "Defect Color 8": 76.87383270263672, "Sunburn Yellow": 1.9575600624084473, "1. Color Defect": 905.9297485351562, "Defect Color 12": 383.4043273925781, "Purple Ripening": 0.9172567129135132}`),
	ColourBlob:     []byte(`{"Total Large Blob Num": 4, "Total Small Blob Num": 3, "Total Medium Blob Num": 9, "Large 1. Color Defect Blob Num": 4, "Large Defect Color 12 Blob Num": 1, "Medium Defect Color 8 Blob Num": 1, "Small 1. Color Defect Blob Num": 3, "Small Defect Color 12 Blob Num": 1, "Medium 1. Color Defect Blob Num": 9, "Medium Defect Color 12 Blob Num": 5}`),
	Diameters:      []byte(`{"Roundness Avg": 78.78199005126953, "Roundness Max": 81.1528549194336, "Roundness Min": 76.77310943603516, "Stem Area Avg": 0, "Stem Area Max": 0, "Stem Area Min": 0, "Squareness Avg": 77.87477111816406, "Squareness Max": 82.51177978515625, "Squareness Min": 74.12979888916016, "Major Diameter (mm)": 82.09465789794922, "Minor Diameter (mm)": 65.62982940673828, "Stem Length Avg (mm)": 2.128243923187256, "Stem Length Max (mm)": 2.2834739685058594, "Stem Length Min (mm)": 2.0236587524414062, "Trapezoid Height Avg": 82.09465789794922, "Trapezoid Height Max": 82.97489929199219, "Trapezoid Height Min": 81.49549865722656, "Max Diameter Avg (mm)": 82.09465789794922, "Max Diameter Max (mm)": 82.97489929199219, "Max Diameter Min (mm)": 81.49549865722656, "Min Diameter Avg (mm)": 64.29834747314453, "Min Diameter Max (mm)": 65.63922119140625, "Min Diameter Min (mm)": 62.239532470703125, "Stem Diameter Avg (mm)": 81.05171203613281, "Stem Diameter Max (mm)": 81.21527099609375, "Stem Diameter Min (mm)": 80.8865966796875, "Trapezoid Shoulder Avg": 64.29668426513672, "Trapezoid Shoulder Max": 65.62982940673828, "Trapezoid Shoulder Min": 62.23786926269531, "Images Used (Diameters)": 0, "Squareness (Kiwifruit) Avg": 27.874778747558594, "Squareness (Kiwifruit) Max": 32.511775970458984, "Squareness (Kiwifruit) Min": 24.129793167114258, "Stem Direction (0-180) Avg": 89.8484878540039, "Stem Direction (0-180) Max": 113, "Stem Direction (0-180) Min": 72, "Vertical Diameter Avg (mm)": 81.4329605102539, "Vertical Diameter Max (mm)": 82.69476318359375, "Vertical Diameter Min (mm)": 79.78959655761719, "Experimental Diameter 1 Avg": 82.09465789794922, "Experimental Diameter 1 Max": 82.09465789794922, "Experimental Diameter 1 Min": 82.09465789794922, "Experimental Diameter 2 Avg": 82.09465789794922, "Experimental Diameter 2 Max": 82.09465789794922, "Experimental Diameter 2 Min": 82.09465789794922, "Longest Horizontal Avg (mm)": 64.39568328857422, "Longest Horizontal Max (mm)": 65.7929916381836, "Longest Horizontal Min (mm)": 62.251365661621094, "Perimeter Diameter Avg (mm)": 72.61814880371094, "Perimeter Diameter Max (mm)": 74.09193420410156, "Perimeter Diameter Min (mm)": 70.82249450683594, "2D Orientation Box Angle Avg": 83.67023468017578, "2D Orientation Box Angle Max": 89.68224334716797, "2D Orientation Box Angle Min": 73.77780151367188, "2D Orientation Box Width Avg": 64.73603057861328, "2D Orientation Box Width Max": 66.71354675292969, "2D Orientation Box Width Min": 62.98246765136719, "Equatorial Diameter Avg (mm)": 73.19650268554688, "Equatorial Diameter Max (mm)": 82.09465789794922, "Equatorial Diameter Min (mm)": 64.29834747314453, "Horizontal Diameter Avg (mm)": 64.33743286132812, "Horizontal Diameter Max (mm)": 65.70247650146484, "Horizontal Diameter Min (mm)": 62.240501403808594, "2D Orientation Box Height Avg": 80.84712982177734, "2D Orientation Box Height Max": 82.87779235839844, "2D Orientation Box Height Min": 78.39619445800781, "Stem Direction (Cherries) Avg": 88, "Stem Direction (Cherries) Max": 108, "Stem Direction (Cherries) Min": 67, "Squareness (Kiwifruit EOE) Avg": 27.874778747558594, "Squareness (Kiwifruit EOE) Max": 32.511775970458984, "Squareness (Kiwifruit EOE) Min": 24.129793167114258, "Vertical Diameter Avg (pixels)": 199.64398193359375, "Vertical Diameter Max (pixels)": 202.8584747314453, "Vertical Diameter Min (pixels)": 195.209228515625, "Perpendicular Diameter Avg (mm)": 64.29668426513672, "Perpendicular Diameter Max (mm)": 65.62982940673828, "Perpendicular Diameter Min (mm)": 62.23786926269531, "Stem Area (Elongated Stems) Avg": 67.15464782714844, "Stem Area (Elongated Stems) Max": 73.11387634277344, "Stem Area (Elongated Stems) Min": 62.57530975341797, "Stem Diameter (Apples) Avg (mm)": 80.84712982177734, "Stem Diameter (Apples) Max (mm)": 82.87779235839844, "Stem Diameter (Apples) Min (mm)": 78.39619445800781, "Horizontal Diameter Avg (pixels)": 159.12098693847656, "Horizontal Diameter Max (pixels)": 167.32772827148438, "Horizontal Diameter Min (pixels)": 153.74900817871094, "Combined Equatorial Diameter Avg (mm)": 0, "Combined Equatorial Diameter Max (mm)": 0, "Combined Equatorial Diameter Min (mm)": 0, "Equatorial Diameter (Apples) Avg (mm)": 64.73603057861328, "Equatorial Diameter (Apples) Max (mm)": 66.71354675292969, "Equatorial Diameter (Apples) Min (mm)": 62.98246765136719, "Stem Length (Elongated Stems) Avg (mm)": 2.128243923187256, "Stem Length (Elongated Stems) Max (mm)": 2.2834739685058594, "Stem Length (Elongated Stems) Min (mm)": 2.0236587524414062, "Stem Diameter (Elongated Stems) Avg (mm)": 81.05171203613281, "Stem Diameter (Elongated Stems) Max (mm)": 81.21527099609375, "Stem Diameter (Elongated Stems) Min (mm)": 80.8865966796875, "2D Orientation Box (Stem Based) Angle Avg": 89.91496276855469, "2D Orientation Box (Stem Based) Angle Max": 106.22219848632812, "2D Orientation Box (Stem Based) Angle Min": 77.17314910888672, "2D Orientation Box (Stem Based) Width Avg": 64.73603057861328, "2D Orientation Box (Stem Based) Width Max": 66.71354675292969, "2D Orientation Box (Stem Based) Width Min": 62.98246765136719, "2D Orientation Box (Stem Based) Height Avg": 80.84712982177734, "2D Orientation Box (Stem Based) Height Max": 82.87779235839844, "2D Orientation Box (Stem Based) Height Min": 78.39619445800781, "Equatorial Diameter (Elongated Stems) Avg (mm)": 65.03803253173828, "Equatorial Diameter (Elongated Stems) Max (mm)": 65.7096939086914, "Equatorial Diameter (Elongated Stems) Min (mm)": 64.01245880126953}`),
	Features:       []byte(`{"Spots": 0, "Flatness": 0, "Symmetry": 0, "Box Shape": 68.29689025878906, "Curvature": 4.111706256866455, "Lumpiness": 12.872077941894531, "Roundness": 76.77310943603516, "Stem Area": 0, "Stem Size": 1018.2723999023438, "Calyx Size": 0, "Elongation": 24.313060760498047, "Smoothness": 0, "Squareness": 0, "Stem Angle": 180, "Flatness v2": 0, "Orientation": 3.337503433227539, "Stem Length": 2.2834739685058594, "Double Factor": 0, "Squareness v2": 0, "Stem Length %": 2.817305088043213, "Concavity Count": 0, "Curved Symmetry": 76.24636840820312, "Grade Optimizer": 50, "Oversize Factor": 0, "Slippage Factor": 0, "Touching Factor": 0, "Stem In Top View": 100, "Flatness v2 (EOE)": 0, "Loose Skin (Onion)": 0, "Squatness (Apples)": 0, "Elongation (Apples)": 0, "Roundness (Overall)": 78.78197479248047, "Slippage (Cherries)": 0, "Squareness v2 (EOE)": 0, "Flatness (Kiwifruit)": 0, "Slippage (Kiwifruit)": 63.6363639831543, "Stem Detection Error": 87.5, "Flatness (Deprecated)": 5.411746978759766, "Squareness (Kiwifruit)": 0, "Stem Diameter (Apples)": 0, "Wobbliness (Kiwifruit)": 0, "Squareness (Deprecated)": 51.278541564941406, "End Over End (Kiwifruit)": 14.75494384765625, "Trapezoid Angle (Degrees)": 180, "Stem Area (Elongated Stems)": 73.11387634277344, "Vertical Diameter Stability": 4.830139636993408, "Equatorial Diameter (Apples)": 0, "Stem Length (Elongated Stems)": 2.2834739685058594, "Stem Length % (Elongated Stems)": 2.817305088043213, "Experimental Shape Characteristic 1": 62.15652847290039, "Experimental Shape Characteristic 2": 62.15652847290039}`),
	Function:       []byte(`{"Green Avg": 159.62355041503906, "Function Value": 66.17357635498047, "Softness 1 (v2)": 39.2720947265625, "Softness 2 (v2)": 27.2069091796875, "IR Variation (Largest Channel)": 48047.48828125}`),
	Timer:          []byte(`{"Filter diameters": 0, "Copy Captured Image": 0, "Chain code time (ms)": 10.990302085876465, "Detecting texture  (ms)": 58.06609344482422, "Shape and Size time (ms)": 49.90249252319336, "Averaging radii time (ms)": 0, "Uniformity correction (ms)": 1.9249995946884155, "Blob Detection - Total (ms)": 438.86370849609375, "Doubles detection time (ms)": 0, "Finding Diameters time (ms)": 0, "Chaincode to radii time (ms)": 0, "Volume calculation time (ms)": 0, "Band color counting time (ms)": 0, "Total time for fruit processing": 569.154296875, "Blob Detection - Find initial blobs (ms)": 438.7788391113281}`),

	ProcessingTime: time.Duration(903936218),
	PrimaryDefect:  "",
	OtherDefects:   nil,
}

const (
	avo2_gm_json = `{"characteristics":{"1. Color Defect":{"category":"Defect Color","defect_grading_pass_index":0,"display_order":10,"index":11,"mode":"Area (mm²)","power":50,"scale":[0]},"1. Color Defect (count)":{"category":"Defect Color","defect_grading_pass_index":0,"display_order":11,"index":12,"min_value":[0,5,10],"mode":"Area (mm²)","scale":[0,0,0]},"1. Cuts":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":21,"index":19,"mode":"Count","scale":[0]},"1. Ground Rub":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":20,"index":45,"mode":"%","scale":[0]},"1. Insect Damage":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":37,"index":61,"mode":"Count","scale":[0]},"1. Leaf Roller":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":29,"index":24,"mode":"Area (mm²)","scale":[0]},"1. Leaf Roller Count":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":30,"index":10,"mode":"Count","scale":[0]},"1. Machine Damage":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":31,"index":17,"mode":"Area (mm²)","scale":[0]},"1. Mc Damage Count":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":32,"index":47,"mode":"Count","scale":[0]},"1. Nodules":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":35,"index":51,"mode":"Area (mm²)","scale":[0]},"1. Puncture":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":22,"index":20,"mode":"Count","scale":[0]},"1. Rat Chew":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":27,"index":16,"mode":"Area (mm²)","scale":[0]},"1. Rat Chew Count":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":28,"index":23,"mode":"Count","scale":[0]},"1. Ridging":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":36,"index":53,"mode":"Area (mm²)","scale":[0]},"1. Rot":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":26,"index":22,"mode":"Count","scale":[0]},"1. Scar":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":14,"index":15,"mode":"%","scale":[0]},"1. Scar 100 Px >":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":15,"index":30,"mode":"Area (mm²)","scale":[0]},"1. Scar Big":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":19,"index":43,"mode":"Count","scale":[0]},"1. Scar Med":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":18,"index":13,"mode":"Count","scale":[0]},"1. Scar Small vein":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":17,"index":7,"mode":"Count","scale":[0]},"1. Scar size count":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":16,"index":36,"mode":"Count","scale":[0]},"1. Stem Damage":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":33,"index":28,"mode":"Area (mm²)","scale":[0]},"1. Sunburn/Rot %":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":23,"index":21,"mode":"%","scale":[0]},"1. Sunburn/Rot Lrg":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":24,"index":29,"mode":"Count","scale":[0]},"1. Sunburn/Rot Sml":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":25,"index":52,"mode":"Count","scale":[0]},"1. Total Blemish":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":38,"index":62,"mode":"Area (mm²)","scale":[0]},"1. White Scales":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":34,"index":50,"mode":"Count","scale":[0]},"2. Sheep Nose":{"category":"Classified Blobs","defect_grading_pass_index":1,"display_order":43,"index":35,"mode":"Count","scale":[0]},"3":{"category":"Classified Blobs","defect_grading_pass_index":2,"display_order":47,"index":57,"mode":"Area (mm²)","scale":[0]},"3. Ground Rub":{"category":"Classified Blobs","defect_grading_pass_index":2,"display_order":52,"index":60,"mode":"Area (mm²)","scale":[0]},"3. Netting":{"category":"Classified Blobs","defect_grading_pass_index":2,"display_order":51,"index":59,"mode":"Area (mm²)","scale":[0]},"3. Ridging":{"category":"Classified Blobs","defect_grading_pass_index":2,"display_order":49,"index":58,"mode":"Area (mm²)","scale":[0]},"3. Ridging (Good)":{"category":"Classified Blobs","defect_grading_pass_index":2,"display_order":50,"index":25,"mode":"Area (mm²)","scale":[0]},"Black":{"category":"Fruit Color","display_order":4,"index":5,"mode":"%","power":50,"scale":[0]},"Brown":{"category":"Fruit Color","display_order":7,"index":8,"mode":"%","power":50,"scale":[0]},"Calyx Classification (1)":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":13,"index":49,"mode":"Area (mm²)","scale":[0]},"Calyx Classification (2)":{"category":"Classified Blobs","defect_grading_pass_index":1,"display_order":42,"index":33,"mode":"Area (mm²)","scale":[0]},"Calyx Classification (3)":{"category":"Classified Blobs","defect_grading_pass_index":2,"display_order":48,"index":56,"mode":"Area (mm²)","scale":[0]},"Class 36":{"category":"Classified Blobs","defect_grading_pass_index":1,"display_order":44,"index":63,"mode":"Area (mm²)","scale":[0]},"Curved Symmetry":{"category":"Shape","display_order":54,"index":9,"mode":"%","scale":[0]},"Dark Green":{"category":"Fruit Color","display_order":0,"index":2,"mode":"%","power":50,"scale":[0]},"Defect Color 12":{"category":"Defect Color","defect_grading_pass_index":2,"display_order":45,"index":54,"mode":"Area (mm²)","power":50,"scale":[0]},"Defect Color 12 (count)":{"category":"Defect Color","defect_grading_pass_index":2,"display_order":46,"index":55,"min_value":[0,5,10],"mode":"Area (mm²)","scale":[0,0,0]},"Defect Color 8":{"category":"Defect Color","defect_grading_pass_index":1,"display_order":39,"index":31,"mode":"Area (mm²)","power":50,"scale":[0]},"Defect Color 8 (count)":{"category":"Defect Color","defect_grading_pass_index":1,"display_order":40,"index":32,"min_value":[0,5,10],"mode":"Area (mm²)","scale":[0,0,0]},"Double Factor":{"category":"Special","display_order":58,"index":37,"mode":"Normal","scale":[0]},"Elongation":{"category":"Shape","display_order":55,"index":18,"mode":"%","scale":[0]},"Green":{"category":"Fruit Color","display_order":1,"index":1,"mode":"%","power":50,"scale":[0]},"Green Avg (Light Green +)":{"category":"Color Function","display_order":60,"index":42,"min_value":[0],"mode":"Function","scale":[0]},"IR Variation (Largest Channel) (Light Green +)":{"category":"Color Function","display_order":62,"index":27,"min_value":[0],"mode":"Function","scale":[0]},"Light Green":{"category":"Fruit Color","display_order":2,"index":0,"mode":"%","power":50,"scale":[0]},"Lumpiness":{"category":"Shape","display_order":53,"index":26,"mode":"%","scale":[0]},"Missed Grade":{"category":"Special","display_order":56,"index":41,"mode":"Normal","scale":[0]},"Purple Ripening":{"category":"Fruit Color","display_order":3,"index":3,"mode":"%","power":50,"scale":[0]},"Slippage":{"category":"Special","display_order":59,"index":39,"mode":"Normal","scale":[0]},"Softness 1 (v2) (Light Green +)":{"category":"Color Function","display_order":61,"index":14,"min_value":[0],"mode":"Function","scale":[0]},"Softness 2 (v2) (Light Green +)":{"category":"Color Function","display_order":63,"index":38,"min_value":[0],"mode":"Function","scale":[0]},"Stem Classification (1)":{"category":"Classified Blobs","defect_grading_pass_index":0,"display_order":12,"index":48,"mode":"Area (mm²)","scale":[0]},"Stem Classification (2)":{"category":"Classified Blobs","defect_grading_pass_index":1,"display_order":41,"index":34,"mode":"Area (mm²)","scale":[0]},"Sticker":{"category":"Fruit Color","display_order":8,"index":46,"mode":"%","power":50,"scale":[0]},"Sunburn Red":{"category":"Fruit Color","display_order":5,"index":44,"mode":"%","power":50,"scale":[0]},"Sunburn Yellow":{"category":"Fruit Color","display_order":6,"index":6,"mode":"%","power":50,"scale":[0]},"Sunburn Yellow + Brown":{"category":"Color Combination","display_order":9,"index":4,"mode":"%","scale":[0]},"Touching Factor":{"category":"Special","display_order":57,"index":40,"mode":"Normal","scale":[0]}},"computer_name":"SPECTRIMMASTER","defect_grading_passes":{"General Pass":{"index":0},"Nett/Ridge":{"index":2},"Sheeps Nose":{"index":1}},"grades":{"A. Recycle":{"criteria_sets":{"A.0. Missed Grade":{"criteria":[{"characteristic_index":41,"thresholds":[{"max":100.0,"min":2.0}]}],"not_a_fruit":false},"A.1. Touching Factor":{"criteria":[{"characteristic_index":40,"thresholds":[{"max":100.0,"min":50.0}]}],"not_a_fruit":false},"A.2. Double Factor":{"criteria":[{"characteristic_index":37,"thresholds":[{"max":100.0,"min":70.0}]}],"not_a_fruit":false}}},"B. Juice":{"criteria_sets":{"B.0. Tree Colour [TC]":{"criteria":[{"characteristic_index":3,"thresholds":[{"max":100.0,"min":100.0}]}],"not_a_fruit":false},"B.1. Sunburn Yellow + Brown [SUN]":{"criteria":[{"characteristic_index":4,"thresholds":[{"max":100.0,"min":100.0}]}],"not_a_fruit":false},"B.10. 1. Leaf Roller [LR]":{"criteria":[{"characteristic_index":24,"thresholds":[{"max":9999.0,"min":100.0}]}],"not_a_fruit":false},"B.11. 1. Leaf Roller Cnt [LR]":{"criteria":[{"characteristic_index":10,"thresholds":[{"max":9999.0,"min":3.0}]}],"not_a_fruit":false},"B.12. 1. Scar Big [SC]":{"criteria":[{"characteristic_index":43,"thresholds":[{"max":9999.0,"min":25.0}]}],"not_a_fruit":false},"B.13. 1. Mc Damage Count G [MD]":{"criteria":[{"characteristic_index":47,"thresholds":[{"max":9999.0,"min":4.0}]}],"not_a_fruit":false},"B.14. 1.White Scales [WS]":{"criteria":[{"characteristic_index":50,"thresholds":[{"max":9999.0,"min":25.0}]}],"not_a_fruit":false},"B.15. 3. Ground Rub [GR]":{"criteria":[{"characteristic_index":60,"thresholds":[{"max":9999.0,"min":4000.0}]}],"not_a_fruit":false},"B.16. 1. Insect damage [ID]":{"criteria":[{"characteristic_index":61,"thresholds":[{"max":9999.0,"min":12.0}]}],"not_a_fruit":false},"B.2. 1. Scar [SC]":{"criteria":[{"characteristic_index":15,"thresholds":[{"max":100.0,"min":60.0}]}],"not_a_fruit":false},"B.3. 1. DiffuseScar [SC]":{"criteria":[{"characteristic_index":45,"thresholds":[{"max":100.0,"min":50.0}]}],"not_a_fruit":false},"B.4. 1. Cuts [CUT]":{"criteria":[{"characteristic_index":19,"thresholds":[{"max":9999.0,"min":5.0}]}],"not_a_fruit":false},"B.5. 1. Sunburn Rot % [SUN]":{"criteria":[{"characteristic_index":21,"thresholds":[{"max":100.0,"min":14.5}]}],"not_a_fruit":false},"B.6. 1. Puncture [PUN]":{"criteria":[{"characteristic_index":20,"thresholds":[{"max":9999.0,"min":4.0}]}],"not_a_fruit":false},"B.7. 1. Rot [ROT]":{"criteria":[{"characteristic_index":22,"thresholds":[{"max":9999.0,"min":2.0}]}],"not_a_fruit":false},"B.8. 1. Rat Chew [RAT]":{"criteria":[{"characteristic_index":16,"thresholds":[{"max":9999.0,"min":400.0}]}],"not_a_fruit":false},"B.9. 1. Rat Chew Cnt [RAT]":{"criteria":[{"characteristic_index":23,"thresholds":[{"max":9999.0,"min":7.0}]}],"not_a_fruit":false}}},"C. Class 2":{"criteria_sets":{"C.0. Tree Colour [TC]":{"criteria":[{"characteristic_index":3,"thresholds":[{"max":100.0,"min":100.0}]}],"not_a_fruit":false},"C.1. Sunburn Yellow + Brown [SUN]":{"criteria":[{"characteristic_index":4,"thresholds":[{"max":100.0,"min":100.0}]}],"not_a_fruit":false},"C.10. 1. Puncture [PUN]":{"criteria":[{"characteristic_index":20,"thresholds":[{"max":9999.0,"min":3.0}]}],"not_a_fruit":false},"C.11. 1.White Scales [WC]":{"criteria":[{"characteristic_index":50,"thresholds":[{"max":9999.0,"min":11.0}]}],"not_a_fruit":false},"C.12. 3. Ground Rub [GD]":{"criteria":[{"characteristic_index":60,"thresholds":[{"max":9999.0,"min":200.0}]}],"not_a_fruit":false},"C.13. 1. Insect Damage [ID]":{"criteria":[{"characteristic_index":61,"thresholds":[{"max":9999.0,"min":11.0}]}],"not_a_fruit":false},"C.2. 1. Sunburn/Rot Lrg [SUN]":{"criteria":[{"characteristic_index":29,"thresholds":[{"max":9999.0,"min":1.0}]}],"not_a_fruit":false},"C.3. 1. Scar Big [SC]":{"criteria":[{"characteristic_index":43,"thresholds":[{"max":9999.0,"min":9.0}]}],"not_a_fruit":false},"C.4. 1. Scar Med [SC]":{"criteria":[{"characteristic_index":13,"thresholds":[{"max":9999.0,"min":5.0}]}],"not_a_fruit":false},"C.5. 1. Scar Small vein [SC]":{"criteria":[{"characteristic_index":7,"thresholds":[{"max":9999.0,"min":30.0}]}],"not_a_fruit":false},"C.6. 2. Sheep Nose [SN]":{"criteria":[{"characteristic_index":35,"thresholds":[{"max":9999.0,"min":5.0}]}],"not_a_fruit":false},"C.7. 1. Stem Damage [SD]":{"criteria":[{"characteristic_index":28,"thresholds":[{"max":9999.0,"min":182.0}]}],"not_a_fruit":false},"C.8. 1. Scar 100 Px > [SC]":{"criteria":[{"characteristic_index":30,"thresholds":[{"max":9999.0,"min":9000.0}]}],"not_a_fruit":false},"C.9. 1. Rat Chew [RC]":{"criteria":[{"characteristic_index":16,"thresholds":[{"max":9999.0,"min":90.0}]}],"not_a_fruit":false}}},"D. Class 1":{"criteria_sets":{"D.0. Tree colour [TC]":{"criteria":[{"characteristic_index":3,"thresholds":[{"max":100.0,"min":100.0}]}],"not_a_fruit":false},"D.1. 1.Nodules [ND]":{"criteria":[{"characteristic_index":51,"thresholds":[{"max":9999.0,"min":1200.0}]}],"not_a_fruit":false},"D.10. 1. Sunburn [SUN]":{"criteria":[{"characteristic_index":21,"thresholds":[{"max":100.0,"min":1.2000000476837158}]}],"not_a_fruit":false},"D.11. 1. Sunburn/Rot Sml [SUN]":{"criteria":[{"characteristic_index":52,"thresholds":[{"max":9999.0,"min":1.0}]}],"not_a_fruit":false},"D.12. 1. Stem Damage [SD]":{"criteria":[{"characteristic_index":28,"thresholds":[{"max":9999.0,"min":400.0}]}],"not_a_fruit":false},"D.13. 1. Scar size count [SC]":{"criteria":[{"characteristic_index":36,"thresholds":[{"max":9999.0,"min":7.0}]}],"not_a_fruit":false},"D.14. 1. Scar Small vein [SC]":{"criteria":[{"characteristic_index":7,"thresholds":[{"max":9999.0,"min":23.0}]}],"not_a_fruit":false},"D.15. 1.White Scales [WC]":{"criteria":[{"characteristic_index":50,"thresholds":[{"max":9999.0,"min":5.0}]}],"not_a_fruit":false},"D.16. 1.Ridging [RID]":{"criteria":[{"characteristic_index":53,"thresholds":[{"max":9999.0,"min":9999.0}]}],"not_a_fruit":false},"D.17. 3. Ridging Good [RID]":{"criteria":[{"characteristic_index":25,"thresholds":[{"max":9999.0,"min":5000.0}]}],"not_a_fruit":false},"D.18. 1. Insect damage [ID]":{"criteria":[{"characteristic_index":61,"thresholds":[{"max":9999.0,"min":8.0}]}],"not_a_fruit":false},"D.19. 1. Total Blemish [BLE]":{"criteria":[{"characteristic_index":62,"thresholds":[{"max":9999.0,"min":3000.0}]}],"not_a_fruit":false},"D.2. 1. Scar 100px [SC]":{"criteria":[{"characteristic_index":30,"thresholds":[{"max":9999.0,"min":2050.0}]}],"not_a_fruit":false},"D.20. 1. Ground Rub [GD]":{"criteria":[{"characteristic_index":45,"thresholds":[{"max":100.0,"min":2.700000047683716}]}],"not_a_fruit":false},"D.21. 1. Leaf Roller [LR}":{"criteria":[{"characteristic_index":24,"thresholds":[{"max":9999.0,"min":59.0}]}],"not_a_fruit":false},"D.3. 1. Scar [SC}":{"criteria":[{"characteristic_index":15,"thresholds":[{"max":100.0,"min":8.0}]}],"not_a_fruit":false},"D.4. 1. Scar Small [SC]":{"criteria":[{"characteristic_index":7,"thresholds":[{"max":9999.0,"min":24.0}]}],"not_a_fruit":false},"D.5. 1. Scar Med [SC]":{"criteria":[{"characteristic_index":13,"thresholds":[{"max":9999.0,"min":7.0}]}],"not_a_fruit":false},"D.6. 2. Sheep Nose [SN]":{"criteria":[{"characteristic_index":35,"thresholds":[{"max":9999.0,"min":2.0}]}],"not_a_fruit":false},"D.7. 1. Scar Big [SC]":{"criteria":[{"characteristic_index":43,"thresholds":[{"max":9999.0,"min":1.0}]}],"not_a_fruit":false},"D.8. 1. Cuts [CUT]":{"criteria":[{"characteristic_index":19,"thresholds":[{"max":9999.0,"min":1.0}]}],"not_a_fruit":false},"D.9. 1. Puncture [PUN]":{"criteria":[{"characteristic_index":20,"thresholds":[{"max":9999.0,"min":1.0}]}],"not_a_fruit":false}}},"F. Premium":{"criteria_sets":{"F. Premium":{"criteria":[{"characteristic_index":1,"thresholds":[{"max":100.0,"min":1.0}]}],"not_a_fruit":false}}},"G. Premium all":{"criteria_sets":{"G. Premium all":{"not_a_fruit":false}}}},"id":113,"is_primary":true,"name":"New Hass July 23","node_id":4,"schema":{"name":"spectrim_grade_map","version":2},"timestamp":"2025-07-09T22:08:07Z"}`
	pldy_gm_json = `{
  "characteristics": {
    "1. Bad Stem": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 11,
      "index": 38,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Bad Stem JUICE #": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 22,
      "index": 45,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Big Blemish": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 15,
      "index": 67,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Big Mark #": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 23,
      "index": 55,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Bitter Pit": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 27,
      "index": 15,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Blemish": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 16,
      "index": 8,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Blemish count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 17,
      "index": 16,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Bruise": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 12,
      "index": 35,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Bruise Area IR1": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 14,
      "index": 44,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Bruise Count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 13,
      "index": 43,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Hail Area": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 18,
      "index": 64,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. IR Def": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 6,
      "index": 12,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "1. IR Def (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 7,
      "index": 13,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "1. Juice / Rot": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 24,
      "index": 58,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Leaf": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 25,
      "index": 40,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Puncture all": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 19,
      "index": 47,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Puncture big": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 21,
      "index": 46,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Puncture small": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 20,
      "index": 56,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Rot All": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 28,
      "index": 17,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Rot Light": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 29,
      "index": 48,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Split": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 26,
      "index": 9,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Sunburn Rot": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 30,
      "index": 18,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Susp Stem By IR #": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 10,
      "index": 39,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1.Cork Scabbing": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 31,
      "index": 19,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1.Cork-Big": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 32,
      "index": 20,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "2. Brown Rot": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 42,
      "index": 10,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "2. Bruise Count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 38,
      "index": 57,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "2. Bruise Juice Sm": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 39,
      "index": 7,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "2. Bruise Juice big": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 41,
      "index": 33,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "2. Bruise Juice medium": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 40,
      "index": 21,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "2. Bruise SC Area": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 37,
      "index": 2,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "2. IR2 Def": {
      "category": "Defect Color",
      "defect_grading_pass_index": 1,
      "display_order": 33,
      "index": 98,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "2. IR2 Def (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 1,
      "display_order": 34,
      "index": 99,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "2. Large Bitter Pit": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 43,
      "index": 22,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "3. Sun Burn": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 48,
      "index": 34,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "3. Sun Burn Def": {
      "category": "Defect Color",
      "defect_grading_pass_index": 2,
      "display_order": 44,
      "index": 65,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "3. Sun Burn Def (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 2,
      "display_order": 45,
      "index": 79,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "4. Clayx Cracking": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 55,
      "index": 37,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "4. Clayx Cracking count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 56,
      "index": 50,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "4. Split Elongated": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 54,
      "index": 36,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "4. Split round": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 53,
      "index": 27,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "4. Stem Puncture": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 57,
      "index": 49,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "Box Shape": {
      "category": "Shape",
      "display_order": 62,
      "index": 11,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Calyx Classification (1)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 9,
      "index": 70,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Calyx Classification (2)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 36,
      "index": 81,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Calyx Classification (3)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 47,
      "index": 41,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Calyx Classification (4)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 52,
      "index": 25,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Class 45": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 58,
      "index": 51,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Dark Red": {
      "category": "Fruit Color",
      "display_order": 0,
      "index": 0,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Defect Color 7": {
      "category": "Defect Color",
      "defect_grading_pass_index": 3,
      "display_order": 49,
      "index": 23,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Defect Color 7 (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 3,
      "display_order": 50,
      "index": 24,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "Double Factor": {
      "category": "Special",
      "display_order": 66,
      "index": 29,
      "mode": "Normal",
      "scale": [
        0
      ]
    },
    "Elongation": {
      "category": "Shape",
      "display_order": 60,
      "index": 59,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Leaf": {
      "category": "Fruit Color",
      "display_order": 4,
      "index": 5,
      "mode": "per 10k",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Light red": {
      "category": "Fruit Color",
      "display_order": 2,
      "index": 3,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Light red + Yellow": {
      "category": "Color Combination",
      "display_order": 5,
      "index": 6,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Minor Diameter (mm)": {
      "category": "Size Criteria",
      "display_order": 63,
      "index": 68,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Red": {
      "category": "Fruit Color",
      "display_order": 1,
      "index": 1,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Roundness": {
      "category": "Shape",
      "display_order": 61,
      "index": 14,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Slippage": {
      "category": "Special",
      "display_order": 64,
      "index": 31,
      "mode": "Normal",
      "scale": [
        0
      ]
    },
    "Softness 1 (v2) (Dark Red +)": {
      "category": "Color Function",
      "display_order": 68,
      "index": 28,
      "min_value": [
        0
      ],
      "mode": "Function",
      "scale": [
        0
      ]
    },
    "Softness 2 (v2) (Dark Red +)": {
      "category": "Color Function",
      "display_order": 67,
      "index": 32,
      "min_value": [
        0
      ],
      "mode": "Function",
      "scale": [
        0
      ]
    },
    "Stem Angle": {
      "category": "Shape",
      "display_order": 59,
      "index": 53,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Stem Classification (1)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 8,
      "index": 69,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Stem Classification (2)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 35,
      "index": 82,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Stem Classification (3)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 46,
      "index": 42,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Stem Classification (4)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 51,
      "index": 26,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Touching Factor": {
      "category": "Special",
      "display_order": 65,
      "index": 30,
      "mode": "Normal",
      "scale": [
        0
      ]
    },
    "Yellow": {
      "category": "Fruit Color",
      "display_order": 3,
      "index": 4,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    }
  },
  "computer_name": "SPECTRIMMASTER",
  "defect_grading_passes": {
    "4. Splits": {
      "index": 3
    },
    "Master": {
      "index": 0
    },
    "Old Bruise IR2": {
      "index": 1
    },
    "Sun Burn": {
      "index": 2
    }
  },
  "grades": {
    "A. Recycle": {
      "criteria_sets": {
        "A.0. Double Factor": {
          "criteria": [
            {
              "characteristic_index": 29,
              "thresholds": [
                {
                  "max": 100,
                  "min": 98
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "A.1. Touching Factor": {
          "criteria": [
            {
              "characteristic_index": 30,
              "thresholds": [
                {
                  "max": 100,
                  "min": 100
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "A.2. Slippage": {
          "criteria": [
            {
              "characteristic_index": 31,
              "thresholds": [
                {
                  "max": 100,
                  "min": 89
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "B. Juice": {
      "criteria_sets": {
        "B.3. 1. Puncture small [PUN]": {
          "criteria": [
            {
              "characteristic_index": 56,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.4. 1. Puncture all [PUN]": {
          "criteria": [
            {
              "characteristic_index": 47,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.8. 1. Sunburn Rot [SUN]": {
          "criteria": [
            {
              "characteristic_index": 18,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "C. Class 2": {
      "criteria_sets": {
        "C.0. Yellow [COL]": {
          "criteria": [
            {
              "characteristic_index": 4,
              "thresholds": [
                {
                  "max": 100,
                  "min": 89
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.1. 2. Bruise SC Area [BR]": {
          "criteria": [
            {
              "characteristic_index": 2,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 100
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.10. 1. Blemish count [BL]": {
          "criteria": [
            {
              "characteristic_index": 16,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.11. 1.Cork Scabbing [SC]": {
          "criteria": [
            {
              "characteristic_index": 19,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.13. 2. Bruise Juice midium [BR]": {
          "criteria": [
            {
              "characteristic_index": 21,
              "thresholds": [
                {
                  "max": 9997,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.14. 1. Split [SP]": {
          "criteria": [
            {
              "characteristic_index": 9,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 180
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.16. 2. Bruise Juice Sm [BR]": {
          "criteria": [
            {
              "characteristic_index": 7,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 5
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.2. 2. Bruise Count [BR]": {
          "criteria": [
            {
              "characteristic_index": 57,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 3
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.3. 1. Bruise [BR]": {
          "criteria": [
            {
              "characteristic_index": 35,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 110
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.4. 1. Bruise Area IR1 [BR]": {
          "criteria": [
            {
              "characteristic_index": 44,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 200
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.5. 1. Bruise Count [BR]": {
          "criteria": [
            {
              "characteristic_index": 43,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 3
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.6. 1. Hail Area [HA]": {
          "criteria": [
            {
              "characteristic_index": 64,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 35
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.7. 1. Bad Stem [STM]": {
          "criteria": [
            {
              "characteristic_index": 38,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.9. 1. Big Blemish [BL]": {
          "criteria": [
            {
              "characteristic_index": 67,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 4
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "D. Premium": {
      "criteria_sets": {
        "D. Red": {
          "criteria": [
            {
              "characteristic_index": 1,
              "thresholds": [
                {
                  "max": 100,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "E. Grade": {
      "criteria_sets": {
        "E. Grade": {
          "not_a_fruit": false
        }
      }
    }
  },
  "id": 115,
  "is_primary": true,
  "name": "pink lady sort view",
  "node_id": 4,
  "schema": {
    "name": "spectrim_grade_map",
    "version": 2
  },
  "timestamp": "2025-05-21T02:52:20Z"
}`
	gsm_gm_json = `{
  "characteristics": {
    "1. Bad Stem": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 25,
      "index": 27,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Bitter pit": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 40,
      "index": 44,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Blemish #": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 23,
      "index": 21,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Blemish (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 16,
      "index": 36,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "1. Bruise": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 30,
      "index": 31,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Bruise Dark": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 32,
      "index": 80,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Bruise IR2": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 33,
      "index": 37,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Bruise colour all": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 31,
      "index": 68,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Corb Scabing": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 42,
      "index": 48,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Cut": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 38,
      "index": 15,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Defects all": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 24,
      "index": 29,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Hail Large": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 34,
      "index": 25,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Hail Large count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 35,
      "index": 75,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Hail Small": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 36,
      "index": 22,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. IR1 Dark": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 9,
      "index": 6,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "1. IR1 Dark (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 10,
      "index": 7,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "1. IR1 Light": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 7,
      "index": 4,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "1. IR1 Light (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 8,
      "index": 5,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "1. IR2 Dark": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 13,
      "index": 10,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "1. IR2 Dark (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 14,
      "index": 11,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "1. IR2 Light": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 11,
      "index": 8,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "1. IR2 Light (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 12,
      "index": 9,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "1. Juice": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 19,
      "index": 26,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Large Bitter pit": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 41,
      "index": 47,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Leaf": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 37,
      "index": 32,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Punctures": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 26,
      "index": 30,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Punctures Ir2": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 29,
      "index": 60,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Punctures Juice": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 27,
      "index": 83,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Punctures Juice all": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 28,
      "index": 19,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Rot": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 21,
      "index": 23,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Russett": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 39,
      "index": 17,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Small Juice)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 20,
      "index": 33,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "2. Dark Bruise Juice IR2": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 51,
      "index": 65,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "2. Defect Color 12": {
      "category": "Defect Color",
      "defect_grading_pass_index": 1,
      "display_order": 43,
      "index": 40,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "2. Defect Color 12 (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 1,
      "display_order": 44,
      "index": 41,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "2. Defects": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 52,
      "index": 59,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "2. Fine Bruise Light": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 50,
      "index": 74,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "2. Heavy Bruise": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 48,
      "index": 46,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "2. Heavy Bruise all": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 49,
      "index": 86,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "2. Rot": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 47,
      "index": 34,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "2. Watercore": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 53,
      "index": 16,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "24. 1. Blemish": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 22,
      "index": 24,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "3. Bruises": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 58,
      "index": 61,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "3. Defect Color 9": {
      "category": "Defect Color",
      "defect_grading_pass_index": 2,
      "display_order": 54,
      "index": 53,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "3. Defect Color 9 (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 2,
      "display_order": 55,
      "index": 55,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "35. 1. Blemish": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 15,
      "index": 35,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "4. Defect Color 13": {
      "category": "Defect Color",
      "defect_grading_pass_index": 3,
      "display_order": 59,
      "index": 49,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "4. Defect Color 13 (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 3,
      "display_order": 60,
      "index": 50,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "4. Shrivel": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 63,
      "index": 54,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Bleach": {
      "category": "Fruit Color",
      "display_order": 2,
      "index": 1,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Bleach + Blush": {
      "category": "Color Combination",
      "display_order": 5,
      "index": 28,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Blush": {
      "category": "Fruit Color",
      "display_order": 1,
      "index": 2,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Blush + Sunburn": {
      "category": "Color Combination",
      "display_order": 6,
      "index": 3,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Box Shape": {
      "category": "Shape",
      "display_order": 65,
      "index": 18,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Calyx Classification (1)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 18,
      "index": 39,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Calyx Classification (2)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 46,
      "index": 42,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Calyx Classification (3)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 57,
      "index": 56,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Calyx Classification (4)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 62,
      "index": 51,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Double Factor": {
      "category": "Special",
      "display_order": 67,
      "index": 13,
      "mode": "Normal",
      "scale": [
        0
      ]
    },
    "Green": {
      "category": "Fruit Color",
      "display_order": 0,
      "index": 0,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Light Yellow": {
      "category": "Fruit Color",
      "display_order": 4,
      "index": 45,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Roundness": {
      "category": "Shape",
      "display_order": 64,
      "index": 20,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Stem Classification (1)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 17,
      "index": 38,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Stem Classification (2)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 45,
      "index": 43,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Stem Classification (3)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 56,
      "index": 58,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Stem Classification (4)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 3,
      "display_order": 61,
      "index": 52,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Sunburn": {
      "category": "Fruit Color",
      "display_order": 3,
      "index": 12,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Touching Factor": {
      "category": "Special",
      "display_order": 66,
      "index": 14,
      "mode": "Normal",
      "scale": [
        0
      ]
    }
  },
  "computer_name": "SPECTRIMMASTER",
  "defect_grading_passes": {
    "Bruising": {
      "index": 2
    },
    "Fine IR & IR2 Pass": {
      "index": 1
    },
    "General Pass": {
      "index": 0
    },
    "Shrivel": {
      "index": 3
    }
  },
  "grades": {
    "A. Recycle": {
      "criteria_sets": {
        "A.0. Touching Factor": {
          "criteria": [
            {
              "characteristic_index": 14,
              "thresholds": [
                {
                  "max": 20,
                  "min": 15
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "A.1. Double Factor": {
          "criteria": [
            {
              "characteristic_index": 13,
              "thresholds": [
                {
                  "max": 20,
                  "min": 20
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "B. Juice": {
      "criteria_sets": {
        "B.0. Blush + Sunburn": {
          "criteria": [
            {
              "characteristic_index": 3,
              "thresholds": [
                {
                  "max": 100,
                  "min": 22
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.1. 2. Rot [ROT]": {
          "criteria": [
            {
              "characteristic_index": 34,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2000
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.10. 1. Large Bitter pit [BP]": {
          "criteria": [
            {
              "characteristic_index": 47,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.2. 1. Rot [ROT]": {
          "criteria": [
            {
              "characteristic_index": 23,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 135
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.3. 2. Watercore [WAT]": {
          "criteria": [
            {
              "characteristic_index": 16,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 500
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.4. 1. Bad Stem [BS]": {
          "criteria": [
            {
              "characteristic_index": 27,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 9999
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.5. 1. Bruise [BR]": {
          "criteria": [
            {
              "characteristic_index": 31,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 340
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.6. 4. Bruises [BR]": {
          "criteria": [
            {
              "characteristic_index": 61,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 380
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.7. 1. Bitter pit [BP]": {
          "criteria": [
            {
              "characteristic_index": 44,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 6
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.8. 1. Blemish [BLE]": {
          "criteria": [
            {
              "characteristic_index": 24,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1500
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.9. 2. Fine Bruise Light [BR]": {
          "criteria": [
            {
              "characteristic_index": 74,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1000
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "C. Class 2": {
      "criteria_sets": {
        "C.0. Blush + Sunburn": {
          "criteria": [
            {
              "characteristic_index": 3,
              "thresholds": [
                {
                  "max": 100,
                  "min": 11
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.1. 4. Bruises [BR]": {
          "criteria": [
            {
              "characteristic_index": 61,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 120
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.10. 1. Corb Scabing [CS]": {
          "criteria": [
            {
              "characteristic_index": 48,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.11. 2. Heavy Bruise all [BR]": {
          "criteria": [
            {
              "characteristic_index": 86,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 3
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.12. 2. Fine Bruise Light [BR]": {
          "criteria": [
            {
              "characteristic_index": 74,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 53
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.13. 1. Hail Large count [HA]": {
          "criteria": [
            {
              "characteristic_index": 75,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.2. 2. Fine BruiseLight [BR]": {
          "criteria": [
            {
              "characteristic_index": 74,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 80
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.3. 1. Bruise colour all [BR]": {
          "criteria": [
            {
              "characteristic_index": 68,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 50
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.4. 1. Bruise [BR]": {
          "criteria": [
            {
              "characteristic_index": 31,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.5. 1. Blemish [BLE]": {
          "criteria": [
            {
              "characteristic_index": 24,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 340
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.6. 1. Blemish # [BLE]": {
          "criteria": [
            {
              "characteristic_index": 21,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 6
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.7. 1. Bitter pit [BP]": {
          "criteria": [
            {
              "characteristic_index": 44,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 4
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.8. 1. Bad Stem [BS]": {
          "criteria": [
            {
              "characteristic_index": 27,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2300
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.9. 2. Defects [DFT]": {
          "criteria": [
            {
              "characteristic_index": 59,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 150
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "D. Premium": {
      "criteria_sets": {
        "D. Green": {
          "criteria": [
            {
              "characteristic_index": 0,
              "thresholds": [
                {
                  "max": 100,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "E. Grade": {
      "criteria_sets": {
        "E. Grade": {
          "not_a_fruit": false
        }
      }
    }
  },
  "id": 117,
  "is_primary": true,
  "name": "GrannySmith DPC 2024",
  "node_id": 4,
  "schema": {
    "name": "spectrim_grade_map",
    "version": 2
  },
  "timestamp": "2025-05-26T07:51:39Z"
}`
	avo_gm_json = `{
  "characteristics": {
    "3": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 47,
      "index": 57,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Color Defect": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 10,
      "index": 11,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "1. Color Defect (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 0,
      "display_order": 11,
      "index": 12,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "1. Cuts": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 21,
      "index": 19,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Ground Rub": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 20,
      "index": 45,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "1. Insect Damage": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 37,
      "index": 61,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Leaf Roller": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 29,
      "index": 24,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Leaf Roller Count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 30,
      "index": 10,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Machine Damage": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 31,
      "index": 17,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Mc Damage Count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 32,
      "index": 47,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Nodules": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 35,
      "index": 51,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Puncture": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 22,
      "index": 20,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Rat Chew": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 27,
      "index": 16,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Rat Chew Count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 28,
      "index": 23,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Ridging": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 36,
      "index": 53,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Rot": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 26,
      "index": 22,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Scar": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 14,
      "index": 15,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "1. Scar 100 Px >": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 15,
      "index": 30,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Scar Big": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 19,
      "index": 43,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Scar Med": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 18,
      "index": 13,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Scar Small vein": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 17,
      "index": 7,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Scar size count": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 16,
      "index": 36,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Stem Damage": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 33,
      "index": 28,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. Sunburn/Rot %": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 23,
      "index": 21,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "1. Sunburn/Rot Lrg": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 24,
      "index": 29,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Sunburn/Rot Sml": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 25,
      "index": 52,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "1. Total Blemish": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 38,
      "index": 62,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "1. White Scales": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 34,
      "index": 50,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "2. Sheep Nose": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 43,
      "index": 35,
      "mode": "Count",
      "scale": [
        0
      ]
    },
    "3. Ground Rub": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 52,
      "index": 60,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "3. Netting": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 51,
      "index": 59,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "3. Ridging": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 49,
      "index": 58,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "3. Ridging (Good)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 50,
      "index": 25,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Black": {
      "category": "Fruit Color",
      "display_order": 4,
      "index": 5,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Brown": {
      "category": "Fruit Color",
      "display_order": 7,
      "index": 8,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Calyx Classification (1)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 13,
      "index": 49,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Calyx Classification (2)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 42,
      "index": 33,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Calyx Classification (3)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 2,
      "display_order": 48,
      "index": 56,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Class 36": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 44,
      "index": 63,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Curved Symmetry": {
      "category": "Shape",
      "display_order": 54,
      "index": 9,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Dark Green": {
      "category": "Fruit Color",
      "display_order": 0,
      "index": 2,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Defect Color 12": {
      "category": "Defect Color",
      "defect_grading_pass_index": 2,
      "display_order": 45,
      "index": 54,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Defect Color 12 (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 2,
      "display_order": 46,
      "index": 55,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "Defect Color 8": {
      "category": "Defect Color",
      "defect_grading_pass_index": 1,
      "display_order": 39,
      "index": 31,
      "mode": "Area (mm²)",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Defect Color 8 (count)": {
      "category": "Defect Color",
      "defect_grading_pass_index": 1,
      "display_order": 40,
      "index": 32,
      "min_value": [
        0,
        5,
        10
      ],
      "mode": "Area (mm²)",
      "scale": [
        0,
        0,
        0
      ]
    },
    "Double Factor": {
      "category": "Special",
      "display_order": 58,
      "index": 37,
      "mode": "Normal",
      "scale": [
        0
      ]
    },
    "Elongation": {
      "category": "Shape",
      "display_order": 55,
      "index": 18,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Green": {
      "category": "Fruit Color",
      "display_order": 1,
      "index": 1,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Green Avg (Light Green +)": {
      "category": "Color Function",
      "display_order": 60,
      "index": 42,
      "min_value": [
        0
      ],
      "mode": "Function",
      "scale": [
        0
      ]
    },
    "IR Variation (Largest Channel) (Light Green +)": {
      "category": "Color Function",
      "display_order": 62,
      "index": 27,
      "min_value": [
        0
      ],
      "mode": "Function",
      "scale": [
        0
      ]
    },
    "Light Green": {
      "category": "Fruit Color",
      "display_order": 2,
      "index": 0,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Lumpiness": {
      "category": "Shape",
      "display_order": 53,
      "index": 26,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Missed Grade": {
      "category": "Special",
      "display_order": 56,
      "index": 41,
      "mode": "Normal",
      "scale": [
        0
      ]
    },
    "Purple Ripening": {
      "category": "Fruit Color",
      "display_order": 3,
      "index": 3,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Slippage": {
      "category": "Special",
      "display_order": 59,
      "index": 39,
      "mode": "Normal",
      "scale": [
        0
      ]
    },
    "Softness 1 (v2) (Light Green +)": {
      "category": "Color Function",
      "display_order": 61,
      "index": 14,
      "min_value": [
        0
      ],
      "mode": "Function",
      "scale": [
        0
      ]
    },
    "Softness 2 (v2) (Light Green +)": {
      "category": "Color Function",
      "display_order": 63,
      "index": 38,
      "min_value": [
        0
      ],
      "mode": "Function",
      "scale": [
        0
      ]
    },
    "Stem Classification (1)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 0,
      "display_order": 12,
      "index": 48,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Stem Classification (2)": {
      "category": "Classified Blobs",
      "defect_grading_pass_index": 1,
      "display_order": 41,
      "index": 34,
      "mode": "Area (mm²)",
      "scale": [
        0
      ]
    },
    "Sticker": {
      "category": "Fruit Color",
      "display_order": 8,
      "index": 46,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Sunburn Red": {
      "category": "Fruit Color",
      "display_order": 5,
      "index": 44,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Sunburn Yellow": {
      "category": "Fruit Color",
      "display_order": 6,
      "index": 6,
      "mode": "%",
      "power": 50,
      "scale": [
        0
      ]
    },
    "Sunburn Yellow + Brown": {
      "category": "Color Combination",
      "display_order": 9,
      "index": 4,
      "mode": "%",
      "scale": [
        0
      ]
    },
    "Touching Factor": {
      "category": "Special",
      "display_order": 57,
      "index": 40,
      "mode": "Normal",
      "scale": [
        0
      ]
    }
  },
  "computer_name": "SPECTRIMMASTER",
  "defect_grading_passes": {
    "General Pass": {
      "index": 0
    },
    "Nett/Ridge": {
      "index": 2
    },
    "Sheeps Nose": {
      "index": 1
    }
  },
  "grades": {
    "A. Recycle": {
      "criteria_sets": {
        "A.0. Missed Grade": {
          "criteria": [
            {
              "characteristic_index": 41,
              "thresholds": [
                {
                  "max": 100,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "A.1. Touching Factor": {
          "criteria": [
            {
              "characteristic_index": 40,
              "thresholds": [
                {
                  "max": 100,
                  "min": 50
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "A.2. Double Factor": {
          "criteria": [
            {
              "characteristic_index": 37,
              "thresholds": [
                {
                  "max": 100,
                  "min": 70
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "B. Juice": {
      "criteria_sets": {
        "B.0. Tree Colour [TC]": {
          "criteria": [
            {
              "characteristic_index": 3,
              "thresholds": [
                {
                  "max": 100,
                  "min": 100
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.1. Sunburn Yellow + Brown [SUN]": {
          "criteria": [
            {
              "characteristic_index": 4,
              "thresholds": [
                {
                  "max": 100,
                  "min": 100
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.10. 1. Leaf Roller [LR]": {
          "criteria": [
            {
              "characteristic_index": 24,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 100
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.11. 1. Leaf Roller Cnt [LR]": {
          "criteria": [
            {
              "characteristic_index": 10,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 3
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.12. 1. Scar Big [SC]": {
          "criteria": [
            {
              "characteristic_index": 43,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 25
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.13. 1. Mc Damage Count G [MD]": {
          "criteria": [
            {
              "characteristic_index": 47,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 4
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.14. Sunburn Yellow + Brown [SUN]": {
          "criteria": [
            {
              "characteristic_index": 4,
              "thresholds": [
                {
                  "max": 100,
                  "min": 45
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.15. 1.White Scales [WS]": {
          "criteria": [
            {
              "characteristic_index": 50,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 25
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.16. Sunburn Yellow [SUN]": {
          "criteria": [
            {
              "characteristic_index": 6,
              "thresholds": [
                {
                  "max": 100,
                  "min": 45
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.17. 3. Ground Rub [GR]": {
          "criteria": [
            {
              "characteristic_index": 60,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 4000
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.18. 1. Insect damage [ID]": {
          "criteria": [
            {
              "characteristic_index": 61,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 7
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.19. Machine Damage [MD]": {
          "criteria": [
            {
              "characteristic_index": 17,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 300
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.2. 1. Scar [SC]": {
          "criteria": [
            {
              "characteristic_index": 15,
              "thresholds": [
                {
                  "max": 100,
                  "min": 60
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.3. 1. DiffuseScar [SC]": {
          "criteria": [
            {
              "characteristic_index": 45,
              "thresholds": [
                {
                  "max": 100,
                  "min": 50
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.4. 1. Cuts [CUT]": {
          "criteria": [
            {
              "characteristic_index": 19,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 5
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.5. 1. Sunburn Rot % [SUN]": {
          "criteria": [
            {
              "characteristic_index": 21,
              "thresholds": [
                {
                  "max": 100,
                  "min": 14.5
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.6. 1. Puncture [PUN]": {
          "criteria": [
            {
              "characteristic_index": 20,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 4
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.7. 1. Rot [ROT]": {
          "criteria": [
            {
              "characteristic_index": 22,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.8. 1. Rat Chew [RAT]": {
          "criteria": [
            {
              "characteristic_index": 16,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 400
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "B.9. 1. Rat Chew Cnt [RAT]": {
          "criteria": [
            {
              "characteristic_index": 23,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "C. Class 2": {
      "criteria_sets": {
        "C.0. Tree Colour [TC]": {
          "criteria": [
            {
              "characteristic_index": 3,
              "thresholds": [
                {
                  "max": 100,
                  "min": 100
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.1. Sunburn Yellow + Brown [SUN]": {
          "criteria": [
            {
              "characteristic_index": 4,
              "thresholds": [
                {
                  "max": 100,
                  "min": 100
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.10. 1. Puncture [PUN]": {
          "criteria": [
            {
              "characteristic_index": 20,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 3
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.11. 1.White Scales [WC]": {
          "criteria": [
            {
              "characteristic_index": 50,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 11
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.12. Sunburn Yellow [SUN]": {
          "criteria": [
            {
              "characteristic_index": 6,
              "thresholds": [
                {
                  "max": 100,
                  "min": 22
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.13. 3. Ground Rub [GD]": {
          "criteria": [
            {
              "characteristic_index": 60,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 200
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.14. 1. Insect Damage [ID]": {
          "criteria": [
            {
              "characteristic_index": 61,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 6
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.2. 1. Sunburn/Rot Lrg [SUN]": {
          "criteria": [
            {
              "characteristic_index": 29,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.3. 1. Scar Big [SC]": {
          "criteria": [
            {
              "characteristic_index": 43,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 9
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.4. 1. Scar Med [SC]": {
          "criteria": [
            {
              "characteristic_index": 13,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 5
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.5. 1. Scar Small vein [SC]": {
          "criteria": [
            {
              "characteristic_index": 7,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 30
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.6. 2. Sheep Nose [SN]": {
          "criteria": [
            {
              "characteristic_index": 35,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 5
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.7. 1. Stem Damage [SD]": {
          "criteria": [
            {
              "characteristic_index": 28,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 182
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.8. 1. Scar 100 Px > [SC]": {
          "criteria": [
            {
              "characteristic_index": 30,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 9000
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "C.9. 1. Rat Chew [RC]": {
          "criteria": [
            {
              "characteristic_index": 16,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 90
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "D. Class 1": {
      "criteria_sets": {
        "D.0. Tree colour [TC]": {
          "criteria": [
            {
              "characteristic_index": 3,
              "thresholds": [
                {
                  "max": 100,
                  "min": 100
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.1. Sunburn Yellow [SUN]": {
          "criteria": [
            {
              "characteristic_index": 6,
              "thresholds": [
                {
                  "max": 100,
                  "min": 22.100000381469727
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.10. 1. Puncture [PUN]": {
          "criteria": [
            {
              "characteristic_index": 20,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.11. 1. Sunburn [SUN]": {
          "criteria": [
            {
              "characteristic_index": 21,
              "thresholds": [
                {
                  "max": 100,
                  "min": 1.2000000476837158
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.12. 1. Sunburn/Rot Sml [SUN]": {
          "criteria": [
            {
              "characteristic_index": 52,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.13. 1. Stem Damage [SD]": {
          "criteria": [
            {
              "characteristic_index": 28,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 400
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.14. 1. Scar size count [SC]": {
          "criteria": [
            {
              "characteristic_index": 36,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 7
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.15. 1. Scar Small vein [SC]": {
          "criteria": [
            {
              "characteristic_index": 7,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 20
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.16. 1.White Scales [WC]": {
          "criteria": [
            {
              "characteristic_index": 50,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 5
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.17. 1.Ridging [RID]": {
          "criteria": [
            {
              "characteristic_index": 53,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 9999
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.18. 3. Ridging Good [RID]": {
          "criteria": [
            {
              "characteristic_index": 25,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 5000
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.19. 1. Insect damage [ID]": {
          "criteria": [
            {
              "characteristic_index": 61,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 3
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.2. 1.Nodules [ND]": {
          "criteria": [
            {
              "characteristic_index": 51,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1200
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.20. 3. Netting [NET]": {
          "criteria": [
            {
              "characteristic_index": 59,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 200
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.21. 1. Total Blemish [BLE]": {
          "criteria": [
            {
              "characteristic_index": 62,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1850
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.22. 1. Ground Rub [GD]": {
          "criteria": [
            {
              "characteristic_index": 45,
              "thresholds": [
                {
                  "max": 100,
                  "min": 2.700000047683716
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.23. 1. Leaf Roller [LR}": {
          "criteria": [
            {
              "characteristic_index": 24,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 59
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.3. 1. Scar 100px [SC]": {
          "criteria": [
            {
              "characteristic_index": 30,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2000
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.4. 1. Scar [SC}": {
          "criteria": [
            {
              "characteristic_index": 15,
              "thresholds": [
                {
                  "max": 100,
                  "min": 7
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.5. 1. Scar Small [SC]": {
          "criteria": [
            {
              "characteristic_index": 7,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 24
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.6. 1. Scar Med [SC]": {
          "criteria": [
            {
              "characteristic_index": 13,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 7
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.7. 2. Sheep Nose [SN]": {
          "criteria": [
            {
              "characteristic_index": 35,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 2
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.8. 1. Scar Big [SC]": {
          "criteria": [
            {
              "characteristic_index": 43,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        },
        "D.9. 1. Cuts [CUT]": {
          "criteria": [
            {
              "characteristic_index": 19,
              "thresholds": [
                {
                  "max": 9999,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "F. Premium": {
      "criteria_sets": {
        "F. Premium": {
          "criteria": [
            {
              "characteristic_index": 1,
              "thresholds": [
                {
                  "max": 100,
                  "min": 1
                }
              ]
            }
          ],
          "not_a_fruit": false
        }
      }
    },
    "G. Premium all": {
      "criteria_sets": {
        "G. Premium all": {
          "not_a_fruit": false
        }
      }
    }
  },
  "id": 113,
  "is_primary": true,
  "name": "New Hass July 23",
  "node_id": 4,
  "schema": {
    "name": "spectrim_grade_map",
    "version": 2
  },
  "timestamp": "2025-05-22T06:30:10Z"
}`
)
