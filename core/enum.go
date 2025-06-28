package core

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
