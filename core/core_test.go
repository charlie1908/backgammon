package core_test

import (
	"backgammon/core"
	"fmt"
	"reflect"
	"testing"
)

func TestMoveTopStoneAndUpdate_ValidMove_WithConsoleOutput(t *testing.T) {
	stones := core.GetInitialStones()

	player := 1
	fromPoint := 0      // Player 1 hareket etmek istediği yer
	toPoint := 3        // Hedef nokta (Geceli acik)
	dice := []int{1, 2} // Zarlar

	//Once kirik tasi var mi kontrolu
	result := core.IsBarEntryAllowed(stones, player, dice)

	if result.FromBar {
		t.Error("Bar'da taş var gozukuyor ama yok, FromBar false olmalı")
	}
	//Kirik tasi yok. Normal hareket kontrolu
	result = core.IsNormalMoveAllowed(stones, player, fromPoint, toPoint, dice)
	if !result.Allowed {
		t.Fatalf("Normal hareket izin verilmedi ama verilmesi bekleniyordu: %d -> %d", fromPoint, toPoint)
	}

	updatedStones, moved := core.MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
	if !moved {
		t.Errorf("Taş hareket etmedi ama hareket mümkün olmalıydı.")
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
	result := core.IsBarEntryAllowed(stones, player, dice)

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
	updatedStones, moved := core.MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
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
	result := core.IsBarEntryAllowed(stones, player, dice)
	if result.FromBar {
		t.Error("Bar'da taş olduğu görünüyor ama olmamalı, FromBar false olmalı")
	}

	// Normal hamle izni kontrolü
	result = core.IsNormalMoveAllowed(stones, player, fromPoint, toPoint, dice)
	if !result.Allowed {
		t.Fatalf("Normal hareket izni verilmedi ama verilmeliydi: %d -> %d", fromPoint, toPoint)
	}

	// Hareketi uygula
	updatedStones, moved := core.MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
	if !moved {
		t.Fatalf("Taş hareket etmedi ama mümkün olmalıydı: %d -> %d", fromPoint, toPoint)
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
	barResult := core.IsBarEntryAllowed(stones, player, dice)
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

	result := core.IsBarEntryAllowed(stones, player, dice)

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

	result := core.IsBarEntryAllowed(stones, player, dice)

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
	dice := []int{1, 1}

	result := core.IsBarEntryAllowed(stones, player, dice)

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
	result := core.IsBarEntryAllowed(stones, player2, dice)

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
	barResult := core.IsBarEntryAllowed(stones, 1, expandedDice)
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
		stones, moved = core.MoveTopStoneAndUpdate(stones, 1, -1, entryPoint)
		if !moved {
			t.Fatalf("Bar'dan taş %d için hareket başarısız", used+1)
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
	stones, moved := core.MoveTopStoneAndUpdate(stones, 1, 4, 9)
	if !moved {
		t.Fatal("4'ten 9'ye taşıma başarısız")
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
