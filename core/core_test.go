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
	fromPoint := 0
	toPoint := 1 // Geçerli nokta, rakip az ya da yok

	updatedStones, moved := core.MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
	if !moved {
		t.Errorf("Taş hareket etmedi ama hareket mümkün olmalıydı.")
	}

	if moved {
		t.Logf("Taş başarıyla hareket etti: Player %d, %d -> %d", player, fromPoint, toPoint)
		core.SortStonesByPlayerPointAndStackDesc(updatedStones)
		t.Log("Taşların Son Durumu:")
		for _, stone := range updatedStones {
			t.Logf("PointIndex: %d, Player: %d, StackIndex: %d, IsTop: %v, MoveOrder: %d",
				stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
		}
	}
}

func TestMoveTopStoneAndUpdate_InvalidMove(t *testing.T) {
	stones := core.GetInitialStones()

	player := 1
	fromPoint := 0
	toPoint := 5 // Burada oyuncu 2'nin 5 taşı var, geçersiz hareket

	updatedStones, moved := core.MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
	if moved {
		t.Errorf("Taş hareket etti ama hareket yasak olmalıydı.")
	}

	// Taşın konumu değişmemeli
	for _, s := range updatedStones {
		if s.Player == player && s.PointIndex == toPoint {
			t.Errorf("Taş yanlışlıkla rakip taşların olduğu noktaya taşındı.")
		}
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

	canEnter, enterableDice := core.CanEnterFromBar(stones, player, dice)
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

	dice := []int{1}
	canEnter, diceList := core.CanEnterFromBar(stones, 2, dice)

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
