package core

import (
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
func UpdateStacks(stones []*LogicalCoordinate, pointsToUpdate []int) []*LogicalCoordinate {
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
