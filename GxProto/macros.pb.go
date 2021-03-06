// Code generated by protoc-gen-go.
// source: macros.proto
// DO NOT EDIT!

package GxProto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// 性别.
type Sex int32

const (
	Sex_GC_SEX_BOY  Sex = 0
	Sex_GC_SEX_GIRL Sex = 1
)

var Sex_name = map[int32]string{
	0: "GC_SEX_BOY",
	1: "GC_SEX_GIRL",
}
var Sex_value = map[string]int32{
	"GC_SEX_BOY":  0,
	"GC_SEX_GIRL": 1,
}

func (x Sex) Enum() *Sex {
	p := new(Sex)
	*p = x
	return p
}
func (x Sex) String() string {
	return proto.EnumName(Sex_name, int32(x))
}
func (x *Sex) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(Sex_value, data, "Sex")
	if err != nil {
		return err
	}
	*x = Sex(value)
	return nil
}

// 可能的终端类型.
type TerminalType int32

const (
	TerminalType_GC_TT_IPHONE  TerminalType = 0
	TerminalType_GC_TT_IPAD    TerminalType = 1
	TerminalType_GC_TT_ANDROid TerminalType = 2
	TerminalType_GC_TT_WP      TerminalType = 3
	TerminalType_GC_TT_PC      TerminalType = 4
)

var TerminalType_name = map[int32]string{
	0: "GC_TT_IPHONE",
	1: "GC_TT_IPAD",
	2: "GC_TT_ANDROid",
	3: "GC_TT_WP",
	4: "GC_TT_PC",
}
var TerminalType_value = map[string]int32{
	"GC_TT_IPHONE":  0,
	"GC_TT_IPAD":    1,
	"GC_TT_ANDROid": 2,
	"GC_TT_WP":      3,
	"GC_TT_PC":      4,
}

func (x TerminalType) Enum() *TerminalType {
	p := new(TerminalType)
	*p = x
	return p
}
func (x TerminalType) String() string {
	return proto.EnumName(TerminalType_name, int32(x))
}
func (x *TerminalType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(TerminalType_value, data, "TerminalType")
	if err != nil {
		return err
	}
	*x = TerminalType(value)
	return nil
}

// 支持的接入平台类型
type PlatformType int32

const (
	PlatformType_GC_PT_RAW_GAS      PlatformType = 0
	PlatformType_GC_PT_PP_ASSISTANT PlatformType = 1
	PlatformType_GC_PT_KUAI_YONG    PlatformType = 2
	PlatformType_GC_PT_91_ASSISTANT PlatformType = 3
	PlatformType_GC_PT_TONGBUTUI    PlatformType = 4
	PlatformType_GC_PT_ITOOLS       PlatformType = 5
	PlatformType_GC_PT_06           PlatformType = 6
	PlatformType_GC_PT_07           PlatformType = 7
	PlatformType_GC_PT_08           PlatformType = 8
	PlatformType_GC_PT_09           PlatformType = 9
	PlatformType_GC_PT_0A           PlatformType = 10
	PlatformType_GC_PT_0B           PlatformType = 11
	PlatformType_GC_PT_0C           PlatformType = 12
	PlatformType_GC_PT_0D           PlatformType = 13
	PlatformType_GC_PT_0E           PlatformType = 14
	PlatformType_GC_PT_0F           PlatformType = 15
	PlatformType_GC_PT_10           PlatformType = 16
	PlatformType_GC_PT_11           PlatformType = 17
	PlatformType_GC_PT_12           PlatformType = 18
	PlatformType_GC_PT_13           PlatformType = 19
	PlatformType_GC_PT_14           PlatformType = 20
	PlatformType_GC_PT_15           PlatformType = 21
	PlatformType_GC_PT_16           PlatformType = 22
	PlatformType_GC_PT_17           PlatformType = 23
	PlatformType_GC_PT_18           PlatformType = 24
	PlatformType_GC_PT_19           PlatformType = 25
	PlatformType_GC_PT_1A           PlatformType = 26
	PlatformType_GC_PT_1B           PlatformType = 27
	PlatformType_GC_PT_1C           PlatformType = 28
	PlatformType_GC_PT_1D           PlatformType = 29
	PlatformType_GC_PT_1E           PlatformType = 30
	PlatformType_GC_PT_1F           PlatformType = 31
	PlatformType_GC_PT_20           PlatformType = 32
	PlatformType_GC_PT_21           PlatformType = 33
	PlatformType_GC_PT_22           PlatformType = 34
	PlatformType_GC_PT_23           PlatformType = 35
	PlatformType_GC_PT_24           PlatformType = 36
	PlatformType_GC_PT_25           PlatformType = 37
	PlatformType_GC_PT_26           PlatformType = 38
	PlatformType_GC_PT_27           PlatformType = 39
	PlatformType_GC_PT_28           PlatformType = 40
	PlatformType_GC_PT_29           PlatformType = 41
	PlatformType_GC_PT_2A           PlatformType = 42
	PlatformType_GC_PT_2B           PlatformType = 43
	PlatformType_GC_PT_2C           PlatformType = 44
	PlatformType_GC_PT_2D           PlatformType = 45
	PlatformType_GC_PT_2E           PlatformType = 46
	PlatformType_GC_PT_2F           PlatformType = 47
)

var PlatformType_name = map[int32]string{
	0:  "GC_PT_RAW_GAS",
	1:  "GC_PT_PP_ASSISTANT",
	2:  "GC_PT_KUAI_YONG",
	3:  "GC_PT_91_ASSISTANT",
	4:  "GC_PT_TONGBUTUI",
	5:  "GC_PT_ITOOLS",
	6:  "GC_PT_06",
	7:  "GC_PT_07",
	8:  "GC_PT_08",
	9:  "GC_PT_09",
	10: "GC_PT_0A",
	11: "GC_PT_0B",
	12: "GC_PT_0C",
	13: "GC_PT_0D",
	14: "GC_PT_0E",
	15: "GC_PT_0F",
	16: "GC_PT_10",
	17: "GC_PT_11",
	18: "GC_PT_12",
	19: "GC_PT_13",
	20: "GC_PT_14",
	21: "GC_PT_15",
	22: "GC_PT_16",
	23: "GC_PT_17",
	24: "GC_PT_18",
	25: "GC_PT_19",
	26: "GC_PT_1A",
	27: "GC_PT_1B",
	28: "GC_PT_1C",
	29: "GC_PT_1D",
	30: "GC_PT_1E",
	31: "GC_PT_1F",
	32: "GC_PT_20",
	33: "GC_PT_21",
	34: "GC_PT_22",
	35: "GC_PT_23",
	36: "GC_PT_24",
	37: "GC_PT_25",
	38: "GC_PT_26",
	39: "GC_PT_27",
	40: "GC_PT_28",
	41: "GC_PT_29",
	42: "GC_PT_2A",
	43: "GC_PT_2B",
	44: "GC_PT_2C",
	45: "GC_PT_2D",
	46: "GC_PT_2E",
	47: "GC_PT_2F",
}
var PlatformType_value = map[string]int32{
	"GC_PT_RAW_GAS":      0,
	"GC_PT_PP_ASSISTANT": 1,
	"GC_PT_KUAI_YONG":    2,
	"GC_PT_91_ASSISTANT": 3,
	"GC_PT_TONGBUTUI":    4,
	"GC_PT_ITOOLS":       5,
	"GC_PT_06":           6,
	"GC_PT_07":           7,
	"GC_PT_08":           8,
	"GC_PT_09":           9,
	"GC_PT_0A":           10,
	"GC_PT_0B":           11,
	"GC_PT_0C":           12,
	"GC_PT_0D":           13,
	"GC_PT_0E":           14,
	"GC_PT_0F":           15,
	"GC_PT_10":           16,
	"GC_PT_11":           17,
	"GC_PT_12":           18,
	"GC_PT_13":           19,
	"GC_PT_14":           20,
	"GC_PT_15":           21,
	"GC_PT_16":           22,
	"GC_PT_17":           23,
	"GC_PT_18":           24,
	"GC_PT_19":           25,
	"GC_PT_1A":           26,
	"GC_PT_1B":           27,
	"GC_PT_1C":           28,
	"GC_PT_1D":           29,
	"GC_PT_1E":           30,
	"GC_PT_1F":           31,
	"GC_PT_20":           32,
	"GC_PT_21":           33,
	"GC_PT_22":           34,
	"GC_PT_23":           35,
	"GC_PT_24":           36,
	"GC_PT_25":           37,
	"GC_PT_26":           38,
	"GC_PT_27":           39,
	"GC_PT_28":           40,
	"GC_PT_29":           41,
	"GC_PT_2A":           42,
	"GC_PT_2B":           43,
	"GC_PT_2C":           44,
	"GC_PT_2D":           45,
	"GC_PT_2E":           46,
	"GC_PT_2F":           47,
}

func (x PlatformType) Enum() *PlatformType {
	p := new(PlatformType)
	*p = x
	return p
}
func (x PlatformType) String() string {
	return proto.EnumName(PlatformType_name, int32(x))
}
func (x *PlatformType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(PlatformType_value, data, "PlatformType")
	if err != nil {
		return err
	}
	*x = PlatformType(value)
	return nil
}

// 运营版本.
type GameRunType int32

const (
	GameRunType_GC_GR_TEST   GameRunType = 0
	GameRunType_GC_GR_91     GameRunType = 1
	GameRunType_GC_GR_APPALE GameRunType = 2
	GameRunType_GC_GR_CW     GameRunType = 3
	GameRunType_GC_GR_EXT1   GameRunType = 4
	GameRunType_GC_GR_EXT2   GameRunType = 5
	GameRunType_GC_GR_EXT3   GameRunType = 6
	GameRunType_GC_GR_EXT4   GameRunType = 7
	GameRunType_GC_GR_EXT5   GameRunType = 8
	GameRunType_GC_GR_EXT6   GameRunType = 9
)

var GameRunType_name = map[int32]string{
	0: "GC_GR_TEST",
	1: "GC_GR_91",
	2: "GC_GR_APPALE",
	3: "GC_GR_CW",
	4: "GC_GR_EXT1",
	5: "GC_GR_EXT2",
	6: "GC_GR_EXT3",
	7: "GC_GR_EXT4",
	8: "GC_GR_EXT5",
	9: "GC_GR_EXT6",
}
var GameRunType_value = map[string]int32{
	"GC_GR_TEST":   0,
	"GC_GR_91":     1,
	"GC_GR_APPALE": 2,
	"GC_GR_CW":     3,
	"GC_GR_EXT1":   4,
	"GC_GR_EXT2":   5,
	"GC_GR_EXT3":   6,
	"GC_GR_EXT4":   7,
	"GC_GR_EXT5":   8,
	"GC_GR_EXT6":   9,
}

func (x GameRunType) Enum() *GameRunType {
	p := new(GameRunType)
	*p = x
	return p
}
func (x GameRunType) String() string {
	return proto.EnumName(GameRunType_name, int32(x))
}
func (x *GameRunType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(GameRunType_value, data, "GameRunType")
	if err != nil {
		return err
	}
	*x = GameRunType(value)
	return nil
}

func init() {
	proto.RegisterEnum("GxProto.Sex", Sex_name, Sex_value)
	proto.RegisterEnum("GxProto.TerminalType", TerminalType_name, TerminalType_value)
	proto.RegisterEnum("GxProto.PlatformType", PlatformType_name, PlatformType_value)
	proto.RegisterEnum("GxProto.GameRunType", GameRunType_name, GameRunType_value)
}
