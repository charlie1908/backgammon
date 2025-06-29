package core

import (
	"backgammon/models"
	"crypto/rand"
	"sort"
)

// Global hareket sırası sayacı, oyun her başladığında sıfırlanmalı
var globalMoveOrder int64 = 0

func ResetMoveOrder() {
	globalMoveOrder = 0
}

func GetInitialStones() []*LogicalCoordinate {
	var stones []*LogicalCoordinate

	type stackKey struct {
		PointIndex int
		Player     int
	}

	stackCount := make(map[stackKey]int)

	// Başlangıç taşları için aynı MoveOrder veriyoruz, çünkü başlangıçta hareket yok
	const initialMoveOrder = 0

	addStack := func(pointIndex, count, player int) {
		for i := 0; i < count; i++ {
			key := stackKey{PointIndex: pointIndex, Player: player}
			stackIndex := stackCount[key]
			isTop := (i == count-1)

			stones = append(stones, &LogicalCoordinate{
				PointIndex:   pointIndex,
				PositionType: PositionTypeEnum.Point,
				Player:       player,
				StackIndex:   stackIndex,
				IsTop:        isTop,
				MoveOrder:    initialMoveOrder,
			})

			stackCount[key]++
		}
	}

	// Oyuncu 1 taşları
	addStack(0, 2, 1)
	addStack(11, 5, 1)
	addStack(16, 3, 1)
	addStack(18, 5, 1)

	// Oyuncu 2 taşları
	addStack(23, 2, 2)
	addStack(12, 5, 2)
	addStack(7, 3, 2)
	addStack(5, 5, 2)

	return stones
}

func SortStonesByPlayerPointAndStackDesc(stones []*LogicalCoordinate) {
	sort.Slice(stones, func(i, j int) bool {
		if stones[i].Player == stones[j].Player {
			if stones[i].PointIndex == stones[j].PointIndex {
				return stones[i].StackIndex > stones[j].StackIndex
			}
			return stones[i].PointIndex < stones[j].PointIndex
		}
		return stones[i].Player < stones[j].Player // Önce Player1 (1), sonra Player2 (2)
	})
}

func SortStonesByPlayerPointAndStackAsc(stones []*LogicalCoordinate) {
	sort.Slice(stones, func(i, j int) bool {
		if stones[i].Player == stones[j].Player {
			if stones[i].PointIndex == stones[j].PointIndex {
				return stones[i].StackIndex < stones[j].StackIndex
			}
			return stones[i].PointIndex < stones[j].PointIndex
		}
		return stones[i].Player < stones[j].Player // Önce Player1 (1), sonra Player2 (2)
	})
}

// Taşı hedef noktaya taşır ve MoveOrder günceller.
// Sadece taşın eski ve yeni noktalarındaki taşlar güncellenir.
func MoveTopStoneAndUpdate(stones []*LogicalCoordinate, player int, fromPointIndex int, toPointIndex int) ([]*LogicalCoordinate, bool) {
	//Karsi rakibin birden fazla tasi var mi ? ilgili PointIndex'inde..
	if !CanMoveToPoint(stones, player, toPointIndex) {
		return stones, false // Hareket yasak
	}

	// 💥 Kırma kontrolü — hedef noktada 1 adet rakip taşı varsa kır ve bar'a gönder
	if CountOpponentStonesAtPoint(stones, player, toPointIndex) == 1 {
		for i, stone := range stones {
			if stone.PointIndex == toPointIndex && stone.Player != player {
				stones[i].PointIndex = -1
				stones[i].PositionType = PositionTypeEnum.Bar
				stones[i].StackIndex = 0
				stones[i].IsTop = true

				globalMoveOrder++
				stones[i].MoveOrder = globalMoveOrder
				break
			}
		}
		stones = UpdateStacks(stones, []int{-1}) // Bar'daki taşların stack bilgisi güncellenir
	}

	var moveIndex int = -1
	// fromPointIndex'teki en üstteki ve player'a ait taşı bul
	for i, stone := range stones {
		if stone.PointIndex == fromPointIndex && stone.IsTop && stone.Player == player {
			moveIndex = i
			break
		}
	}

	if moveIndex == -1 {
		// Taş bulunamadı, hareket yapılmadı
		return stones, false
	}

	oldPointIndex := stones[moveIndex].PointIndex
	stones[moveIndex].PointIndex = toPointIndex
	stones[moveIndex].PositionType = PositionTypeEnum.Point

	globalMoveOrder++
	stones[moveIndex].MoveOrder = globalMoveOrder

	// Güncellemeleri yap
	stones = UpdateStacks(stones, []int{oldPointIndex, toPointIndex})

	return stones, true
}

func CanMoveToPoint(stones []*LogicalCoordinate, player int, toPointIndex int) bool {
	//Tasin oynanacagi yerde rakip tas var mi ve en fazla 1 tane mi ?
	opponentCount := CountOpponentStonesAtPoint(stones, player, toPointIndex)

	// Eğer rakip taş sayısı 0 veya 1 ise hareket mümkün
	// 2 veya daha fazla rakip taş varsa hareket yasak
	return opponentCount <= 1
}

func CountOpponentStonesAtPoint(stones []*LogicalCoordinate, player int, pointIndex int) int {
	count := 0
	for _, stone := range stones {
		if stone.PointIndex == pointIndex && stone.Player != player {
			count++
		}
	}
	return count
}

// Bar’da taşı olan oyuncu başka hamle yapamaz!
func PlayerMustEnterFromBar(stones []*LogicalCoordinate, player int) bool {
	for _, stone := range stones {
		if stone.Player == player && stone.PositionType == PositionTypeEnum.Bar {
			return true
		}
	}
	return false
}

// Kirik tasi girebilecek mi? Girerse hangi zar veya zarlar ile girebilecek.
func CanEnterFromBar(stones []*LogicalCoordinate, player int, dice []int) (bool, []int) {
	var enterableDice []int
	for _, die := range dice {
		entryPoint := GetEntryPoint(player, die)
		if entryPoint < 0 || entryPoint > 23 {
			continue
		}
		if CanMoveToPoint(stones, player, entryPoint) {
			enterableDice = append(enterableDice, die)
		}
	}
	return len(enterableDice) > 0, enterableDice
}

// Player'a gore atilan zarin tavlada karsiligi
func GetEntryPoint(player int, die int) int {
	if player == 1 {
		return die - 1 // 1 → 0, 6 → 5 => 0..5 (kendi başlangıç)
	} else if player == 2 {
		return 24 - die // 1 → 23, 6 → 18 => 18..23 (kendi başlangıç)
	}
	return -1
}

// Önce bar girişi kontrol edilir. Kirik tas var mi ?
func IsBarEntryAllowed(stones []*LogicalCoordinate, player int, dice []int) models.MoveCheckResult {
	result := models.MoveCheckResult{}

	if PlayerMustEnterFromBar(stones, player) {
		result.FromBar = true
		canEnter, enterableDice := CanEnterFromBar(stones, player, dice)
		result.CanEnterFromBar = canEnter
		result.EnterableDice = enterableDice
		result.Allowed = canEnter
		return result
	}

	// Bar’da taş yok, izin yok demiyoruz, sadece bar durumu değil. Yani Kirik tasi yok..
	result.FromBar = false
	result.Allowed = false
	return result
}

func IsNormalMoveAllowed_Old(stones []*LogicalCoordinate, player int, fromPointIndex int, toPointIndex int, dice []int) models.MoveCheckResult {
	result := models.MoveCheckResult{}

	// Normal hamle kontrolü
	canMove := CanMoveToPoint(stones, player, toPointIndex)
	if canMove {
		result.CanMoveNormally = true
		result.NormalDice = dice // opsiyonel: zar mesafesiyle uyum kontrolü yapılabilir
		result.Allowed = true
	} else {
		result.Allowed = false
	}

	result.FromBar = false
	return result
}

// Önce tek zarla hareket deneniyor,
// Sonra zarların iki farklı sırasıyla adım adım kontrol ediliyor,
// Eğer biri uygunsa hareket mümkün kabul ediliyor.
func IsNormalMoveAllowed(stones []*LogicalCoordinate, player int, fromPointIndex int, toPointIndex int, dice []int) models.MoveCheckResult {
	result := models.MoveCheckResult{}

	// Oyuncuya göre hareket yönü belirle
	direction := 1
	if player == 2 {
		direction = -1
	}

	distance := (toPointIndex - fromPointIndex) * direction
	if distance < 0 {
		// Negatif mesafe hareket değil, izin yok
		result.Allowed = false
		return result
	}

	canMove := false

	// Tek zarla hareket kontrolü sadece mesafe 6 veya daha küçükse yapılır
	if distance <= 6 {
		for _, d := range dice {
			if d == distance && CanMoveToPoint(stones, player, fromPointIndex+direction*d) {
				canMove = true
				break
			}
		}
	}

	if !canMove && len(dice) == 2 {
		d1, d2 := dice[0], dice[1]

		// Önce d1 sonra d2 ile hareket dene
		posAfterFirst := fromPointIndex + direction*d1
		posAfterSecond := posAfterFirst + direction*d2
		if CanMoveToPoint(stones, player, posAfterFirst) && CanMoveToPoint(stones, player, posAfterSecond) && posAfterSecond == toPointIndex {
			canMove = true
		}

		// Önce d2 sonra d1 ile hareket dene
		posAfterFirst = fromPointIndex + direction*d2
		posAfterSecond = posAfterFirst + direction*d1
		if CanMoveToPoint(stones, player, posAfterFirst) && CanMoveToPoint(stones, player, posAfterSecond) && posAfterSecond == toPointIndex {
			canMove = true
		}
	}

	result.FromBar = false
	result.CanMoveNormally = canMove
	result.NormalDice = dice
	result.Allowed = canMove

	return result
}

/*func MoveStoneAndUpdate(stones []*LogicalCoordinate, index int, newPointIndex int, player int) ([]*LogicalCoordinate, bool) {

	// Taşın belirtilen oyuncuya ait olup olmadığını kontrol et
	if stones[index].Player != player {
		return stones, false
	}

	// Taşın yığının en üstünde olup olmadığını kontrol et
	if !stones[index].IsTop {
		return stones, false
	}

	oldPointIndex := stones[index].PointIndex

	stones[index].PointIndex = newPointIndex
	stones[index].PositionType = PositionTypeEnum.Point

	globalMoveOrder++
	stones[index].MoveOrder = globalMoveOrder

	return UpdateStacks(stones, []int{oldPointIndex, newPointIndex}), true
}*/

// Sadece verilen noktaların taşlarını günceller
func UpdateStacks_Old(stones []*LogicalCoordinate, pointsToUpdate []int) []*LogicalCoordinate {
	pointGroups := make(map[int][]*LogicalCoordinate)
	for _, stone := range stones {
		pointGroups[stone.PointIndex] = append(pointGroups[stone.PointIndex], stone)
	}

	pointsSet := make(map[int]bool)
	for _, p := range pointsToUpdate {
		pointsSet[p] = true
	}

	for pointIndex := range pointsSet {
		group := pointGroups[pointIndex]

		// MoveOrder'a göre sırala: küçük olan daha erken taş (altta), büyük olan üstte
		sort.Slice(group, func(i, j int) bool {
			return group[i].MoveOrder < group[j].MoveOrder
		})

		for i := range group {
			group[i].StackIndex = i
			group[i].IsTop = (i == len(group)-1)
		}

		pointGroups[pointIndex] = group
	}

	// Güncellenmiş taşları düz listeye toplar
	var updatedStones []*LogicalCoordinate
	for _, group := range pointGroups {
		updatedStones = append(updatedStones, group...)
	}

	return updatedStones
}

// Sadece verilen noktaların taşlarını günceller
func UpdateStacks(stones []*LogicalCoordinate, pointsToUpdate []int) []*LogicalCoordinate {
	// pointsToUpdate'deki her nokta için işlemi yap
	for _, pointIndex := range pointsToUpdate {
		// İlgili noktadaki taşları filtrele
		group := []*LogicalCoordinate{}
		for _, stone := range stones {
			if stone.PointIndex == pointIndex {
				group = append(group, stone)
			}
		}

		// MoveOrder'a göre sırala (küçük olan altta)
		sort.Slice(group, func(i, j int) bool {
			return group[i].MoveOrder < group[j].MoveOrder
		})

		// StackIndex ve IsTop değerlerini güncelle
		for i := range group {
			group[i].StackIndex = i
			group[i].IsTop = (i == len(group)-1)
		}
	}

	// stones zaten pointer tip olduğu için güncellemeler doğrudan yansıdı
	return stones
}

// Zar atma fonksiyonları...

func rollDie() (int, error) {
	for {
		b := make([]byte, 1)
		if _, err := rand.Read(b); err != nil {
			return 0, err
		}
		if b[0] < 252 {
			return int(b[0]%6) + 1, nil
		}
	}
}

func RollDice() (int, int, error) {
	d1, err := rollDie()
	if err != nil {
		return 0, 0, err
	}
	d2, err := rollDie()
	if err != nil {
		return 0, 0, err
	}
	return d1, d2, nil
}
