package core

// Mantıksal taş konumu (oyun kuralları bazlı)
type LogicalCoordinate struct {
	//Tavlada 24 nokta (PointIndex) var. Üst sıra (12–23): X = 0–11, Alt sıra (11–0): X = 0–11
	//Y: Bir noktadaki taş yığınının yüksekliği. En alttaki taş: Y = 0, İkinci taş: Y = 1, vs.
	PointIndex   int // 0–23 arası, -1 olabilir.
	PositionType int
	Player       int   // 1 veya 2
	StackIndex   int   // 👈 yığındaki sırası (en alttaki 0)
	IsTop        bool  // 👈 yığının en üstündeki taş mı?
	MoveOrder    int64 // Taşın hareket sırası => Bir yigina son eklenen en tepede olacak..StackIndex en usteki olacak..
}

// Grafiksel (X,Y) koordinat sistemi – UI ya da görselleştirme için
type VisualCoordinate struct {
	X int // 0–11 arası (noktalar)
	Y int // Yığındaki konum (örneğin taş yüksekliği)
}

// X'ten pointIndex'e dönüşüm fonksiyonu => UI'dan gelen degere gore kordinalandirma icin..
func VisualToPointIndex(x int, isTop bool) int {
	if isTop {
		return 12 + x // üst sıra: 12–23
	}
	return 11 - x // alt sıra: 11–0
}
