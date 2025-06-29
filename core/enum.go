package core

import "reflect"

var PositionTypeEnum = newPositionType()

func newPositionType() *positionType {
	return &positionType{
		Point:    1, // 0–23 numaralı noktalar
		Bar:      2, // Kırılan taşların beklediği yer
		OffBoard: 3, // Oyundan çıkan taşlar
	}
}

type positionType struct {
	Point    int
	Bar      int
	OffBoard int
}

func GetEnumName(enum interface{}, value int) string {
	v := reflect.ValueOf(enum).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		if int(v.Field(i).Int()) == value {
			return t.Field(i).Name
		}
	}
	return "Unknown"
}
