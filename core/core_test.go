package core_test

import (
	"backgammon/core"
	"fmt"
	"log"
	"reflect"
	"slices"
	"testing"
)

func TestMoveTopStoneAndUpdate_ValidMove_WithConsoleOutput(t *testing.T) {
	stones := core.GetInitialStones()

	player := 1
	fromPoint := 0      // Player 1 hareket etmek istediği yer
	toPoint := 3        // Hedef nokta (Geceli acik)
	dice := []int{1, 2} // Zarlar

	//Once kirik tasi var mi kontrolu
	result := core.IsAllBarEntryAllowed(stones, player, dice)

	if result.FromBar {
		t.Error("Bar'da taş var gozukuyor ama yok, FromBar false olmalı")
	}
	//Kirik tasi yok. Normal hareket kontrolu
	result = core.IsNormalMoveAllowed(stones, player, fromPoint, toPoint, dice)
	if !result.Allowed {
		t.Fatalf("Normal hareket izin verilmedi ama verilmesi bekleniyordu: %d -> %d", fromPoint, toPoint)
	}

	updatedStones, moved, broken := core.MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
	if !moved {
		t.Errorf("Taş hareket etmedi ama hareket mümkün olmalıydı.")
	}

	//Kirilan taslar yazdirilir..
	if moved && len(broken) > 0 {
		log.Printf("Player %d kırdı: %+v", player, broken[0].PointIndex)
	}

	t.Logf("Taş başarıyla hareket etti: Player %d, %d -> %d", player, fromPoint, toPoint)

	core.SortStonesByPlayerPointAndStackDesc(updatedStones)
	t.Log("Taşların Son Durumu:")
	for _, stone := range updatedStones {
		t.Logf("PointIndex: %d, Player: %d, StackIndex: %d, IsTop: %v, MoveOrder: %d",
			stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
	}
}

func TestMoveTopStoneAndUpdate_InvalidMove(t *testing.T) {
	stones := core.GetInitialStones()

	player := 1
	fromPoint := 0
	toPoint := 5        // Burada Player2'nin 5'de, 5 taşı var, geçersiz hareket.
	dice := []int{1, 4} // Zarlar

	//Once kirik tasi var mi kontrolu
	result := core.IsAllBarEntryAllowed(stones, player, dice)

	if result.FromBar {
		t.Error("Bar'da taş var gozukuyor ama yok, FromBar false olmalı")
	}
	//Kirik tasi yok. Normal hareket kontrolu. Ama hareket edememesi lazim..
	result = core.IsNormalMoveAllowed(stones, player, fromPoint, toPoint, dice)
	if result.Allowed {
		t.Fatalf("Normal hareket izin verilmemeli idi ama izin verildi: %d -> %d", fromPoint, toPoint)
	}

	// Geçersiz hamle olduğundan MoveTopStoneAndUpdate çağrılmamalı
	// Bu satıra kadar gelindiyse test başarılıdır
	t.Logf("Geçersiz hareket 'IsNormalMoveAllowed()' ile doğru şekilde engellendi: Player %d, %d -> %d", player, fromPoint, toPoint)

	//Burada moved zaten false gelmeli. Bakalim 2. kontrol calisiyor mu ?
	updatedStones, moved, _ := core.MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
	if moved {
		t.Errorf("Taş hareket etti ama hareket yasak olmalıydı.")
	} else {
		t.Log("MoveTopStoneAndUpdate() icindeki kontroller calisti ve hareket doğru şekilde engellendi")
	}

	// Taşın konumu değişmemeli
	for _, s := range updatedStones {
		if s.Player == player && s.PointIndex == toPoint {
			t.Errorf("Taş yanlışlıkla rakip taşların olduğu noktaya taşındı.")
		}
	}
}

func TestMoveTopStoneAndUpdate_CaptureOpponentStoneAndSendToBar(t *testing.T) {
	stones := []*core.LogicalCoordinate{
		// Player 1’in Point 3’teki tek taşı (kırılacak)
		{
			Player:       1,
			PointIndex:   3,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   0,
			IsTop:        true,
			MoveOrder:    0,
		},
		// Player 2’nin Point 7’deki taşı (hareket edecek)
		{
			Player:       2,
			PointIndex:   7,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   2,
			IsTop:        true,
			MoveOrder:    0,
		},
		{
			Player:       2,
			PointIndex:   7,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   1,
			IsTop:        false,
			MoveOrder:    0,
		},
		{
			Player:       2,
			PointIndex:   7,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   0,
			IsTop:        false,
			MoveOrder:    0,
		},
	}

	player := 2
	fromPoint := 7
	toPoint := 3
	dice := []int{4, 2}

	//Before
	t.Log("BEFORE")
	core.SortStonesByPlayerPointAndStackDesc(stones)
	t.Log("Taşların Son Durumu:")
	for _, stone := range stones {
		t.Logf("PointIndex: %d, Player: %d, StackIndex: %d, IsTop: %v, PositionType: %v, MoveOrder: %d",
			stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, core.GetEnumName(core.PositionTypeEnum, stone.PositionType), stone.MoveOrder)
	}
	t.Log("----------------------------------------------")

	// Önce bar'da taş var mı kontrolü
	result := core.IsAllBarEntryAllowed(stones, player, dice)
	if result.FromBar {
		t.Error("Bar'da taş olduğu görünüyor ama olmamalı, FromBar false olmalı")
	}

	// Normal hamle izni kontrolü
	result = core.IsNormalMoveAllowed(stones, player, fromPoint, toPoint, dice)
	if !result.Allowed {
		t.Fatalf("Normal hareket izni verilmedi ama verilmeliydi: %d -> %d", fromPoint, toPoint)
	}

	// Hareketi uygula
	updatedStones, moved, broken := core.MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
	if !moved {
		t.Fatalf("Taş hareket etmedi ama mümkün olmalıydı: %d -> %d", fromPoint, toPoint)
	}

	//Kirilan taslar yazdirilir..
	if moved && len(broken) > 0 {
		log.Printf("Player %d kırdı: %+v", player, broken[0].PointIndex)
	}

	// Kırılan taş bar’a gitmiş mi?
	barCount := 0
	for _, stone := range updatedStones {
		if stone.Player == 1 && stone.PositionType == core.PositionTypeEnum.Bar {
			barCount++
		}
	}
	if barCount != 1 {
		t.Errorf("Kırılan taş bar'a gönderilmedi. Beklenen: 1, Gerçek: %d", barCount)
	}

	// Hedef noktada Player 2'nin taşı var mı?
	targetOccupied := false
	for _, stone := range updatedStones {
		if stone.Player == 2 && stone.PointIndex == toPoint {
			targetOccupied = true
			break
		}
	}
	if !targetOccupied {
		t.Errorf("Player 2'nin taşı hedef noktaya yerleştirilmedi: %d", toPoint)
	}

	t.Logf("Taş başarıyla hareket etti ve rakip taşı kırdı: Player %d, %d -> %d", player, fromPoint, toPoint)

	t.Log("AFTER")
	core.SortStonesByPlayerPointAndStackDesc(updatedStones)
	t.Log("Taşların Son Durumu:")
	for _, stone := range updatedStones {
		t.Logf("PointIndex: %d, Player: %d, StackIndex: %d, IsTop: %v, PositionType: %v, MoveOrder: %d",
			stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, core.GetEnumName(core.PositionTypeEnum, stone.PositionType), stone.MoveOrder)
	}
}

// Player 2 nin kirilan tasi 1 ile girebilir ama 6 ile giremez cunku Player1'in tasi var..
func TestCanEnterFromBar_Player2_WithBrokenStone(t *testing.T) {
	stones := core.GetInitialStones()

	// Player 2'nin 23 noktasındaki taşlardan 1'ini Bar'a taşı (kırık)
	for i, tile := range stones {
		if tile.Player == 2 && tile.PointIndex == 23 {
			stones[i].PositionType = core.PositionTypeEnum.Bar
			stones[i].PointIndex = -1 // Bar için özel PointIndex
			break
		}
	}

	player := 2
	dice := []int{1, 6}

	//canEnter, enterableDice := core.CanEnterFromBar(stones, player, dice)
	canEnter, enterableDice := core.CanAllBarStonesEnter(stones, player, dice)
	if !canEnter {
		t.Error("Player 2 should be able to enter from bar with at least one dice, but can't")
	}

	// Sadece 1 ile giriş mümkün olmalı
	if len(enterableDice) != 1 || enterableDice[0] != 1 {
		t.Errorf("Expected enterable dice [1], got %v", enterableDice)
	} else {
		t.Logf("PASS: Player 2 can enter from bar only with dice: %v", enterableDice)
	}
}

// Player 2 kirik tasi ile, Player 1'in 23'deki tasini => 1,1 double atarak kirar..
func TestCanEnterFromBar_Player2CanCaptureSingleOpponent(t *testing.T) {
	stones := []*core.LogicalCoordinate{}

	// Player 1: sadece 1 taşı 23. noktaya koy
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   23,
		PositionType: core.PositionTypeEnum.Point,
		Player:       1,
		StackIndex:   0,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Player 2: bar'da bir taş
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   -1,
		PositionType: core.PositionTypeEnum.Bar,
		Player:       2,
		StackIndex:   0,
		IsTop:        true,
		MoveOrder:    0,
	})

	dice := core.ExpandDice([]int{1, 1})
	//canEnter, diceList := core.CanEnterFromBar(stones, 2, dice)
	canEnter, diceList := core.CanAllBarStonesEnter(stones, 2, dice)

	if !canEnter {
		t.Errorf("Player 2 bar’dan 1 zar ile giriş yapabilmeli (23. noktadaki tek rakip taşı kırarak), ama yapamıyor.")
	} else {
		t.Logf("PASS: Player 2 bar’dan %v ile girerek rakibin taşını kırabilir.", diceList)
	}
}

func TestIsNormalMoveAllowed_NormalMove(t *testing.T) {
	stones := core.GetInitialStones()
	player := 1
	from := 0
	to := 1
	dice := []int{1, 2}

	// Önce bar girişi kontrol edilir
	barResult := core.IsAllBarEntryAllowed(stones, player, dice)
	if barResult.FromBar {
		t.Error("Bar'da taş yok, FromBar false olmalı")
	}

	// Bar'da taş yoksa normal hamleye bakılır
	normalResult := core.IsNormalMoveAllowed(stones, player, from, to, dice)
	if !normalResult.CanMoveNormally {
		t.Error("Normal hamle yapılabilir durumda olmalı")
	}
	if !normalResult.Allowed {
		t.Error("Hareket izinli olmalı")
	}
}

func TestIsNormalMoveAllowedWithDistanceCheck_Allowed(t *testing.T) {
	player := 1
	stones := []*core.LogicalCoordinate{
		{PointIndex: 5, Player: player, IsTop: true},
		// Diğer taşlar...
	}
	dice := []int{2, 3}
	from := 5
	to := 10 // 5 + 2 + 3 = 10

	result := core.IsNormalMoveAllowed(stones, player, from, to, dice)

	if !result.Allowed {
		t.Errorf("Bekleniyor: hareket izinli, ancak izin verilmedi")
	}
}

func TestIsNormalMoveAllowed_NotAllowed(t *testing.T) {
	player := 1
	stones := []*core.LogicalCoordinate{
		{PointIndex: 5, Player: player, IsTop: true},
		// Diğer taşlar...
	}
	dice := []int{2, 3}
	from := 5
	to := 11 // 5 + 2 + 3 = 10, 11 olamaz

	result := core.IsNormalMoveAllowed(stones, player, from, to, dice)

	if result.Allowed {
		t.Errorf("Bekleniyor: hareket izinli değil, ama izin verildi")
	}
}

func TestIsBarEntryAllowed_PartialEntry(t *testing.T) {
	stones := []*core.LogicalCoordinate{}
	player := 1

	// Bar'da player 1 taşı var
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   -1,
		PositionType: core.PositionTypeEnum.Bar,
		Player:       player,
		StackIndex:   0,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Entry 0: player 2'nin 2 taşı (kapalı)
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   0,
		PositionType: core.PositionTypeEnum.Point,
		Player:       2,
		StackIndex:   0,
		IsTop:        false,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   0,
		PositionType: core.PositionTypeEnum.Point,
		Player:       2,
		StackIndex:   1,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Entry 1: boş
	dice := []int{1, 2}

	result := core.IsAllBarEntryAllowed(stones, player, dice)

	if !result.FromBar {
		t.Error("Bar'da taş var, FromBar true olmalı")
	}

	expectedDice := []int{2} // Sadece 2 ile giriş mümkün

	if !reflect.DeepEqual(result.EnterableDice, expectedDice) {
		t.Errorf("Beklenen giriş zarları: %v, gelen: %v", expectedDice, result.EnterableDice)
	}

	if !result.Allowed {
		t.Error("En az bir zarla giriş mümkünken Allowed true olmalı")
	}
}

// Cift Zar atildiginde [3,3] => 3 kirik tas PointIndex 2'ye girilebilir.
func TestIsBarEntryAllowed_NoEntryForThreeBrokenStones(t *testing.T) {
	stones := []*core.LogicalCoordinate{}
	player := 1

	// 3 adet kırık taş (Bar'da)
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   -1,
			PositionType: core.PositionTypeEnum.Bar,
			Player:       player,
			StackIndex:   i,
			IsTop:        i == 2, // sadece en üst taş IsTop true
			MoveOrder:    0,
		})
	}

	// Player 2, giriş noktalarını kapatıyor: Entry 0 ve Entry 1
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   0,
		PositionType: core.PositionTypeEnum.Point,
		Player:       2,
		StackIndex:   0,
		IsTop:        false,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   0,
		PositionType: core.PositionTypeEnum.Point,
		Player:       2,
		StackIndex:   1,
		IsTop:        true,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   1,
		PositionType: core.PositionTypeEnum.Point,
		Player:       2,
		StackIndex:   0,
		IsTop:        false,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   1,
		PositionType: core.PositionTypeEnum.Point,
		Player:       2,
		StackIndex:   1,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Çift zarlar (double değil)
	dice := core.ExpandDice([]int{3, 3})
	//dice := []int{3, 3}

	result := core.IsAllBarEntryAllowed(stones, player, dice)

	if !result.FromBar {
		t.Error("Bar'da taş var, FromBar true olmalı")
	}

	if !result.Allowed {
		t.Error("Hiçbir taş giremeyecekken Allowed false olmalı")
	}

	if len(result.EnterableDice) > 0 && result.EnterableDice[0] != 3 {
		t.Errorf("Giriş yapılabilecek sadece zar 3 olmalı, gelen: %v", result.EnterableDice)
	}

	t.Logf("Giris Yapilabilme Sonucu:%v Giris Yapılabılen zarlar: %v", result.Allowed, result.EnterableDice)
}

func TestIsBarEntryAllowed_NoEntry(t *testing.T) {
	stones := []*core.LogicalCoordinate{}
	player := 1

	// Bar'da player 1 taşı var
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   -1,
		PositionType: core.PositionTypeEnum.Bar,
		Player:       player,
		StackIndex:   0,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Entry 0: player 2'nin 2 taşı (kapalı)
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   0,
		PositionType: core.PositionTypeEnum.Point,
		Player:       2,
		StackIndex:   0,
		IsTop:        false,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   0,
		PositionType: core.PositionTypeEnum.Point,
		Player:       2,
		StackIndex:   1,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Zarlar sadece 1 (entry 0'a giriş, kapalı)
	//dice := []int{1, 1}
	// Çift zarlar (double değil)
	dice := core.ExpandDice([]int{1, 1})
	result := core.IsAllBarEntryAllowed(stones, player, dice)

	if !result.FromBar {
		t.Error("Bar'da taş var, FromBar true olmalı")
	}
	if result.CanEnterFromBar {
		t.Error("Giriş mümkün olmamalı çünkü entry noktası kapalı")
	}
	if result.Allowed {
		t.Error("Giriş engelliyse Allowed false olmalı")
	}
	if len(result.EnterableDice) != 0 {
		t.Errorf("Giriş yapılamayan zara rağmen enterableDice boş olmalı, ama: %v", result.EnterableDice)
	}
}

func TestPlayer2CannotMoveToBlockedPoint(t *testing.T) {
	stones := []*core.LogicalCoordinate{}

	player1 := 1
	player2 := 2

	// Player1'in 3 noktasında 2 taşı var (blokaj)
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   3,
		PositionType: core.PositionTypeEnum.Point,
		Player:       player1,
		StackIndex:   0,
		IsTop:        false,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   3,
		PositionType: core.PositionTypeEnum.Point,
		Player:       player1,
		StackIndex:   1,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Player1'in 7 noktasında 3 taşı var (örnek başka blokaj noktası)
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   7,
		PositionType: core.PositionTypeEnum.Point,
		Player:       player1,
		StackIndex:   0,
		IsTop:        false,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   7,
		PositionType: core.PositionTypeEnum.Point,
		Player:       player1,
		StackIndex:   1,
		IsTop:        false,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   7,
		PositionType: core.PositionTypeEnum.Point,
		Player:       player1,
		StackIndex:   2,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Player2'nin 10 noktasında 2 taşı var
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   10,
		PositionType: core.PositionTypeEnum.Point,
		Player:       player2,
		StackIndex:   0,
		IsTop:        false,
		MoveOrder:    0,
	})
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   10,
		PositionType: core.PositionTypeEnum.Point,
		Player:       player2,
		StackIndex:   1,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Taşların durumunu yazdır
	for _, s := range stones {
		fmt.Printf("Player %d taş: PointIndex=%d, StackIndex=%d, IsTop=%t\n", s.Player, s.PointIndex, s.StackIndex, s.IsTop)
	}

	from := 10          // Player 2 hareket etmek istediği yer
	to := 3             // Hedef nokta (blokajlı)
	dice := []int{3, 4} // Zarlar

	//Once kirik tasi var mi kontrolu
	result := core.IsAllBarEntryAllowed(stones, player2, dice)

	if result.FromBar {
		t.Error("Bar'da taş var gozukuyor ama yok, FromBar false olmalı")
	}
	//----------------

	//Kirik tasi yok ise normal hareket kontrolu
	result = core.IsNormalMoveAllowed(stones, player2, from, to, dice)

	if result.Allowed {
		t.Error("Blokajlı noktaya hareket izinli olmamalı ama izin verildi")
	} else {
		t.Log("Blokajlı noktaya hareket engellendi, test başarılı")
	}
}

// 24 Bear off Tek Taşı çift atıp 0 noktasından => 24 ile dışarı almak
func TestIsNormalMoveAllowed_DoubleDiceValidPath(t *testing.T) {
	stones := []*core.LogicalCoordinate{}
	player := 1

	// Oyuncunun taşı 0. noktada (başlangıç)
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex:   0,
		PositionType: core.PositionTypeEnum.Point,
		Player:       player,
		StackIndex:   0,
		IsTop:        true,
		MoveOrder:    0,
	})

	// Ara noktalar ve hedef boş, yani hareket serbest
	// Dice = [6,6,6,6] => Toplam 24 ilerleme -> 0 → 6 → 12 → 18 → 24 (bear off yapılmazsa 23 yeterli)
	toPoint := 24 // Bear off'a çıkmak gibi düşün. 24 Demek taşı toplamak demek..

	dice := []int{6, 6, 6, 6}

	result := core.IsNormalMoveAllowed(stones, player, 0, toPoint, dice)

	if !result.Allowed {
		t.Errorf("Double zarla 4 adımda hedefe ulaşmak mümkün olmalıydı. Sonuç: %+v", result)
	}
	t.Log("Tek Taş dışarı başarı ile alındı...")
}

func Test_Player1_BarAndMove_WithDoubleFour(t *testing.T) {
	core.ResetMoveOrder()

	var stones []*core.LogicalCoordinate

	// Player 1: 3 kırık taş (Bar'da)
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   -1,
			PositionType: core.PositionTypeEnum.Bar,
			Player:       1,
			StackIndex:   i,
			IsTop:        i == 2,
			MoveOrder:    0,
		})
	}

	// Player 2: 5 taş PointIndex 5'te
	for i := 0; i < 5; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   5,
			PositionType: core.PositionTypeEnum.Point,
			Player:       2,
			StackIndex:   i,
			IsTop:        i == 4,
			MoveOrder:    0,
		})
	}

	// Player 1: 2 taş PointIndex 0'da (bar girişi için engel değil)
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   0,
			PositionType: core.PositionTypeEnum.Point,
			Player:       1,
			StackIndex:   i,
			IsTop:        i == 1,
			MoveOrder:    0,
		})
	}

	// Player 2: 3 taş PointIndex 7'de
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   7,
			PositionType: core.PositionTypeEnum.Point,
			Player:       2,
			StackIndex:   i,
			IsTop:        i == 2,
			MoveOrder:    0,
		})
	}

	// Zar: Double 5 → [5, 5, 5, 5]
	dice := []int{5, 5}
	expandedDice := core.ExpandDice(dice)

	// 1. Bar girişine izin var mı?
	barResult := core.IsAllBarEntryAllowed(stones, 1, expandedDice)
	if !barResult.FromBar {
		t.Fatal("Bar'da kırık taş varken FromBar true olmalı")
	}
	if !barResult.Allowed {
		t.Fatal("Bar girişine izin verilmedi, oysa en az 1 taş girebilmeli")
	}
	if len(barResult.UsedDice) != 3 {
		t.Fatalf("Bar'dan girişte kullanılan zar sayısı 3 olmalı, mevcut: %d", len(barResult.UsedDice))
	}
	if len(barResult.RemainingDice) != 1 || barResult.RemainingDice[0] != 5 {
		t.Fatalf("Kalan zarlar [5] olmalı, mevcut: %v", barResult.RemainingDice)
	}

	// 2. Bar'dan mümkün olduğunca taşları Player 1 adina sokalım
	used := 0
	for _, die := range barResult.EnterableDice {
		if used >= 3 {
			break // sadece 3 kırık taş var
		}
		entryPoint := core.GetEntryPoint(1, die)

		var moved bool
		var broken []*core.LogicalCoordinate
		stones, moved, broken = core.MoveTopStoneAndUpdate(stones, 1, -1, entryPoint)
		if !moved {
			t.Fatalf("Bar'dan taş %d için hareket başarısız", used+1)
		}
		//Kirilan taslar yazdirilir..
		if moved && len(broken) > 0 {
			log.Printf("Player %d kırdı: %+v", 1, broken[0].PointIndex)
		}
		used++
	}

	if used != 3 {
		t.Fatalf("Bar'dan 3 taş yerine %d taş sokulabildi", used)
	}

	// 3. Kalan 1 hamleyle Player 1 icin 4 → 9 oynayalım
	moveResult := core.IsNormalMoveAllowed(stones, 1, 4, 9, barResult.RemainingDice)
	if !moveResult.Allowed {
		t.Fatal("4'ten 9'ye hareket izni verilmedi (Player 1)")
	}
	stones, moved, broken := core.MoveTopStoneAndUpdate(stones, 1, 4, 9)
	if !moved {
		t.Fatal("4'ten 9'ye taşıma başarısız")
	}

	//Kirilan taslar yazdirilir..
	if moved && len(broken) > 0 {
		log.Printf("Player %d kırdı: %+v", 1, broken[0].PointIndex)
	}

	// --- Kontroller ---

	// PointIndex 4'te 2 taş olmalı (bar giriş taşları)
	count := 0
	for _, s := range stones {
		if s.Player == 1 && s.PointIndex == 4 {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("PointIndex 4'te beklenen 2 taş yok, mevcut: %d", count)
	}

	// PointIndex 9'de Player 1'ye ait 1 taş olmalı
	player1Count := 0
	for _, s := range stones {
		if s.Player == 1 && s.PointIndex == 9 { // Player 1'in 1 tasi Point 9'a geldi..
			player1Count++
		}
	}
	if player1Count != 1 { // yeni tas geldi 1 oldu
		t.Fatalf("PointIndex 9'de Player 1'nin taş sayısı beklenenden farklı, mevcut: %d", player1Count)
	}
}

func TestPlayer2PossibleMoves(t *testing.T) {
	player1 := 1
	player2 := 2

	var stones []*core.LogicalCoordinate

	// Player 1 taşları
	for i := 0; i < 5; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 0,
			Player:     player1,
			IsTop:      i == 4, // sadece sonuncu top taş
		})
	}
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 11,
			Player:     player1,
			IsTop:      i == 1,
		})
	}
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 13,
			Player:     player1,
			IsTop:      i == 1,
		})
	}
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 15,
			Player:     player1,
			IsTop:      i == 1,
		})
	}

	// Player 2 taşları
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 4, Player: player2, IsTop: true})
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 5, Player: player2, IsTop: i == 1})
	}
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 6, Player: player2, IsTop: i == 2})
	}
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 7, Player: player2, IsTop: i == 1})
	}
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 12, Player: player2, IsTop: i == 1})
	}
	for i := 0; i < 5; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 23, Player: player2, IsTop: i == 4})
	}

	dice := []int{3, 2}

	// Player 2 için 23 ve 12 noktasından hareketler
	possibleFrom23 := core.GetPossibleMovePoints(stones, player2, 23, dice)
	possibleFrom12 := core.GetPossibleMovePoints(stones, player2, 12, dice)

	t.Logf("Player 2 taşları 23 noktasından gidebileceği noktalar: %v", possibleFrom23)
	t.Logf("Player 2 taşları 12 noktasından gidebileceği noktalar: %v", possibleFrom12)

	expected := []int{18, 20, 21}
	if !reflect.DeepEqual(possibleFrom23, expected) {
		t.Fatalf("PointIndex 18, 20 ve 21 olmasi gerekir!")
	}

	expected2 := []int{7, 9, 10}
	if !reflect.DeepEqual(possibleFrom12, expected2) {
		t.Fatalf("PointIndex 7, 9 ve 10 olmasi gerekir!")
	}
}

func TestPlayer2PossibleMovesFromBar(t *testing.T) {
	player1 := 1
	player2 := 2

	var stones []*core.LogicalCoordinate

	//Player 2 kirik tasi
	stones = append(stones, &core.LogicalCoordinate{PointIndex: -1, Player: player2, IsTop: true, PositionType: core.PositionTypeEnum.Bar})

	// Player 1 taşları
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 18, Player: player1, IsTop: true})
	//stones = append(stones, &core.LogicalCoordinate{PointIndex: 18, Player: player1, IsTop: true})
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 19, Player: player2, IsTop: i == 1})
	}
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 20, Player: player1, IsTop: i == 1})
	}
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 22, Player: player1, IsTop: i == 2})
	}
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 23, Player: player1, IsTop: i == 1})
	}

	dice := []int{6, 3}

	// Player 2 için -1 noktasından hareketler
	possibleFromBar := core.GetPossibleMovePoints(stones, player2, -1, dice)
	t.Logf("Player 2 taşları -1 noktasından gidebileceği noktalar: %v", possibleFromBar)

	expected := []int{18, 21}
	//expected := []int{21}
	if !reflect.DeepEqual(possibleFromBar, expected) {
		t.Fatalf("PointIndex 18 ve 21 olmasi gerekir!")
	}

	dice2 := []int{4, 1}

	// Player 2 için -1 noktasından hareketler
	possibleFromBar2 := core.GetPossibleMovePoints(stones, player2, -1, dice2)
	t.Logf("Player 2 taşları 12 noktasından gidebileceği noktalar: %v", possibleFromBar2)

	if len(possibleFromBar2) != 0 {
		t.Fatalf("PointIndex hic olmamasi gerekir!")
	}
}

func TestPlayer2PossibleMovesFor3Dice(t *testing.T) {
	player1 := 1
	player2 := 2

	var stones []*core.LogicalCoordinate

	//Player 1
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 0, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 1, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})

	// Player 2 taşları
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 4, Player: player2, IsTop: i == 1})
	}
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 5, Player: player2, IsTop: i == 1})
	}

	dice := []int{1, 1, 1}

	// Player 1 için 0 noktasından hareketler
	possibleFromBar := core.GetPossibleMovePoints(stones, player1, 0, dice)
	t.Logf("Player 1 taşları 0 noktasından gidebileceği noktalar: %v", possibleFromBar)

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(possibleFromBar, expected) {
		t.Fatalf("PointIndex 1,2 ve 3 olmasi gerekir!")
	}

	// Player 1 için 1 noktasından hareketler
	possibleFromBar1 := core.GetPossibleMovePoints(stones, player1, 1, dice)
	t.Logf("Player 1 taşları  noktasından gidebileceği noktalar: %v", possibleFromBar1)

	expected1 := []int{2, 3}
	if !reflect.DeepEqual(possibleFromBar1, expected1) {
		t.Fatalf("PointIndex 2 ve 3 olmasi gerekir!")
	}
}

func TestPlayer2PossibleMovesForDoubleDice(t *testing.T) {
	player2 := 2

	stones := core.GetInitialStones()

	dice := core.ExpandDice([]int{3, 3})

	// Player 2 için 23 noktasından hareketler
	possibleFromBar := core.GetPossibleMovePoints(stones, player2, 23, dice)
	t.Logf("Player 2 taşları 23 noktasından gidebileceği noktalar: %v", possibleFromBar)

	expected := []int{14, 17, 20}
	if !reflect.DeepEqual(possibleFromBar, expected) {
		t.Fatalf("PointIndex 14,17 ve 20 olmasi gerekir!")
	}

	dice1 := core.ExpandDice([]int{4, 4})
	// Player 1 için 1 noktasından hareketler
	possibleFromBar1 := core.GetPossibleMovePoints(stones, player2, 23, dice1)
	t.Logf("Player 2 taşları  noktasından gidebileceği noktalar: %v", possibleFromBar1)

	expected1 := []int{15, 19}
	if !reflect.DeepEqual(possibleFromBar1, expected1) {
		t.Fatalf("PointIndex 19 ve 15 olmasi gerekir!")
	}
}

func TestPlayer1ComplexPossibleMoves(t *testing.T) {
	player1 := 1
	player2 := 2
	var stones []*core.LogicalCoordinate

	// --- Player 1 Taşları (hareket etmesi beklenen oyuncu) ---
	// PointIndex 23: 2 taş
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 23,
			Player:     player1,
			IsTop:      i == 1,
		})
	}
	// PointIndex 16: 1 taş
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex: 16,
		Player:     player1,
		IsTop:      true,
	})
	// PointIndex 11: 3 taş
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 11,
			Player:     player1,
			IsTop:      i == 2,
		})
	}

	// --- Player 2 Taşları (engel teşkil edecek) ---
	// PointIndex 18: 2 taş (blok, girilemez)
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 18,
			Player:     player2,
			IsTop:      i == 1,
		})
	}
	// PointIndex 21: 1 taş (vurulabilir)
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex: 21,
		Player:     player2,
		IsTop:      true,
	})

	// PointIndex 13: 3 taş
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 13,
			Player:     player2,
			IsTop:      i == 2,
		})
	}

	// Zar: [5, 2]
	dice := []int{5, 2}

	// --- Test ---
	from23 := core.GetPossibleMovePoints(stones, player1, 23, dice)
	from18 := core.GetPossibleMovePoints(stones, player2, 18, dice)
	from11 := core.GetPossibleMovePoints(stones, player1, 11, dice)

	t.Logf("Player 1 - 23'ten oynanabilir noktalar: %v", from23)
	t.Logf("Player 2 - 18'den oynanabilir noktalar: %v", from18)
	t.Logf("Player 1 - 11'den oynanabilir noktalar: %v", from11)

	if from23 != nil {
		t.Fatalf("23'ten Player1 hicbir yere gidememeli ama bulundu: %v", from23)
	}

	expectedFrom18 := []int{13, 16}
	if !reflect.DeepEqual(from18, expectedFrom18) {
		t.Fatalf("18'den 13 ve 16'ya gidilebilmeli, ama su sonuc bulundu: %v", from18)
	}
	expectedFrom11 := []int{16}
	if !reflect.DeepEqual(from11, expectedFrom11) {
		t.Fatalf("11'den 16'ya gidilebilmeli ama, bulundu: %v", from18)
	}
}

func TestPlayer1ComplexPossibleMoves2(t *testing.T) {
	player1 := 1
	player2 := 2
	var stones []*core.LogicalCoordinate

	// --- Player 1 Taşları (hareket etmesi beklenen oyuncu) ---
	// PointIndex 23: 2 taş =>Player2
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 23,
			Player:     player2,
			IsTop:      i == 1,
		})
	}
	// PointIndex 20: 2 taş => =>Player1
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 20,
			Player:     player1,
			IsTop:      i == 1,
		})
	}

	// --- Player 2 Taşları (engel teşkil edecek) ---
	// PointIndex 18: 3 taş (blok, girilemez)
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 18,
			Player:     player2,
			IsTop:      i == 2,
		})
	}

	// PointIndex 16: 1 taş =>Player1
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex: 16,
		Player:     player1,
		IsTop:      true,
	})

	// PointIndex 15: 2 taş => Player2
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 15,
			Player:     player2,
			IsTop:      i == 1,
		})
	}

	// PointIndex 12: 2 taş => Player2
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 12,
			Player:     player2,
			IsTop:      i == 1,
		})
	}

	// PointIndex 11: 3 taş => Player1
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 11,
			Player:     player1,
			IsTop:      i == 2,
		})
	}

	// PointIndex 4: 3 taş => Player1
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 4,
			Player:     player1,
			IsTop:      i == 2,
		})
	}

	// Zar: [5, 2]
	dice := []int{5, 2}

	// --- Test ---
	from16 := core.GetPossibleMovePoints(stones, player1, 16, dice)
	from11 := core.GetPossibleMovePoints(stones, player1, 11, dice)
	from4 := core.GetPossibleMovePoints(stones, player1, 4, dice)

	t.Logf("Player 1 - 16'ten oynanabilir noktalar: %v", from16)
	t.Logf("Player 1 - 11'den oynanabilir noktalar: %v", from11)
	t.Logf("Player 1 - 4'den oynanabilir noktalar: %v", from4)

	expectedFrom16 := []int{21}
	if !reflect.DeepEqual(from16, expectedFrom16) {
		t.Fatalf("16'den 21'ya gidilebilmeli ama, bulundu: %v", from16)
	}

	expectedFrom11 := []int{13, 16}
	if !reflect.DeepEqual(from11, expectedFrom11) {
		t.Fatalf("18'den 16'ya gidilebilmeli, ama su sonuc bulundu: %v", from11)
	}
	expectedFrom4 := []int{6, 9, 11}
	if !reflect.DeepEqual(from4, expectedFrom4) {
		t.Fatalf("4'den 9 veya 11'e gidilebilmeli ama, bulundu: %v", from4)
	}
}

func TestPlayer3PossibleMovesForDoubleDiceFor2(t *testing.T) {
	player1 := 1
	player2 := 2

	var stones []*core.LogicalCoordinate

	//Player 2
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 20, Player: player2, IsTop: i == 1})
	}

	// Player 1 taşları
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 16, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 23, Player: player1, IsTop: i == 1})
	}
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 18, Player: player1, IsTop: i == 2})
	}

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 15, Player: player1, IsTop: i == 1})
	}

	stones = append(stones, &core.LogicalCoordinate{PointIndex: 13, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})
	/*for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 13, Player: player1, IsTop: i == 1})
	}*/

	dice := core.ExpandDice([]int{4, 3})

	possibleFromBar := core.GetPossibleMovePoints(stones, player2, 20, dice)
	t.Logf("Player 2 taşları  noktasından gidebileceği noktalar: %v", possibleFromBar)

	expected := []int{13, 16, 17}

	if !reflect.DeepEqual(possibleFromBar, expected) {
		t.Fatalf("PointIndex 13,16,17 olmasi gerekir!")
	}

	_, ok2, usedDice2, remainingDice2, broken := core.TryMoveStone(stones, player2, 20, possibleFromBar[0], dice)
	if !ok2 {
		t.Fatalf("Player 2 %v 'ye giebilmeli idi!", possibleFromBar[0])
	}

	t.Logf("Başarılı Player 2 %v noktasina gitti. Kullanılan zarlar: %v, Kalan zarlar: %v", possibleFromBar[0], usedDice2, remainingDice2)

	/*if len(broken) > 0 {
		log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player2, broken[0].PointIndex, broken[0].Player)
	}*/

	expectedBroken := []int{16, 13}
	var brokenReal []int

	for _, b := range broken {
		brokenReal = append(brokenReal, b.PointIndex)
		log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player2, b.PointIndex, b.Player)
	}

	if !reflect.DeepEqual(brokenReal, expectedBroken) {
		t.Fatalf("Player 1 icin PointIndex 13 ve 16 taslarinin kirilmasi gerekir!")
	}
}

func TestPlayer3PossibleMovesForDoubleBarDiceFor3(t *testing.T) {
	player1 := 1
	player2 := 2

	var stones []*core.LogicalCoordinate

	//Player 2
	stones = append(stones, &core.LogicalCoordinate{PointIndex: -1, Player: player2, IsTop: true, PositionType: core.PositionTypeEnum.Bar})
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 16, Player: player2, IsTop: true, PositionType: core.PositionTypeEnum.Point})

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 20, Player: player2, IsTop: i == 1})
	}

	// Player 1 taşları
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 23, Player: player1, IsTop: i == 1})
	}
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 18, Player: player1, IsTop: i == 2})
	}

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 15, Player: player1, IsTop: i == 1})
	}

	stones = append(stones, &core.LogicalCoordinate{PointIndex: 12, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 8, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})
	/*	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 12, Player: player1, IsTop: i == 1})
	}*/

	dice := core.ExpandDice([]int{4, 4})

	// Player 1 için 0 noktasından hareketler
	possibleFromBar := core.GetPossibleMovePoints(stones, player2, -1, dice)
	t.Logf("Player 2 taşları 0 noktasından gidebileceği noktalar: %v", possibleFromBar)

	expected := []int{20}
	if !reflect.DeepEqual(possibleFromBar, expected) {
		t.Fatalf("PointIndex 20 olmasi gerekir!")
	}

	// Player 2 için 4 noktasından hareketler
	possibleFromBar1 := core.GetPossibleMovePoints(stones, player2, 4, dice)
	t.Logf("Player 2 taşları  noktasından gidebileceği noktalar: %v", possibleFromBar1)

	if len(possibleFromBar1) != 0 {
		t.Fatalf("PointIndex hic olmamasi gerekir!")
	}

	//Kirik tasi girdik
	newStones, ok, usedDice, remainingDice, _ := core.TryMoveStone(stones, player2, -1, possibleFromBar[0], dice)
	if !ok {
		t.Fatalf("Kirik Player 2 icin girilebilmeli idi!")
	}
	t.Logf("Başarılı kirik zar girildi. Kullanılan zarlar: %v, Kalan zarlar: %v", usedDice, remainingDice)
	possibleFromBar2 := core.GetPossibleMovePoints(newStones, player2, possibleFromBar[0], remainingDice)
	t.Logf("Player 2 taşları  noktasından gidebileceği noktalar: %v", possibleFromBar2)

	expected2 := []int{8, 12, 16}
	if !reflect.DeepEqual(possibleFromBar2, expected2) {
		t.Fatalf("PointIndex 8,12,16 olmasi gerekir!")
	}

	_, ok2, usedDice2, remainingDice2, broken := core.TryMoveStone(newStones, player2, possibleFromBar[0], possibleFromBar2[0], remainingDice)
	if !ok2 {
		t.Fatalf("Player 2 %v 'ye giebilmeli idi!", possibleFromBar2[0])
	}

	t.Logf("Başarılı Player 2 %v noktasina gitti. Kullanılan zarlar: %v, Kalan zarlar: %v", possibleFromBar2[0], usedDice2, remainingDice2)

	/*if len(broken) > 0 {
		log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player2, broken[0].PointIndex, broken[0].Player)
	}*/

	expectedBroken := []int{12, 8}
	var brokenReal []int

	for _, b := range broken {
		brokenReal = append(brokenReal, b.PointIndex)
		log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player2, b.PointIndex, b.Player)
	}

	if !reflect.DeepEqual(brokenReal, expectedBroken) {
		t.Fatalf("Player 1 icin PointIndex 12,8 taslarinin kirilmasi gerekir!")
	}
}

func TestPlayer4PossibleMovesForDoubleDiceFor4(t *testing.T) {
	player1 := 1
	player2 := 2

	var stones []*core.LogicalCoordinate

	//Player 2
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 16, Player: player2, IsTop: true, PositionType: core.PositionTypeEnum.Point})

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 20, Player: player2, IsTop: i == 1})
	}

	// Player 1 taşları
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 18, Player: player1, IsTop: i == 2})
	}

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 15, Player: player1, IsTop: i == 1})
	}

	stones = append(stones, &core.LogicalCoordinate{PointIndex: 12, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 8, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 4, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})

	dice := core.ExpandDice([]int{4, 4})

	// Player 2 için 4 noktasından hareketler
	possibleFromBar := core.GetPossibleMovePoints(stones, player2, 20, dice)
	t.Logf("Player 2 taşları  noktasından gidebileceği noktalar: %v", possibleFromBar)

	expected := []int{4, 8, 12, 16}
	if !reflect.DeepEqual(possibleFromBar, expected) {
		t.Fatalf("PointIndex 4,8,12,16 olmasi gerekir!")
	}

	//4 Zar'da => 4zar ile hedefe ulasma
	//_, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player2, 20, possibleFromBar[0], dice)

	//4 Zar'da => 3zar ile hedefe ulasma
	_, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player2, 20, 8, dice)
	if !ok {
		t.Fatalf("Player 2 %v 'ye giebilmeli idi!", possibleFromBar[0])
	}

	//t.Logf("Başarılı Player 2 %v noktasina gitti. Kullanılan zarlar: %v, Kalan zarlar: %v", possibleFromBar[0], usedDice, remainingDice)
	t.Logf("Başarılı Player 2 %v noktasina gitti. Kullanılan zarlar: %v, Kalan zarlar: %v", 8, usedDice, remainingDice)

	/*if len(broken) > 0 {
		log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player2, broken[0].PointIndex, broken[0].Player)
	}*/

	//expectedBroken := []int{12, 8, 4}
	expectedBroken := []int{12, 8}
	var brokenReal []int

	for _, b := range broken {
		brokenReal = append(brokenReal, b.PointIndex)
		log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player2, b.PointIndex, b.Player)
	}

	if !reflect.DeepEqual(brokenReal, expectedBroken) {
		t.Fatalf("Player 1 icin PointIndex 12,8 taslarinin kirilmasi gerekir!")
	}
}

func TestPlayer1PossibleMovesForDoubleDiceFor4To1PointIndex(t *testing.T) {
	player1 := 1
	player2 := 2

	var stones []*core.LogicalCoordinate

	//Player 2

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 20, Player: player2, IsTop: i == 1})
	}

	// Player 1 taşları
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 16, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 18, Player: player1, IsTop: i == 2})
	}

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 15, Player: player1, IsTop: i == 1})
	}

	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{PointIndex: 12, Player: player1, IsTop: i == 1})
	}

	stones = append(stones, &core.LogicalCoordinate{PointIndex: 8, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})
	stones = append(stones, &core.LogicalCoordinate{PointIndex: 4, Player: player1, IsTop: true, PositionType: core.PositionTypeEnum.Point})

	dice := core.ExpandDice([]int{4, 4})

	// Player 2 için 4 noktasından hareketler
	possibleFromBar := core.GetPossibleMovePoints(stones, player2, 20, dice)
	t.Logf("Player 2 taşları  noktasından gidebileceği noktalar: %v", possibleFromBar)

	expected := []int{16}
	if !reflect.DeepEqual(possibleFromBar, expected) {
		t.Fatalf("PointIndex 16 olmasi gerekir!")
	}

	//4 Zar'da => 4zar ile hedefe ulasma
	//_, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player2, 20, possibleFromBar[0], dice)

	//4 Zar'da => 3zar ile hedefe ulasma
	_, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player2, 20, 8, dice)
	if ok {
		t.Fatalf("Player 2 %v 'ye gidememesi gerekiyordu!", possibleFromBar[0])
	}

	//t.Logf("Başarılı Player 2 %v noktasina gitti. Kullanılan zarlar: %v, Kalan zarlar: %v", possibleFromBar[0], usedDice, remainingDice)
	t.Logf("Player 2 %v noktasina gidememistir. Kullanılan zarlar: %v, Kalan zarlar: %v", 4, usedDice, remainingDice)

	/*if len(broken) > 0 {
		log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player2, broken[0].PointIndex, broken[0].Player)
	}*/

	if len(broken) > 0 {
		t.Fatalf("Player 1 icin hicbir tasin kirilmamasi gerekir!")
	}
}

func TestPlayer1BearOffMoves(t *testing.T) {
	player1 := 1

	var stones []*core.LogicalCoordinate

	// Player 1 taşları, 22 noktasında 1 taş var ve top o
	for i := 0; i < 1; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 22,
			Player:     player1,
			IsTop:      true,
		})
	}
	// Player 2 taşları, 18 noktasında 1 taş var ve top o
	for i := 0; i < 1; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 18,
			Player:     2,
			IsTop:      true,
		})
	}

	dice := []int{1, 5}

	possibleMoves := core.GetPossibleMovePoints(stones, player1, 22, dice)
	t.Logf("Player 1 taşları 22 noktasından gidebileceği noktalar: %v", possibleMoves)

	expected := []int{23, 24} // 23 normal, 24 bear off

	if !reflect.DeepEqual(possibleMoves, expected) {
		t.Fatalf("Beklenen hareketler %v iken, bulunan hareketler %v", expected, possibleMoves)
	}

	dice = core.ExpandDice([]int{6, 6})

	possibleMoves2 := core.GetPossibleMovePoints(stones, 2, 18, dice)
	t.Logf("Player 2 taşları 18 noktasından gidebileceği noktalar: %v", possibleMoves2)

	expected2 := []int{0, 6, 12} // 12, 6, 0 normal, [24 bear off] tas iceride olmadan olmuyor. Disardan dogrudan gelen tasin Bear Off olmasi icin "CanBearOffStone()" komple degismesi lazim

	if !reflect.DeepEqual(possibleMoves2, expected2) {
		t.Fatalf("Beklenen hareketler %v iken, bulunan hareketler %v", expected2, possibleMoves2)
	}
}

func TestPlayer2PossibleMoveForThreeDice(t *testing.T) {
	player2 := 2

	var stones []*core.LogicalCoordinate

	// Player 2 taşları, 12 noktasında 5 taş var ve top o
	for i := 0; i < 5; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 12,
			Player:     player2,
			IsTop:      i == 4,
		})
	}
	// Player 1 taşları, 4 noktasında 2 taş var ve top o
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 4,
			Player:     1,
			IsTop:      i == 1,
		})
	}

	dice := core.ExpandDice([]int{2, 2})

	possibleMoves := core.GetPossibleMovePoints(stones, player2, 12, dice)
	t.Logf("Player 2 taşları 12 noktasından gidebileceği noktalar: %v", possibleMoves)

	expected := []int{6, 8, 10}

	if !reflect.DeepEqual(possibleMoves, expected) {
		t.Fatalf("Beklenen hareketler %v iken, bulunan hareketler %v", expected, possibleMoves)
	}
}

func TestAreAllStonesInBearOffArea_Valid(t *testing.T) {
	player := 1
	var stones []*core.LogicalCoordinate

	// Player 1'in tüm taşları toplama alanında (18–23)
	for i := 18; i <= 23; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: i,
			Player:     player,
		})
	}

	// Toplanmış bir taş da olabilir
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex: 24,
		Player:     player,
	})

	ok := core.AreAllStonesInBearOffArea(stones, player)
	if !ok {
		t.Fatal("Beklenen: true, ancak false döndü")
	}
}

func TestAreAllStonesInBearOffArea_Invalid(t *testing.T) {
	player := 1
	var stones []*core.LogicalCoordinate

	// Toplama alanında taşlar
	for i := 18; i <= 22; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: i,
			Player:     player,
		})
	}

	//Hatali: Kirik Tasi var
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex: -1,
		Player:     player,
	})

	// Hatalı: 10. noktada bir taş
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex: 10,
		Player:     player,
	})

	ok := core.AreAllStonesInBearOffArea(stones, player)
	if ok {
		t.Fatal("Beklenen: false, ancak true döndü")
	}
}

func TestCanBearOffStone_InValid(t *testing.T) {
	player := 1
	dice := []int{3, 5}

	// Taşların toplama alanında olduğu bir durum (player 1 için 18-23 arası)
	stones := []*core.LogicalCoordinate{
		{Player: player, PointIndex: 18, IsTop: true},
		{Player: player, PointIndex: 19, IsTop: true},
		{Player: player, PointIndex: 20, IsTop: true},
		{Player: player, PointIndex: 21, IsTop: true},
		{Player: player, PointIndex: 22, IsTop: true},
		{Player: player, PointIndex: 23, IsTop: true},
	}

	// PointIndex 20'deki taş, zar 4 olmasa da 3 zarı ile çıkamaz, 5 ile çıkabilir
	// Mesafe: 23 - 20 + 1 = 4, zarlarda 3 ve 5 var, 3 eşit değil, 5 büyük

	pointIndex := 20

	canBearOff, _, _ := core.CanBearOffStone(stones, player, pointIndex, dice)

	if canBearOff {
		t.Fatalf("Taş toplanamamli ama toplanabiliyor. Geride tas var 5(19) ve 6(18)'da")
	}
}

// Tas Toplamayi test amacli yazilmis test. CanBearOffStone()
// TestCanBearOffStone_Valid tests the functionality of CanBearOffStone ensuring a player can bear off a stone with valid conditions.
func TestCanBearOffStone_Valid(t *testing.T) {
	player := 1
	dice := []int{3, 5}

	// Player 1 taşları sadece toplama alanında, daha geride taş yok
	stones := []*core.LogicalCoordinate{
		{Player: player, PointIndex: 20, IsTop: true}, // Toplanacak taş
		{Player: player, PointIndex: 21, IsTop: true},
		{Player: player, PointIndex: 22, IsTop: true},
		{Player: player, PointIndex: 23, IsTop: true},
	}

	pointIndex := 20

	canBearOff, remainingDice, usedDice := core.CanBearOffStone(stones, player, pointIndex, dice)

	if !canBearOff {
		t.Fatalf("Taş toplanabilmeli ama toplanamıyor")
	}

	t.Logf("Kullanilan zar: %v", usedDice)
	// Mesafe = 23 - 20 + 1 = 4, zar 3 yok, 5 var; 5 ile toplanabilir, kalan zar 3 olmalı
	expectedRemaining := []int{3}
	if !reflect.DeepEqual(remainingDice, expectedRemaining) {
		t.Fatalf("Kalan zarlar yanlış, beklenen: %v, bulunan: %v", expectedRemaining, remainingDice)
	}
}

// TestPlayer2CollectStones verifies that Player 2 can collect all stones according to the game rules without errors.
func TestPlayer2CollectStones(t *testing.T) {
	maxTries := 100
	tries := 0

	player := 2
	stones := []*core.LogicalCoordinate{}

	// Player 2 - PointIndex 5: 5 taş
	for i := 0; i < 5; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 5,
			Player:     player,
			IsTop:      i == 4,
		})
	}
	// Player 2 - PointIndex 2: 4 taş
	for i := 0; i < 4; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 2,
			Player:     player,
			IsTop:      i == 3,
		})
	}
	// Player 2 - PointIndex 3: 3 taş
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 3,
			Player:     player,
			IsTop:      i == 2,
		})
	}
	// Player 2 - PointIndex 1: 2 taş
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 1,
			Player:     player,
			IsTop:      i == 1,
		})
	}
	// Player 2 - PointIndex 2: 1 taş
	stones = append(stones, &core.LogicalCoordinate{
		PointIndex: 2,
		Player:     player,
		IsTop:      true,
	})
	// Player 1 - PointIndex 0: 2 taş
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex: 0,
			Player:     1,
			IsTop:      i == 1,
		})
	}

	//Tum taslar bitene kadar
	for !core.IsFinishedForPlayer(stones, player) {
		tries++
		if tries > maxTries {
			t.Fatalf("Test sonsuz döngüye girdi, taşlar toplanamadı")
		}
		dice, err := core.RollDice()
		if err != nil {
			t.Fatal(err)
		}
		slices.SortFunc(dice, func(a, b int) int {
			return b - a // Büyükten küçüğe sıralama
		})

		for len(dice) > 0 {
			moved := false
			for index := 5; index >= 0; index-- {
				result, remainingDice, usedDice := core.CanBearOffStone(stones, player, index, dice)
				if result {
					var broken []*core.LogicalCoordinate
					stones, result, broken = core.MoveTopStoneAndUpdate(stones, player, index, 24)

					if !result {
						t.Fatalf("MoveTopStoneAndUpdate başarısız oldu: index %d", index)
					}

					//Kirilan taslar yazdirilir..
					if result && len(broken) > 0 {
						log.Printf("Player %d kırdı: %+v", player, broken[0].PointIndex)
					}

					//Zarlar Player'a gore gercek degeri ile gosterilir.
					entryPoints := []int{}
					for _, d := range dice {
						entryPoints = append(entryPoints, d-1)
					}
					var tmpRemainingDice []int
					for i, _ := range remainingDice {
						tmpRemainingDice = append(tmpRemainingDice, remainingDice[i]-1)
					}
					if len(remainingDice) > 0 {
						t.Logf("Toplanan Tas %d. Zarlar %v Kullanilan Zar: %v Kalan Zar: %v", index, entryPoints, usedDice[0]-1, tmpRemainingDice)
					} else {
						t.Logf("Toplanan Tas %d. Zarlar %v Kullanilan Zar: %v", index, entryPoints, usedDice[0]-1)
					}
					dice = remainingDice
					moved = true
					break
				}
				if index == 0 && !moved && len(remainingDice) > 0 {
					var tmpRemainingDice []int
					for i, _ := range remainingDice {
						tmpRemainingDice = append(tmpRemainingDice, remainingDice[i]-1)
					}
					t.Logf("Kullanilamayan Zar: %v", tmpRemainingDice)
				}
			}
			if !moved {
				break // Bu zarlarla daha fazla hamle yapılamıyor
			}
		}
	}

	t.Log("Player 2 icin tum taslar toplandi")
	for _, stone := range stones {
		if stone.Player == player {
			if stone.PointIndex != 24 {
				t.Fatalf("Tum Taslar Toplanamadi")
			}
			t.Logf("Toplanan Tas %d", stone.PointIndex)
		}
	}
}

// *********Cok Onemli Test************
// TestPlayer2CollectStonesWithBiggerDice verifies that Player 2 can collect all stones using dice rolls with a larger span.
func TestPlayer2CollectStonesWithBiggerDice(t *testing.T) {
	maxTries := 100
	tries := 0

	player := 2
	stones := []*core.LogicalCoordinate{}

	// Player 2 - PointIndex 4: 4 taş
	/*for i := 0; i < 4; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   4,
			Player:       player,
			IsTop:        i == 3,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   i,
			MoveOrder:    0,
		})
	}*/

	// Player 2 - PointIndex 3: 5 taş
	for i := 0; i < 5; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   3,
			Player:       player,
			IsTop:        i == 4,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   i,
			MoveOrder:    0,
		})
	}
	// Player 2 - PointIndex 2: 3 taş
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   2,
			Player:       player,
			IsTop:        i == 2,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   i,
			MoveOrder:    0,
		})
	}
	// Player 2 - PointIndex 1: 3 taş
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   1,
			Player:       player,
			IsTop:        i == 2,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   i,
			MoveOrder:    0,
		})
	}
	// Player 1 - PointIndex 4: 3 taş
	for i := 0; i < 3; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   4,
			Player:       1,
			IsTop:        i == 2,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   i,
			MoveOrder:    0,
		})
	}
	// Player 1 - PointIndex 0: 2 taş
	for i := 0; i < 2; i++ {
		stones = append(stones, &core.LogicalCoordinate{
			PointIndex:   0,
			Player:       1,
			IsTop:        i == 1,
			PositionType: core.PositionTypeEnum.Point,
			StackIndex:   i,
			MoveOrder:    0,
		})
	}

	//Tum taslar bitene kadar
	for !core.IsFinishedForPlayer(stones, player) {
		tries++
		if tries > maxTries {
			t.Fatalf("Test sonsuz döngüye girdi, taşlar toplanamadı")
		}
		dice, err := core.RollDice()
		if err != nil {
			t.Fatal(err)
		}
		slices.SortFunc(dice, func(a, b int) int {
			return b - a // Büyükten küçüğe sıralama
		})

		for len(dice) > 0 {
			moved := false
			for index := 5; index >= 0; index-- {
				result, remainingDice, usedDice := core.CanBearOffStone(stones, player, index, dice)
				if result {
					var broken []*core.LogicalCoordinate
					stones, result, broken = core.MoveTopStoneAndUpdate(stones, player, index, 24)
					if !result {
						t.Fatalf("MoveTopStoneAndUpdate başarısız oldu: index %d", index)
					}

					//Kirilan taslar yazdirilir..
					if result && len(broken) > 0 {
						log.Printf("Player %d kırdı: %+v", player, broken[0].PointIndex)
					}

					//Zarlar Player'a gore gercek degeri ile gosterilir.
					entryPoints := []int{}
					for _, d := range dice {
						entryPoints = append(entryPoints, d-1)
					}
					var tmpRemainingDice []int
					for i, _ := range remainingDice {
						tmpRemainingDice = append(tmpRemainingDice, remainingDice[i]-1)
					}
					if len(remainingDice) > 0 {
						t.Logf("Toplanan Tas %d. Zarlar %v Kullanilan Zar: %v Kalan Zar: %v", index, entryPoints, usedDice[0]-1, tmpRemainingDice)
					} else {
						t.Logf("Toplanan Tas %d. Zarlar %v Kullanilan Zar: %v", index, entryPoints, usedDice[0]-1)
					}
					dice = remainingDice
					moved = true
					break
				}
				if index == 0 && !moved && len(remainingDice) > 0 {
					var tmpRemainingDice []int
					for i, _ := range remainingDice {
						tmpRemainingDice = append(tmpRemainingDice, remainingDice[i]-1)
					}
					t.Logf("Kullanilamayan Zar: %v", tmpRemainingDice)
				}
			}
			if !moved {
				break // Bu zarlarla daha fazla hamle yapılamıyor
			}
		}
	}

	t.Log("Player 2 icin tum taslar toplandi")
	for _, stone := range stones {
		if stone.Player == player {
			if stone.PointIndex != 24 {
				t.Fatalf("Tum Taslar Toplanamadi")
			}
			t.Logf("Toplanan Tas %d", stone.PointIndex)
		}
	}
}

func TestTryMoveStone_FiveTurnsAlternatingPlayers(t *testing.T) {
	stones := core.GetInitialStones()

	// Örnek zarlar (her tur için player ayrı ayrı)
	playerDice := map[int][][]int{
		1: {
			{3, 2},
			{5, 1}, //Kirik Tasta 5-1 => 4'e girmeye calisacak
			{6, 5}, //Kirik Tasi  6 -5 => 4'e girmeye calisacak 6 dolu zaten.
			{2, 3},
			{3, 3},
		},
		2: {
			{4, 2},
			{3, 2},
			{1, 6},
			{5, 3},
			{4, 1},
		},
	}

	// Başlangıç hamleleri (fromPoint -> toPoint) sabit, sadece örnek
	moves := map[int][][2]int{
		1: {
			{0, 3},   // player 1, tur 1: 0 -> 3
			{3, 4},   // tur 2: 3 -> 4 Kirik Tasi Var aslinda -1 => 4 [5,1] zarlar
			{0, 4},   // tur 3: 0 -> 4 Kirik Tasi Var aslinda -1 => 4 [6,5] zarlar. 6 ile giremez 5 ile 4'e kirik tasini sokar.
			{4, 6},   // tur 4: 10 -> 15 Player 2'nin tasini kirar.
			{16, 19}, // tur 5: 16 -> 19
		},
		2: {
			{7, 3},   // player 2, tur 1: 7 -> 3
			{3, 0},   // tur 2: 3 -> 0 (bear off için)
			{12, 6},  // tur 3: 12 -> 6
			{23, 19}, // tur 4: 23 -> 19 Kirik tasi var -1 => 19 [5,3] zarlar ile 5 ile 19'a kirik tasini sokar.
			{-1, 23}, // tur 5: 4 -> 0
		},
	}

	for turn := 0; turn < 5; turn++ {
		for player := 1; player <= 2; player++ {
			fromPoint := moves[player][turn][0]
			toPoint := moves[player][turn][1]
			dice := core.ExpandDice(playerDice[player][turn])

			//Demek ki kirik tasi var...Otomatik -1'e ata...
			if core.PlayerMustEnterFromBar(stones, player) {
				fromPoint = -1
			}
			//--------------------------
			t.Logf("Turn %d, Player %d, Move: %d -> %d, Dice: %v", turn+1, player, fromPoint, toPoint, dice)

			newStones, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player, fromPoint, toPoint, dice)
			if !ok {
				t.Errorf("Player %d hamlesi başarısız oldu: %d -> %d", player, fromPoint, toPoint)
				continue
			}
			//Kirilan taslar yazdirilir..
			if ok && len(broken) > 0 {
				log.Printf("Player %d kırdı: %+v", player, broken[0].PointIndex)
			}

			stones = newStones

			t.Logf("Başarılı hareket. Kullanılan zarlar: %v, Kalan zarlar: %v", usedDice, remainingDice)
			//core.SortStonesByPlayerPointAndStackDesc(stones)

			/*t.Log("Taşların güncel durumu:")
			for _, stone := range stones {
				t.Logf("PointIndex: %2d, Player: %d, StackIndex: %d, IsTop: %v, MoveOrder: %d",
					stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
			}*/
		}
	}
}

// Butun Tavla Dinamikleri Test Edilir. Sadece bot yok.
func TestFullSmilation(t *testing.T) {
	//Butun taslar dizildi
	stones := core.GetInitialStones()

	// Atilan zarlar (her tur için player ayrı ayrı)
	playerDice := map[int][][]int{
		1: {
			{6, 5},
			{4, 2},
			{6, 4},
			{6, 2},
			{5, 1},
			{5, 4},
			{2, 1},
			{5, 2},
			{2, 4},
			{3, 2},
			{3, 2},
			{3, 2},
			{6, 1},
			{6, 2},
			{6, 3},
			{6, 2},
			{5, 5},
			{6, 5},
			{2, 2},
			{6, 5},
			{4, 3},
			{4, 3},
			{1, 1},
			{3, 1},
			{2, 2},
			{6, 3},
			{6, 6},
			{3, 1},
			{6, 4},
			{6, 5},
			{1, 6},
		},
		2: {
			{4, 2},
			{6, 1},
			{5, 3},
			{4, 2},
			{6, 4},
			{5, 3},
			{4, 2},
			{5, 2},
			{1, 4},
			{6, 4},
			{2, 1},
			{5, 1},
			{5, 4},
			{4, 1},
			{5, 4},
			{5, 3},
			{4, 4},
			{5, 2},
			{2, 2},
			{4, 4},
			{5, 1},
			{3, 2},
			{6, 5},
			{5, 2},
			{4, 1},
			{6, 3},
			{5, 4},
			{5, 3},
			{4, 3},
			{3, 3},
			{5, 5},
		},
	}

	// Oynanan Taslar (fromPoint -> toPoint)
	moves := map[int][][][2]int{
		//Player 1
		1: {
			{{0, 11}},
			{{16, 20}, {18, 20}},
			{{16, 22}, {18, 22}},
			{{16, 22}, {11, 13}},
			{{13, 18}, {0, 1}},
			{{-1, 4}, {4, 8}}, // Kirik Tasini  girdi..
			{{11, 14}},
			{{-1, 4}, {18, 20}},
			{{4, 10}},
			{{11, 14}, {10, 12}},
			{{11, 14}, {12, 14}},
			{{14, 16}, {20, 23}},                   //Player 2 in => PointIndex 16 ve 23 kirildi...
			{},                                     // Player 1'e gele geldi
			{},                                     // Player 1'e 2. kere gele geldi
			{},                                     // Player 1'e 3. kere gele geldi
			{},                                     // Player 1'e 4. kere gele geldi
			{{-1, 4}, {-1, 4}, {11, 16}, {11, 16}}, //Player 1 2 tasi da girer ve Player 2 PointIndex 16 tasi kirar..
			{{4, 9}, {4, 10}},
			{{9, 17}},
			{{17, 22}, {14, 20}},
			{{10, 17}},
			{{17, 21}, {18, 21}},
			{{16, 17}, {16, 17}, {22, 23}, {22, 23}}, //Player 2 PointIndex 17 tasi kirar..
			{{14, 18}},
			{{17, 19}, {17, 19}, {14, 18}},
			{{18, 24}, {21, 24}},
			{{18, 24}, {18, 24}, {18, 24}, {19, 24}},
			{{21, 24}, {23, 24}},
			{{19, 24}, {20, 24}},
			{{20, 24}, {20, 24}},
			{{22, 24}, {23, 24}},
		},
		//Player 2
		2: {
			{{5, 3}, {7, 3}},
			{{12, 6}, {7, 6}}, //
			{{7, 2}, {5, 2}},  //
			{{12, 6}},         //
			{{12, 6}, {5, 1}}, // Player 1 in => PointIndex 1 kirildi...
			{{6, 1}, {6, 3}},
			{{12, 10}, {12, 8}},
			{{8, 6}, {10, 5}},
			{{5, 0}},
			{{6, 0}, {23, 19}},
			{{19, 16}},
			{{-1, 23}, {-1, 19}}, //Player 1 in => PointIndex 23 kirildi
			{{23, 19}, {6, 1}},
			{{6, 2}, {1, 0}},
			{{19, 10}},
			{{19, 16}, {10, 5}},
			{}, //Player 2 gele atti..
			{{-1, 19}, {19, 17}},
			{}, // Player 2 gele atti..
			{}, // Player 2 2. kere gele atti..
			{{-1, 19}, {2, 1}},
			{{19, 17}, {3, 0}},
			{{-1, 19}, {19, 13}},
			{{13, 6}},
			{{6, 2}, {0, 24}},
			{{5, 24}, {2, 24}},
			{{3, 24}, {5, 0}},
			{{2, 24}, {5, 0}},
			{{2, 24}, {3, 24}},
			{{1, 24}, {1, 24}, {1, 24}, {0, 24}},
			{{0, 24}, {0, 24}, {0, 24}, {0, 24}},
		},
	}

	for turn := 0; turn < len(playerDice[1]); turn++ {
		for player := 1; player <= 2; player++ {
			turnMoves := moves[player][turn]
			dice := core.ExpandDice(playerDice[player][turn])

			t.Logf("===== Turn %d | Player %d | Zarlar: %v =====", turn+1, player, dice)

			for moveIndex, move := range turnMoves {
				if len(dice) == 0 {
					t.Logf("Player %d için zar kalmadı, hamle durduruluyor", player)
					break
				}

				fromPoint := move[0]
				toPoint := move[1]

				// Kirik taş varsa, otomatik bar'dan gir
				if core.PlayerMustEnterFromBar(stones, player) {
					fromPoint = -1
				}

				t.Logf("Move %d.%d: %d -> %d, Dice: %v", turn+1, moveIndex+1, fromPoint, toPoint, dice)

				newStones, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player, fromPoint, toPoint, dice)
				if !ok {
					t.Errorf("Player %d hamlesi başarısız oldu: %d -> %d", player, fromPoint, toPoint)
					break // başarısızsa durdurabiliriz, ya da continue
				}

				if len(broken) > 0 {
					log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player, broken[0].PointIndex, broken[0].Player)
				}

				stones = newStones
				dice = remainingDice

				t.Logf("Başarılı hareket. Kullanılan zarlar: %v, Kalan zarlar: %v", usedDice, remainingDice)

				/*core.SortStonesByPlayerPointAndStackDesc(stones)

				t.Log("Taşların güncel durumu:")
				for _, stone := range stones {
					t.Logf("PointIndex: %2d, Player: %d, StackIndex: %d, IsTop: %v, MoveOrder: %d",
						stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
				}*/
			}

			if len(turnMoves) == 0 && core.PlayerMustEnterFromBar(stones, player) {
				log.Printf("Player %d kirik tasi var ve gele geldi!", player)
			}
		}
	}
	//Notify
	log.Printf("Player %d'nin Topladigi Toplam Tas : %d", 2, core.CountCollectedStones(stones, 2))
	log.Printf("Player %d'nin Topladigi Toplam Tas : %d", 1, core.CountCollectedStones(stones, 1))

	if core.IsFinishedForPlayer(stones, 1) {
		log.Printf("Player %d Kazandi!", 1)
	} else if core.IsFinishedForPlayer(stones, 2) {
		log.Printf("Player %d Kazandi!", 2)
	}
}

// Butun Tavla Dinamikleri 2. Senaryo ile Test Edilir. Sadece bot yok.
func TestFullSmilation2(t *testing.T) {
	//Butun taslar dizildi
	stones := core.GetInitialStones()

	// Atilan zarlar (her tur için player ayrı ayrı)
	playerDice := map[int][][]int{
		// 24
		1: {
			{5, 3},
			{5, 4},
			{4, 4},
			{3, 5},
			{5, 3},
			{3, 6},
			{3, 5},
			{4, 6},
			{3, 2},
			{4, 1},
			{3, 5},
			{1, 3},
			{6, 6},
			{2, 1},
			{5, 6},
			{2, 3},
			{1, 1},
			{3, 1},
			{3, 2},
			{5, 4},
			{1, 6},
			{6, 2},
			{1, 6},
			{3, 6},
			{2, 1},
			{6, 6},
			{5, 1},
			{5, 6},
			{5, 2},
			{4, 2},
			{2, 2},
			{5, 2},
		},
		2: {
			{2, 5},
			{3, 2},
			{1, 2},
			{6, 2},
			{1, 5},
			{1, 2},
			{3, 5},
			{4, 2},
			{4, 6},
			{4, 1},
			{6, 4},
			{4, 3},
			{4, 1},
			{6, 3},
			{4, 3},
			{3, 3},
			{4, 2},
			{4, 3},
			{5, 4},
			{4, 4},
			{5, 5},
			{6, 2},
			{4, 5},
			{6, 3},
			{4, 1},
			{2, 2},
			{1, 1},
			{5, 1},
			{1, 2},
			{6, 6},
			{6, 2},
			{6, 2},
		},
	}

	// Oynanan Taslar (fromPoint -> toPoint)
	moves := map[int][][][2]int{
		//Player 1
		1: {
			{{18, 21}, {16, 21}},
			{{11, 16}, {11, 15}},
			{{11, 15}, {15, 19}, {15, 19}, {18, 22}},
			{{-1, 4}, {16, 19}},
			{{11, 16}, {11, 14}},
			{{14, 17}, {4, 10}},
			{{-1, 4}, {17, 20}},
			{{-1, 3}, {3, 9}},
			{{4, 9}},
			{{16, 20}, {19, 20}},
			{{18, 23}, {16, 19}},
			{{0, 1}, {16, 19}},
			{{9, 21}, {9, 21}}, //......
			{{-1, 0}, {-1, 1}}, //Player 2 PointIndex 1 kirilir.
			{{18, 23}},         //Player 2 PointIndex 23 kirildi.
			{{21, 23}, {18, 21}},
			{{0, 1}, {0, 1}, {21, 22}, {21, 22}},
			{{19, 23}},
			{{1, 4}, {19, 21}},
			{{4, 13}},
			{{1, 8}},
			{{8, 16}},
			{{1, 8}},
			{{13, 16}, {8, 14}},
			{{14, 16}, {21, 22}},
			{{16, 22}, {16, 22}, {16, 22}, {19, 24}}, // Player 1Toplamaya  basladi **************
			{{19, 24}, {23, 24}},
			{{20, 24}, {20, 24}},
			{{21, 24}, {22, 24}},
			{{21, 24}, {22, 24}},
			{{22, 24}, {22, 24}, {22, 24}, {22, 24}},
			{{23, 24}, {23, 24}},
		},
		//Player 2
		2: {
			{{12, 5}},
			{{12, 7}},
			{{23, 22}, {22, 20}}, // Player 1 PointIndex 22 kirildi
			{{20, 12}},
			{{23, 17}},
			{{-1, 23}, {12, 10}},
			{{23, 20}, {10, 5}},
			{{7, 3}, {5, 3}},
			{{20, 10}},
			{{10, 6}, {7, 6}},
			{{12, 6}, {5, 1}},
			{},                 //Gele atti...
			{{-1, 23}, {5, 1}}, // Player 1 PointIndex 23 kirilir, Player 1 PointIndex 1 kirilir.
			{},
			{},
			{},
			{},
			{},
			{},
			{},
			{},
			{{-1, 18}},
			{},
			{{-1, 18}, {7, 4}},
			{{12, 7}},
			{{6, 4}, {6, 4}, {7, 5}, {7, 5}},
			{{6, 5}, {12, 9}},
			{{18, 12}},
			{{12, 9}},
			{{18, 6}, {9, 3}, {9, 3}},
			{{6, 4}, {5, 24}},
			{{5, 24}, {4, 2}},
		},
	}

	for turn := 0; turn < len(playerDice[1]); turn++ {
		for player := 1; player <= 2; player++ {
			turnMoves := moves[player][turn]
			dice := core.ExpandDice(playerDice[player][turn])

			t.Logf("===== Turn %d | Player %d | Zarlar: %v =====", turn+1, player, dice)

			for moveIndex, move := range turnMoves {
				if len(dice) == 0 {
					t.Logf("Player %d için zar kalmadı, hamle durduruluyor", player)
					break
				}

				fromPoint := move[0]
				toPoint := move[1]

				// Kirik taş varsa, otomatik bar'dan gir
				if core.PlayerMustEnterFromBar(stones, player) {
					fromPoint = -1
				}

				t.Logf("Move %d.%d: %d -> %d, Dice: %v", turn+1, moveIndex+1, fromPoint, toPoint, dice)

				newStones, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player, fromPoint, toPoint, dice)
				if !ok {
					t.Errorf("Player %d hamlesi başarısız oldu: %d -> %d", player, fromPoint, toPoint)
					break // başarısızsa durdurabiliriz, ya da continue
				}

				if len(broken) > 0 {
					log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player, broken[0].PointIndex, broken[0].Player)
				}

				stones = newStones
				dice = remainingDice

				t.Logf("Başarılı hareket. Kullanılan zarlar: %v, Kalan zarlar: %v", usedDice, remainingDice)

				/*core.SortStonesByPlayerPointAndStackDesc(stones)

				t.Log("Taşların güncel durumu:")
				for _, stone := range stones {
					t.Logf("PointIndex: %2d, Player: %d, StackIndex: %d, IsTop: %v, MoveOrder: %d",
						stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
				}*/
			}

			if len(turnMoves) == 0 && core.PlayerMustEnterFromBar(stones, player) {
				log.Printf("Player %d kirik tasi var ve gele geldi!", player)
			}
		}
	}
	//Notify
	log.Printf("Player %d'nin Topladigi Toplam Tas : %d", 2, core.CountCollectedStones(stones, 2))
	log.Printf("Player %d'nin Topladigi Toplam Tas : %d", 1, core.CountCollectedStones(stones, 1))

	if core.IsFinishedForPlayer(stones, 1) {
		log.Printf("Player %d Kazandi!", 1)
	} else if core.IsFinishedForPlayer(stones, 2) {
		log.Printf("Player %d Kazandi!", 2)
	}
}

// Butun Tavla Dinamikleri 3. Senaryo ile Test Edilir. Sadece bot yok.
func TestFullSmilation3(t *testing.T) {
	//Butun taslar dizildi
	stones := core.GetInitialStones()

	// 14
	// Atilan zarlar (her tur için player ayrı ayrı)
	playerDice := map[int][][]int{
		//
		1: {
			{3, 6},
			{4, 4},
			{4, 3},
			{6, 6},
			{2, 2},
			{1, 3},
			{1, 2},
			{4, 3},
			{3, 2},
			{3, 3},
			{6, 3},
			{6, 1},
			{6, 3},
			{4, 1},
			{2, 6},
			{2, 4},
			{2, 4},
			{6, 4},
			{5, 5},
		},
		2: {
			{3, 5},
			{3, 3},
			{4, 6},
			{3, 3},
			{4, 3},
			{3, 4},
			{1, 2},
			{1, 5},
			{5, 6},
			{3, 1},
			{2, 1},
			{4, 6},
			{5, 5},
			{1, 1},
			{5, 2},
			{1, 5},
			{3, 4},
			{6, 1},
			{3, 1},
		},
	}

	// Oynanan Taslar (fromPoint -> toPoint)
	moves := map[int][][][2]int{
		//Player 1
		1: {
			{{0, 6}, {0, 3}},
			{{6, 14}, {11, 15}, {11, 15}},
			{{11, 14}, {16, 20}},
			{{3, 21}, {16, 22}},
			{{16, 20}, {11, 13}, {11, 13}},
			{{18, 21}, {18, 19}},
			{{19, 22}},
			{{15, 18}, {15, 19}},
			{{18, 23}},
			{{13, 19}, {13, 19}},
			{{14, 20}, {14, 17}},
			{{17, 18}, {18, 24}},
			{{18, 24}, {21, 24}},
			{{20, 24}, {23, 24}},
			{{22, 24}, {18, 24}},
			{{20, 24}, {22, 24}},
			{{20, 24}, {18, 20}},
			{{19, 24}, {20, 24}},
			{{19, 24}, {19, 24}, {21, 24}},
		},
		//Player 2
		2: {
			{{5, 2}, {7, 2}},
			{{23, 17}, {23, 17}},
			{{12, 2}},
			{{12, 6}, {12, 6}},
			{{7, 3}, {6, 3}},
			{{12, 9}, {12, 8}},
			{{6, 5}, {7, 5}},
			{{9, 4}, {5, 4}},
			{{17, 12}, {17, 11}},
			{{8, 5}, {12, 11}},
			{{11, 8}},
			{{11, 5}, {8, 4}},
			{{4, 24}, {4, 24}, {5, 0}, {5, 0}},
			{{5, 4}, {5, 4}, {5, 4}, {2, 1}},
			{{4, 24}, {1, 24}},
			{{4, 24}, {5, 4}},
			{{2, 24}, {3, 24}},
			{{5, 24}, {4, 3}},
			{},
		},
	}

	for turn := 0; turn < len(playerDice[1]); turn++ {
		for player := 1; player <= 2; player++ {
			turnMoves := moves[player][turn]
			dice := core.ExpandDice(playerDice[player][turn])

			t.Logf("===== Turn %d | Player %d | Zarlar: %v =====", turn+1, player, dice)

			for moveIndex, move := range turnMoves {
				if len(dice) == 0 {
					t.Logf("Player %d için zar kalmadı, hamle durduruluyor", player)
					break
				}

				fromPoint := move[0]
				toPoint := move[1]

				// Kirik taş varsa, otomatik bar'dan gir
				if core.PlayerMustEnterFromBar(stones, player) {
					fromPoint = -1
				}

				t.Logf("Move %d.%d: %d -> %d, Dice: %v", turn+1, moveIndex+1, fromPoint, toPoint, dice)

				newStones, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player, fromPoint, toPoint, dice)
				if !ok {
					t.Errorf("Player %d hamlesi başarısız oldu: %d -> %d", player, fromPoint, toPoint)
					break // başarısızsa durdurabiliriz, ya da continue
				}

				if len(broken) > 0 {
					log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player, broken[0].PointIndex, broken[0].Player)
				}

				stones = newStones
				dice = remainingDice

				t.Logf("Başarılı hareket. Kullanılan zarlar: %v, Kalan zarlar: %v", usedDice, remainingDice)

				/*core.SortStonesByPlayerPointAndStackDesc(stones)

				t.Log("Taşların güncel durumu:")
				for _, stone := range stones {
					t.Logf("PointIndex: %2d, Player: %d, StackIndex: %d, IsTop: %v, MoveOrder: %d",
						stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
				}*/
			}

			if len(turnMoves) == 0 && core.PlayerMustEnterFromBar(stones, player) {
				log.Printf("Player %d kirik tasi var ve gele geldi!", player)
			}
		}
	}
	//Notify
	log.Printf("Player %d'nin Topladigi Toplam Tas : %d", 2, core.CountCollectedStones(stones, 2))
	log.Printf("Player %d'nin Topladigi Toplam Tas : %d", 1, core.CountCollectedStones(stones, 1))

	if core.IsFinishedForPlayer(stones, 1) {
		log.Printf("Player %d Kazandi!", 1)
	} else if core.IsFinishedForPlayer(stones, 2) {
		log.Printf("Player %d Kazandi!", 2)
	}
}

// Butun Tavla Dinamikleri 4. Senaryo ile Test Edilir. Sadece bot yok.
func TestFullSmilation4(t *testing.T) {
	//Butun taslar dizildi
	stones := core.GetInitialStones()

	// 9
	// Atilan zarlar (her tur için player ayrı ayrı)
	playerDice := map[int][][]int{
		//
		1: {
			{3, 3},
			{2, 6},
			{6, 2},
			{1, 3},
			{6, 1},
			{1, 4},
			{3, 2},
			{6, 1},
			{2, 1},
			{2, 5},
			{5, 4},
			{5, 5},
			{1, 5},
			{2, 2},
			{2, 3},
			{5, 6},
			{1, 1}, //--------------------
			{2, 5},
			{4, 1},
			{3, 2},
			{6, 5},
			{4, 2},
		},
		2: {
			{1, 4},
			{6, 5},
			{6, 3},
			{5, 4},
			{2, 1},
			{6, 1},
			{2, 4},
			{4, 3},
			{6, 5},
			{6, 1},
			{3, 4},
			{1, 5},
			{5, 5},
			{5, 6},
			{4, 1},
			{2, 6},
			{1, 6},
			{2, 5},
			{5, 4},
			{4, 6},
			{4, 3},
			{3, 3},
		},
	}

	// Oynanan Taslar (fromPoint -> toPoint)
	moves := map[int][][][2]int{
		//Player 1
		1: {
			{{0, 6}, {0, 6}},
			{{11, 19}},
			{{11, 19}},
			{{11, 15}}, //Player 2 Pointindex 14 kirildi.
			{{-1, 0}, {0, 6}},
			{{16, 20}, {18, 19}},
			{{18, 20}, {18, 21}},
			{{16, 22}, {21, 22}}, //{{6, 13}},
			{{16, 18}, {19, 20}},
			{{6, 11}, {20, 22}},
			{{11, 20}},
			{{6, 16}, {6, 16}},
			{{18, 23}, {22, 23}},
			{{16, 18}, {16, 18}, {11, 13}, {11, 13}},
			{{13, 18}},
			{{13, 18}, {18, 24}},
			{{23, 24}, {23, 24}, {22, 23}, {23, 24}}, //{{23, 24}, {23, 24}, {22, 24}}, Once oynayacak sonra toplayacak..Dogrudan olursa isler karisiyor..
			{{19, 24}, {22, 24}},
			{{20, 24}, {20, 21}},
			{{21, 24}, {18, 20}},
			{{18, 24}, {19, 24}},
			{{20, 24}, {18, 20}},
		},
		//Player 2
		2: {
			{{12, 7}},
			{{23, 12}},
			{{23, 14}},
			{{-1, 20}, {20, 15}}, //Player 1 PointIndex 15 krildi
			{{15, 12}},
			{{7, 0}},
			{{5, 3}, {7, 3}},
			{{12, 5}},
			{{12, 1}},
			{{7, 0}},
			{{7, 3}, {5, 2}},
			{{12, 6}},
			{{12, 7}, {12, 7}, {12, 7}, {7, 2}},
			{{7, 1}, {7, 2}},
			{{6, 2}, {0, 24}},
			{{5, 24}, {1, 24}},
			{{5, 24}, {0, 24}},
			{{1, 24}, {5, 0}},
			{{3, 24}, {5, 0}},
			{{3, 24}, {3, 24}},
			{{2, 24}, {2, 24}},
			{{2, 24}, {2, 24}, {0, 24}, {0, 24}},
		},
	}

	for turn := 0; turn < len(playerDice[1]); turn++ {
		for player := 1; player <= 2; player++ {
			turnMoves := moves[player][turn]
			dice := core.ExpandDice(playerDice[player][turn])

			t.Logf("===== Turn %d | Player %d | Zarlar: %v =====", turn+1, player, dice)

			for moveIndex, move := range turnMoves {
				if len(dice) == 0 {
					t.Logf("Player %d için zar kalmadı, hamle durduruluyor", player)
					break
				}

				fromPoint := move[0]
				toPoint := move[1]

				// Kirik taş varsa, otomatik bar'dan gir
				if core.PlayerMustEnterFromBar(stones, player) {
					fromPoint = -1
				}

				t.Logf("Move %d.%d: %d -> %d, Dice: %v", turn+1, moveIndex+1, fromPoint, toPoint, dice)

				newStones, ok, usedDice, remainingDice, broken := core.TryMoveStone(stones, player, fromPoint, toPoint, dice)
				if !ok {
					t.Errorf("Player %d hamlesi başarısız oldu: %d -> %d", player, fromPoint, toPoint)
					break // başarısızsa durdurabiliriz, ya da continue
				}

				if len(broken) > 0 {
					log.Printf("Player %d kırdı: PointIndex=%d, Player=%d", player, broken[0].PointIndex, broken[0].Player)
				}

				stones = newStones
				dice = remainingDice

				t.Logf("Başarılı hareket. Kullanılan zarlar: %v, Kalan zarlar: %v", usedDice, remainingDice)

				/*core.SortStonesByPlayerPointAndStackDesc(stones)

				t.Log("Taşların güncel durumu:")
				for _, stone := range stones {
					t.Logf("PointIndex: %2d, Player: %d, StackIndex: %d, IsTop: %v, MoveOrder: %d",
						stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
				}*/
			}

			if len(turnMoves) == 0 && core.PlayerMustEnterFromBar(stones, player) {
				log.Printf("Player %d kirik tasi var ve gele geldi!", player)
			}
		}
	}
	//Notify
	log.Printf("Player %d'nin Topladigi Toplam Tas : %d", 2, core.CountCollectedStones(stones, 2))
	log.Printf("Player %d'nin Topladigi Toplam Tas : %d", 1, core.CountCollectedStones(stones, 1))

	if core.IsFinishedForPlayer(stones, 1) {
		log.Printf("Player %d Kazandi!", 1)
	} else if core.IsFinishedForPlayer(stones, 2) {
		log.Printf("Player %d Kazandi!", 2)
	}
}
