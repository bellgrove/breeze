package processor

import (
	"encoding/json"
	"testing"
)

func TestGrademap_Grade(t *testing.T) {
	type args struct {
		fr_json string
		gm_json string
	}
	tests := []struct {
		name string
		g    Grademap
		args args
	}{
		{"Pink lady C2", Grademap{}, args{c2_fruit_json, gsm_gm_json}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.g.Update([]byte(tt.args.gm_json))
			if err != nil {
				t.Errorf("FromJson() error = %v", err)
				return
			}
			var f Fruit
			json.Unmarshal([]byte(tt.args.fr_json), &f)
			// f.VisionGrade = "B"
			tt.g.Grade(&f)

			if f.PrimaryDefect != "SC" {
				t.Errorf("Grade() = %#v, want %#v", f.PrimaryDefect, "SC")
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

const (
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
