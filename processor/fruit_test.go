package processor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"
)

const (
	c2_fruit_json = `{"carrierId": "0201020018A17DE8","sizer": {"schema": {"name": "sizer_fruit","version": 12},"status": "delivered","lane": 2,"frame": 1,"carrier": {"id": "0201020018A17DE8","hal": 77,"rod": 181,"cup": 685,"weight_g": 162.8,"wai_g": 0.18,"side": "down"},"batch": {"id": 6846,"guid": "6facce9f-a1f6-451e-9cb6-b2e2010db3f3","name": "1080222"},"variety": {"id": "28e18df9-f105-4533-be7f-3169a588928b","userid": 4,"name": "Pink lady"},"product": {"id": "89a05e1b-29a1-4f37-8507-5edabed9757b","name": "C1","packname": "Prepack  1kg"},"category": {"id": 0,"name": "Packed Fruit"},"outlet": {"id": 1,"name": "ST1","is_totalled": true},"size": {"id": 9,"name": "70"},"grade": {"id": 2,"name": "Class2"},"quality": {"id": 1,"name": "Q1"},"is_sampled": false,"area_cm2": 49.7,"weight_g": 183.6,"carton_equivalent": 0.0142857,"density_kg_m3": 801,"left_offset_g": -75.2,"major_mm": 78.5,"minor_mm": 80.2,"volume_ml": 229,"vision_grade": "B","vision_value": 170,"sensor": {},"timestamp": "2025-05-20T05:52:43.615Z"},"spectrim": {"carrier": {"hal": 48,"id": "0201020018A17DE8"},"center_offsets": {"Fruit Center Offset X Avg (mm)": -10.1615571975708,"Fruit Center Offset X Max (mm)": 6.6738724708557129,"Fruit Center Offset X Min (mm)": -25.084108352661133,"Fruit Center Offset Y Avg (mm)": -0.94157648086547852,"Fruit Center Offset Y Max (mm)": 3.5564899444580078,"Fruit Center Offset Y Min (mm)": -5.0483717918396},"classified_blob": {"1. Bitter Pit": 11,"1. Blemish": 38.88409423828125,"1. Leaf": 408.27804565429688,"1.Cork Scabbing": 2,"2. Bruise Juice Sm": 2,"2. Bruise SC Area": 56.766407012939453,"4. Clayx Cracking": 108.43728637695313,"4. Clayx Cracking count": 1,"4. Stem Puncture": 6},"color": {"1. IR Def": 540.296142578125,"2. IR2 Def": 360.9256591796875,"3. Sun Burn Def": 153.53758239746094,"Conveyor": 0,"Dark Red": 1.7886418104171753,"Defect Color 7": 79.016036987304688,"Leaf": 146.81230163574219,"Light red": 38.056072235107422,"Red": 6.3772172927856445,"Yellow": 52.304420471191406},"color_blob": {"Large 1. IR Def Blob Num": 1,"Large 2. IR2 Def Blob Num": 1,"Large Defect Color 7 Blob Num": 1,"Medium 1. IR Def Blob Num": 4,"Medium 2. IR2 Def Blob Num": 4,"Medium 3. Sun Burn Def Blob Num": 1,"Small 1. IR Def Blob Num": 13,"Small 2. IR2 Def Blob Num": 12,"Small 3. Sun Burn Def Blob Num": 9,"Small Defect Color 7 Blob Num": 1,"Total Large Blob Num": 1,"Total Medium Blob Num": 4,"Total Small Blob Num": 13},"computer_name": "SPECTRIM2","diameters": {"2D Orientation Box (Stem Based) Angle Avg": 100.00965881347656,"2D Orientation Box (Stem Based) Angle Max": 164.223876953125,"2D Orientation Box (Stem Based) Angle Min": 29.026142120361328,"2D Orientation Box (Stem Based) Height Avg": 79.6578140258789,"2D Orientation Box (Stem Based) Height Max": 83.2258529663086,"2D Orientation Box (Stem Based) Height Min": 76.15313720703125,"2D Orientation Box (Stem Based) Width Avg": 76.096122741699219,"2D Orientation Box (Stem Based) Width Max": 78.63720703125,"2D Orientation Box (Stem Based) Width Min": 73.025802612304688,"2D Orientation Box Angle Avg": 43.129554748535156,"2D Orientation Box Angle Max": 86.788314819335938,"2D Orientation Box Angle Min": 15.776121139526367,"2D Orientation Box Height Avg": 79.493362426757812,"2D Orientation Box Height Max": 83.2258529663086,"2D Orientation Box Height Min": 76.15313720703125,"2D Orientation Box Width Avg": 76.26055908203125,"2D Orientation Box Width Max": 81.280807495117188,"2D Orientation Box Width Min": 73.025802612304688,"Combined Equatorial Diameter Avg (mm)": 0,"Combined Equatorial Diameter Max (mm)": 0,"Combined Equatorial Diameter Min (mm)": 0,"Equatorial Diameter (Apples) Avg (mm)": 76.096122741699219,"Equatorial Diameter (Apples) Max (mm)": 78.63720703125,"Equatorial Diameter (Apples) Min (mm)": 73.025802612304688,"Equatorial Diameter (Elongated Stems) Avg (mm)": 77.068603515625,"Equatorial Diameter (Elongated Stems) Max (mm)": 77.226646423339844,"Equatorial Diameter (Elongated Stems) Min (mm)": 76.910560607910156,"Equatorial Diameter Avg (mm)": 78.499114990234375,"Equatorial Diameter Max (mm)": 81.118698120117188,"Equatorial Diameter Min (mm)": 75.8795394897461,"Experimental Diameter 1 Avg": 81.118698120117188,"Experimental Diameter 1 Max": 81.118698120117188,"Experimental Diameter 1 Min": 81.118698120117188,"Experimental Diameter 2 Avg": 81.118698120117188,"Experimental Diameter 2 Max": 81.118698120117188,"Experimental Diameter 2 Min": 81.118698120117188,"Horizontal Diameter Avg (mm)": 78.732269287109375,"Horizontal Diameter Avg (pixels)": 195.632568359375,"Horizontal Diameter Max (mm)": 81.507537841796875,"Horizontal Diameter Max (pixels)": 206.93513488769531,"Horizontal Diameter Min (mm)": 74.42645263671875,"Horizontal Diameter Min (pixels)": 185.85850524902344,"Images Used (Diameters)": 0,"Longest Horizontal Avg (mm)": 78.7685317993164,"Longest Horizontal Max (mm)": 81.507537841796875,"Longest Horizontal Min (mm)": 74.541336059570312,"Major Diameter (mm)": 78.499114990234375,"Max Diameter Avg (mm)": 81.118698120117188,"Max Diameter Max (mm)": 83.519607543945312,"Max Diameter Min (mm)": 77.968032836914062,"Min Diameter Avg (mm)": 75.8795394897461,"Min Diameter Max (mm)": 77.968620300292969,"Min Diameter Min (mm)": 73.010940551757812,"Minor Diameter (mm)": 80.173126220703125,"Perimeter Diameter Avg (mm)": 78.641471862792969,"Perimeter Diameter Max (mm)": 79.941604614257812,"Perimeter Diameter Min (mm)": 77.015205383300781,"Perpendicular Diameter Avg (mm)": 77.3313217163086,"Perpendicular Diameter Max (mm)": 80.173126220703125,"Perpendicular Diameter Min (mm)": 74.407806396484375,"Roundness Avg": 93.74444580078125,"Roundness Max": 98.632583618164062,"Roundness Min": 90.011917114257812,"Squareness (Kiwifruit EOE) Avg": 45.444751739501953,"Squareness (Kiwifruit EOE) Max": 49.940853118896484,"Squareness (Kiwifruit EOE) Min": 39.589729309082031,"Squareness (Kiwifruit) Avg": 45.444751739501953,"Squareness (Kiwifruit) Max": 49.940853118896484,"Squareness (Kiwifruit) Min": 39.589729309082031,"Squareness Avg": 95.654342651367188,"Squareness Max": 100,"Squareness Min": 89.589729309082031,"Stem Area (Elongated Stems) Avg": 16.928781509399414,"Stem Area (Elongated Stems) Max": 23.954299926757812,"Stem Area (Elongated Stems) Min": 9.9032630920410156,"Stem Area Avg": 0,"Stem Area Max": 0,"Stem Area Min": 0,"Stem Diameter (Apples) Avg (mm)": 79.6578140258789,"Stem Diameter (Apples) Max (mm)": 83.2258529663086,"Stem Diameter (Apples) Min (mm)": 76.15313720703125,"Stem Diameter (Elongated Stems) Avg (mm)": 79.2881088256836,"Stem Diameter (Elongated Stems) Max (mm)": 80.375717163085938,"Stem Diameter (Elongated Stems) Min (mm)": 78.20050048828125,"Stem Diameter Avg (mm)": 79.2881088256836,"Stem Diameter Max (mm)": 80.375717163085938,"Stem Diameter Min (mm)": 78.20050048828125,"Stem Direction (0-180) Avg": 98.9868392944336,"Stem Direction (0-180) Max": 175.5,"Stem Direction (0-180) Min": 3,"Stem Direction (Cherries) Avg": 98.9868392944336,"Stem Direction (Cherries) Max": 175.5,"Stem Direction (Cherries) Min": 3,"Stem Length (Elongated Stems) Avg (mm)": 2.2178459167480469,"Stem Length (Elongated Stems) Max (mm)": 2.2775039672851562,"Stem Length (Elongated Stems) Min (mm)": 2.1581878662109375,"Stem Length Avg (mm)": 2.2178459167480469,"Stem Length Max (mm)": 2.2775039672851562,"Stem Length Min (mm)": 2.1581878662109375,"Trapezoid Height Avg": 79.3222885131836,"Trapezoid Height Max": 82.925399780273438,"Trapezoid Height Min": 75.931831359863281,"Trapezoid Shoulder Avg": 76.344474792480469,"Trapezoid Shoulder Max": 78.169204711914062,"Trapezoid Shoulder Min": 73.084640502929688,"Vertical Diameter Avg (mm)": 79.053840637207031,"Vertical Diameter Avg (pixels)": 195.88578796386719,"Vertical Diameter Max (mm)": 81.081459045410156,"Vertical Diameter Max (pixels)": 200.64421081542969,"Vertical Diameter Min (mm)": 77.1736831665039,"Vertical Diameter Min (pixels)": 191.05311584472656},"features": {"Box Shape": 62.6961784362793,"Calyx Size": 313.6007080078125,"Concavity Count": 0,"Curvature": 0,"Curved Symmetry": 0,"Double Factor": 0,"Elongation": 6.6741628646850586,"Elongation (Apples)": 0,"End Over End (Kiwifruit)": 47.590534210205078,"Equatorial Diameter (Apples)": 0,"Experimental Shape Characteristic 1": 53.337078094482422,"Experimental Shape Characteristic 2": 53.337078094482422,"Flatness": 0,"Flatness (Deprecated)": 9.0821571350097656,"Flatness (Kiwifruit)": 0,"Flatness v2": 0,"Flatness v2 (EOE)": 0,"Grade Optimizer": 50,"Loose Skin (Onion)": 0,"Lumpiness": 0,"Orientation": 10.287500381469727,"Oversize Factor": 0,"Roundness": 90.011917114257812,"Roundness (Overall)": 93.744453430175781,"Slippage (Cherries)": 0,"Slippage (Kiwifruit)": 52.6315803527832,"Slippage Factor": 0,"Smoothness": 0,"Spots": 0,"Squareness": 0,"Squareness (Deprecated)": 52.282451629638672,"Squareness (Kiwifruit)": 0,"Squareness v2": 0,"Squareness v2 (EOE)": 0,"Squatness (Apples)": 0,"Stem Angle": 149.10572814941406,"Stem Area": 0,"Stem Area (Elongated Stems)": 23.954299926757812,"Stem Detection Error": 95,"Stem Diameter (Apples)": 0,"Stem In Top View": 100,"Stem Length": 2.2775039672851562,"Stem Length %": 2.8724408149719238,"Stem Length % (Elongated Stems)": 2.8724408149719238,"Stem Length (Elongated Stems)": 2.2775039672851562,"Stem Size": 408.27804565429688,"Symmetry": 0,"Touching Factor": 0,"Trapezoid Angle (Degrees)": 180,"Vertical Diameter Stability": 27.467977523803711,"Wobbliness (Kiwifruit)": 0},"function": {"Function Value": 170.03651428222656,"Softness 1 (v2)": 93.725830078125,"Softness 2 (v2)": 95.486557006835938},"grade": "B","grade_before_override": "@","grade_match": {"Grade A Match": 0,"Grade B Match": 99.998001098632812,"Grade C Match": 100},"grade_match_strong": {"Grade A Strong Match": 0,"Grade B Strong Match": 79.998001098632812,"Grade C Strong Match": 79.999992370605469},"grade_type": 0,"is_captured": false,"lane": 2,"log_level": 0,"node_id": 2,"processed_rotation_degrees": -360,"schema": {"name": "spectrim_fruit","version": 2},"skin_images_count": 38,"stem_detection_error": 1,"stem_in_top_view": 1132944720,"subgrade_index_1based": 12,"surface_area_mm2": 17522.30859375,"timer": {"Averaging radii time (ms)": 0,"Band color counting time (ms)": 0,"Blob Detection - Find initial blobs (ms)": 678.73809814453125,"Blob Detection - Total (ms)": 678.82904052734375,"Chain code time (ms)": 11.056600570678711,"Chaincode to radii time (ms)": 0,"Copy Captured Image": 0,"Detecting texture  (ms)": 159.29267883300781,"Doubles detection time (ms)": 0,"Filter diameters": 0,"Finding Diameters time (ms)": 0,"Shape and Size time (ms)": 43.070102691650391,"Total time for fruit processing": 903.93621826171875,"Uniformity correction (ms)": 1.7439001798629761,"Volume calculation time (ms)": 0},"timestamp": "2025-05-20T05:54:07Z","total_rotation_degrees": -384.35006713867188,"vision_grade_value": 170.03651428222656,"volume_ml": 229.03057861328125},"inspectra": null,"inspectraFruitData": null}`
)

func TestFromJson(t *testing.T) {
	c2_time, _ := time.Parse(time.RFC3339, "2025-05-20T05:52:43.615Z")
	c2_fruit := Fruit{
		SizerTime:        c2_time,
		CarrierId:        "0201020018A17DE8",
		SchemaVer:        12,
		Status:           "delivered",
		Lane:             2,
		Frame:            1,
		Rod:              181,
		Cup:              685,
		CupWeight:        162.8,
		CupWAI:           0.18,
		Side:             "down",
		BatchId:          6846,
		BatchName:        "1080222",
		BatchGuid:        "6facce9f-a1f6-451e-9cb6-b2e2010db3f3",
		VarietyName:      "Pink lady",
		VarietyGuid:      "28e18df9-f105-4533-be7f-3169a588928b",
		ProductName:      "C1",
		ProductGuid:      "89a05e1b-29a1-4f37-8507-5edabed9757b",
		ProductPack:      "Prepack  1kg",
		OutletName:       "ST1",
		OutletId:         1,
		OutletTotalled:   true,
		SizeName:         "70",
		SizeId:           9,
		GradeName:        "Class2",
		GradeId:          2,
		IsSampled:        false,
		Area:             17522.30859375,
		Weight:           183.6,
		CartonEquivalent: 0.0142857,
		Density:          801,
		LeftOffset:       -75.2,
		MajorDim:         78.5,
		MinorDim:         80.2,
		Volume:           229.03057861328125,
		VisionGrade:      "B",
		VisionValue:      170.03651428222656,

		RotationTotal:      -384.35006713867188,
		RotationProcessed:  -360,
		SubgradeIndex:      12,
		SkinImages:         38,
		StemDetectionError: 1,

		CenterOffsets:  []byte(`{"Fruit Center Offset X Avg (mm)":-10.1615571975708,"Fruit Center Offset X Max (mm)":6.673872470855713,"Fruit Center Offset X Min (mm)":-25.084108352661133,"Fruit Center Offset Y Avg (mm)":-0.9415764808654785,"Fruit Center Offset Y Max (mm)":3.556489944458008,"Fruit Center Offset Y Min (mm)":-5.0483717918396}`),
		ClassifiedBlob: []byte(`{"1. Bitter Pit":11,"1. Blemish":38.88409423828125,"1. Leaf":408.2780456542969,"1.Cork Scabbing":2,"2. Bruise Juice Sm":2,"2. Bruise SC Area":56.76640701293945,"4. Clayx Cracking":108.43728637695312,"4. Clayx Cracking count":1,"4. Stem Puncture":6}`),
		Colour:         []byte(`{"1. IR Def":540.296142578125,"2. IR2 Def":360.9256591796875,"3. Sun Burn Def":153.53758239746094,"Conveyor":0,"Dark Red":1.7886418104171753,"Defect Color 7":79.01603698730469,"Leaf":146.8123016357422,"Light red":38.05607223510742,"Red":6.3772172927856445,"Yellow":52.304420471191406}`),
		ColourBlob:     []byte(`{"Large 1. IR Def Blob Num":1,"Large 2. IR2 Def Blob Num":1,"Large Defect Color 7 Blob Num":1,"Medium 1. IR Def Blob Num":4,"Medium 2. IR2 Def Blob Num":4,"Medium 3. Sun Burn Def Blob Num":1,"Small 1. IR Def Blob Num":13,"Small 2. IR2 Def Blob Num":12,"Small 3. Sun Burn Def Blob Num":9,"Small Defect Color 7 Blob Num":1,"Total Large Blob Num":1,"Total Medium Blob Num":4,"Total Small Blob Num":13}`),
		Diameters:      []byte(`{"2D Orientation Box (Stem Based) Angle Avg":100.00965881347656,"2D Orientation Box (Stem Based) Angle Max":164.223876953125,"2D Orientation Box (Stem Based) Angle Min":29.026142120361328,"2D Orientation Box (Stem Based) Height Avg":79.6578140258789,"2D Orientation Box (Stem Based) Height Max":83.2258529663086,"2D Orientation Box (Stem Based) Height Min":76.15313720703125,"2D Orientation Box (Stem Based) Width Avg":76.09612274169922,"2D Orientation Box (Stem Based) Width Max":78.63720703125,"2D Orientation Box (Stem Based) Width Min":73.02580261230469,"2D Orientation Box Angle Avg":43.129554748535156,"2D Orientation Box Angle Max":86.78831481933594,"2D Orientation Box Angle Min":15.776121139526367,"2D Orientation Box Height Avg":79.49336242675781,"2D Orientation Box Height Max":83.2258529663086,"2D Orientation Box Height Min":76.15313720703125,"2D Orientation Box Width Avg":76.26055908203125,"2D Orientation Box Width Max":81.28080749511719,"2D Orientation Box Width Min":73.02580261230469,"Combined Equatorial Diameter Avg (mm)":0,"Combined Equatorial Diameter Max (mm)":0,"Combined Equatorial Diameter Min (mm)":0,"Equatorial Diameter (Apples) Avg (mm)":76.09612274169922,"Equatorial Diameter (Apples) Max (mm)":78.63720703125,"Equatorial Diameter (Apples) Min (mm)":73.02580261230469,"Equatorial Diameter (Elongated Stems) Avg (mm)":77.068603515625,"Equatorial Diameter (Elongated Stems) Max (mm)":77.22664642333984,"Equatorial Diameter (Elongated Stems) Min (mm)":76.91056060791016,"Equatorial Diameter Avg (mm)":78.49911499023438,"Equatorial Diameter Max (mm)":81.11869812011719,"Equatorial Diameter Min (mm)":75.8795394897461,"Experimental Diameter 1 Avg":81.11869812011719,"Experimental Diameter 1 Max":81.11869812011719,"Experimental Diameter 1 Min":81.11869812011719,"Experimental Diameter 2 Avg":81.11869812011719,"Experimental Diameter 2 Max":81.11869812011719,"Experimental Diameter 2 Min":81.11869812011719,"Horizontal Diameter Avg (mm)":78.73226928710938,"Horizontal Diameter Avg (pixels)":195.632568359375,"Horizontal Diameter Max (mm)":81.50753784179688,"Horizontal Diameter Max (pixels)":206.9351348876953,"Horizontal Diameter Min (mm)":74.42645263671875,"Horizontal Diameter Min (pixels)":185.85850524902344,"Images Used (Diameters)":0,"Longest Horizontal Avg (mm)":78.7685317993164,"Longest Horizontal Max (mm)":81.50753784179688,"Longest Horizontal Min (mm)":74.54133605957031,"Major Diameter (mm)":78.49911499023438,"Max Diameter Avg (mm)":81.11869812011719,"Max Diameter Max (mm)":83.51960754394531,"Max Diameter Min (mm)":77.96803283691406,"Min Diameter Avg (mm)":75.8795394897461,"Min Diameter Max (mm)":77.96862030029297,"Min Diameter Min (mm)":73.01094055175781,"Minor Diameter (mm)":80.17312622070312,"Perimeter Diameter Avg (mm)":78.64147186279297,"Perimeter Diameter Max (mm)":79.94160461425781,"Perimeter Diameter Min (mm)":77.01520538330078,"Perpendicular Diameter Avg (mm)":77.3313217163086,"Perpendicular Diameter Max (mm)":80.17312622070312,"Perpendicular Diameter Min (mm)":74.40780639648438,"Roundness Avg":93.74444580078125,"Roundness Max":98.63258361816406,"Roundness Min":90.01191711425781,"Squareness (Kiwifruit EOE) Avg":45.44475173950195,"Squareness (Kiwifruit EOE) Max":49.940853118896484,"Squareness (Kiwifruit EOE) Min":39.58972930908203,"Squareness (Kiwifruit) Avg":45.44475173950195,"Squareness (Kiwifruit) Max":49.940853118896484,"Squareness (Kiwifruit) Min":39.58972930908203,"Squareness Avg":95.65434265136719,"Squareness Max":100,"Squareness Min":89.58972930908203,"Stem Area (Elongated Stems) Avg":16.928781509399414,"Stem Area (Elongated Stems) Max":23.954299926757812,"Stem Area (Elongated Stems) Min":9.903263092041016,"Stem Area Avg":0,"Stem Area Max":0,"Stem Area Min":0,"Stem Diameter (Apples) Avg (mm)":79.6578140258789,"Stem Diameter (Apples) Max (mm)":83.2258529663086,"Stem Diameter (Apples) Min (mm)":76.15313720703125,"Stem Diameter (Elongated Stems) Avg (mm)":79.2881088256836,"Stem Diameter (Elongated Stems) Max (mm)":80.37571716308594,"Stem Diameter (Elongated Stems) Min (mm)":78.20050048828125,"Stem Diameter Avg (mm)":79.2881088256836,"Stem Diameter Max (mm)":80.37571716308594,"Stem Diameter Min (mm)":78.20050048828125,"Stem Direction (0-180) Avg":98.9868392944336,"Stem Direction (0-180) Max":175.5,"Stem Direction (0-180) Min":3,"Stem Direction (Cherries) Avg":98.9868392944336,"Stem Direction (Cherries) Max":175.5,"Stem Direction (Cherries) Min":3,"Stem Length (Elongated Stems) Avg (mm)":2.217845916748047,"Stem Length (Elongated Stems) Max (mm)":2.2775039672851562,"Stem Length (Elongated Stems) Min (mm)":2.1581878662109375,"Stem Length Avg (mm)":2.217845916748047,"Stem Length Max (mm)":2.2775039672851562,"Stem Length Min (mm)":2.1581878662109375,"Trapezoid Height Avg":79.3222885131836,"Trapezoid Height Max":82.92539978027344,"Trapezoid Height Min":75.93183135986328,"Trapezoid Shoulder Avg":76.34447479248047,"Trapezoid Shoulder Max":78.16920471191406,"Trapezoid Shoulder Min":73.08464050292969,"Vertical Diameter Avg (mm)":79.05384063720703,"Vertical Diameter Avg (pixels)":195.8857879638672,"Vertical Diameter Max (mm)":81.08145904541016,"Vertical Diameter Max (pixels)":200.6442108154297,"Vertical Diameter Min (mm)":77.1736831665039,"Vertical Diameter Min (pixels)":191.05311584472656}`),
		Features:       []byte(`{"Box Shape":62.6961784362793,"Calyx Size":313.6007080078125,"Concavity Count":0,"Curvature":0,"Curved Symmetry":0,"Double Factor":0,"Elongation":6.674162864685059,"Elongation (Apples)":0,"End Over End (Kiwifruit)":47.59053421020508,"Equatorial Diameter (Apples)":0,"Experimental Shape Characteristic 1":53.33707809448242,"Experimental Shape Characteristic 2":53.33707809448242,"Flatness":0,"Flatness (Deprecated)":9.082157135009766,"Flatness (Kiwifruit)":0,"Flatness v2":0,"Flatness v2 (EOE)":0,"Grade Optimizer":50,"Loose Skin (Onion)":0,"Lumpiness":0,"Orientation":10.287500381469727,"Oversize Factor":0,"Roundness":90.01191711425781,"Roundness (Overall)":93.74445343017578,"Slippage (Cherries)":0,"Slippage (Kiwifruit)":52.6315803527832,"Slippage Factor":0,"Smoothness":0,"Spots":0,"Squareness":0,"Squareness (Deprecated)":52.28245162963867,"Squareness (Kiwifruit)":0,"Squareness v2":0,"Squareness v2 (EOE)":0,"Squatness (Apples)":0,"Stem Angle":149.10572814941406,"Stem Area":0,"Stem Area (Elongated Stems)":23.954299926757812,"Stem Detection Error":95,"Stem Diameter (Apples)":0,"Stem In Top View":100,"Stem Length":2.2775039672851562,"Stem Length %":2.872440814971924,"Stem Length % (Elongated Stems)":2.872440814971924,"Stem Length (Elongated Stems)":2.2775039672851562,"Stem Size":408.2780456542969,"Symmetry":0,"Touching Factor":0,"Trapezoid Angle (Degrees)":180,"Vertical Diameter Stability":27.46797752380371,"Wobbliness (Kiwifruit)":0}`),
		Function:       []byte(`{"Function Value":170.03651428222656,"Softness 1 (v2)":93.725830078125,"Softness 2 (v2)":95.48655700683594}`),
		Timer:          []byte(`{"Averaging radii time (ms)":0,"Band color counting time (ms)":0,"Blob Detection - Find initial blobs (ms)":678.7380981445312,"Blob Detection - Total (ms)":678.8290405273438,"Chain code time (ms)":11.056600570678711,"Chaincode to radii time (ms)":0,"Copy Captured Image":0,"Detecting texture  (ms)":159.2926788330078,"Doubles detection time (ms)":0,"Filter diameters":0,"Finding Diameters time (ms)":0,"Shape and Size time (ms)":43.07010269165039,"Total time for fruit processing":903.9362182617188,"Uniformity correction (ms)":1.743900179862976,"Volume calculation time (ms)":0}`),

		ProcessingTime: time.Duration(903936218),
		PrimaryDefect:  "",
		OtherDefects:   nil,
	}

	type args struct {
		val []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Fruit
		wantErr bool
	}{
		{"Class 2", args{[]byte(c2_fruit_json)}, c2_fruit, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Fruit
			err := json.Unmarshal(tt.args.val, &got)
			// got, err := FromJson(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := ValidateFruit(got, tt.want); err != nil {
				t.Errorf("Unmarshal = %s", err)
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("FromJson() = %#v, want %#v", got, tt.want)
			// }
		})
	}
}

func ValidateFruit(x Fruit, y Fruit) (err error) {
	structType := reflect.TypeOf(x)
	structVal := reflect.ValueOf(x)
	structVal2 := reflect.ValueOf(y)
	fieldNum := structVal.NumField()

	for i := range fieldNum {
		field1 := structVal.Field(i)
		field2 := structVal2.Field(i)
		fieldName := structType.Field(i).Name

		if field1.Kind() == reflect.Slice {
			// field1.
			if field1.CanConvert(reflect.SliceOf(reflect.TypeOf(""))) {
				v1 := field1.Interface().([]string)
				v2 := field2.Interface().([]string)
				// field1.
				if !reflect.DeepEqual(v1, v2) {
					slog.Info("What", "field", fieldName, "f1", v1, "f2", v2)
					err = fmt.Errorf("%v%s does not match; ", err, fieldName)
				}
			} else {
				v1 := field1.Interface().([]byte)
				v2 := field2.Interface().([]byte)
				if !reflect.DeepEqual(v1, v2) {
					slog.Info("What", "field", fieldName, "f1", v1, "f2", v2)
					err = fmt.Errorf("%v%s does not match; ", err, fieldName)
				}
			}
		} else {
			if !field1.Equal(field2) {
				slog.Info("What", "f1", field1, "f2", field2)
				err = fmt.Errorf("%v%s does not match; ", err, fieldName)
			}
		}
	}

	return err
}
