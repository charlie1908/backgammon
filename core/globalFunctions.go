package core

import (
	"backgammon/models"
	"crypto/rand"
	"slices"
	"sort"
)

// Global hareket sırası sayacı, oyun her başladığında sıfırlanmalı
var globalMoveOrder int64 = 0

// Yeni oyunda globalMoveOrder'i sifirlar.
func ResetMoveOrder() {
	globalMoveOrder = 0
}

// Tavlada baslangic taslarini dizer.
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

// Tum oyuncu taslarini once Playerlara gore sonra PontIndex ve en son StackIndex'e gore Desc dizer
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

// 1-) Kirik tasi var mi "IsBarEntryAllowed()" ve Taslar istenen PointIndex'e hareket edebiliyor mu "IsNormalMoveAllowed()" bakildiktan sonra cagrilir.
// 2-) Taslarin PointIndexlerini gercekten yer degistirerek ayrica ayrildigi ve tasindigi guruplarin Stackindexlerini degistirerek taslari gerckten hareket ettirir.
// 3-) Arada rakibin kirilan tasi var ise kirar..
func MoveTopStoneAndUpdate(stones []*LogicalCoordinate, player int, fromPointIndex int, toPointIndex int) ([]*LogicalCoordinate, bool) {

	//Normalde IsBarEntryAllowed(), IsNormalMoveAllowed() functionlari bu Methodun basinda cagrilip bakilmali. Ama gene de bir ekstra 3 kontrol koyma ihtiyaci duydum.

	// 1. Gecerli oyuncun belirtilen yerde en ustte tasi var mi ?
	if !PlayerHasTopStoneAt(stones, player, fromPointIndex) {
		return stones, false // Oyuncunun bu noktada üstte taşı yok, hareket edemez
	}

	// 2. Hedef nokta geçerli mi?
	//if toPointIndex < 0 || toPointIndex >= 24 {
	//Taslar Toplana da bilir..
	if toPointIndex < 0 || toPointIndex > 24 {
		return stones, false
	}

	// toPointIndex 24 değilse karsi rakibin taslari ile [len(stone)>1] blokaj ve kırma kontrollerini yap
	if toPointIndex != 24 {
		//Karsi rakibin birden fazla tasi var mi ? ilgili PointIndex'inde..
		// 3. Hedef nokta rakip tarafindan blokaj altinda mi ?
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
	//stones[moveIndex].PositionType = PositionTypeEnum.Point
	if toPointIndex == 24 {
		stones[moveIndex].PositionType = PositionTypeEnum.OffBoard
	} else {
		stones[moveIndex].PositionType = PositionTypeEnum.Point
	}

	globalMoveOrder++
	stones[moveIndex].MoveOrder = globalMoveOrder

	// StackIndex Güncellemelerini yap
	stones = UpdateStacks(stones, []int{oldPointIndex, toPointIndex})

	return stones, true
}

// Oynayacak Player'in belirtilen from noktasindaki taş dilimlerinde en üstte taşa sahip olup olmadığını kontrol eder.
func PlayerHasTopStoneAt(stones []*LogicalCoordinate, player int, pointIndex int) bool {
	if pointIndex == 24 {
		return false // offboard'daki taşlar zaten üstte olamaz, oynanamaz
	}

	for _, stone := range stones {
		if stone.Player == player && stone.PointIndex == pointIndex && stone.IsTop {
			return true
		}
	}
	return false
}

// Tasin oynayacagi PointIndex bos mu, ya da rakibin sadece 1 pulu mu var ?
func CanMoveToPoint(stones []*LogicalCoordinate, player int, toPointIndex int) bool {
	// 24 = OffBoard (toplama alanı), buralara doğrudan gidilebilir
	if toPointIndex == 24 {
		return true
	}

	//Tasin oynanacagi yerde rakip tas var mi ve en fazla 1 tane mi ?
	opponentCount := CountOpponentStonesAtPoint(stones, player, toPointIndex)

	// Eğer rakip taş sayısı 0 veya 1 ise hareket mümkün
	// 2 veya daha fazla rakip taş varsa hareket yasak
	return opponentCount <= 1
}

// Rakip oyuncunun belirtilen noktada (PointIndex) kac tasi var onu hesaplar..
func CountOpponentStonesAtPoint(stones []*LogicalCoordinate, player int, pointIndex int) int {
	if pointIndex == 24 {
		return 0 // Toplanmış taşlar oyun dışında
	}

	count := 0
	for _, stone := range stones {
		if stone.PointIndex == pointIndex && stone.Player != player {
			count++
		}
	}
	return count
}

// Bar’da Kirik taşı olan oyuncu başka hamle yapamaz!
func PlayerMustEnterFromBar(stones []*LogicalCoordinate, player int) bool {
	for _, stone := range stones {
		if stone.Player == player && stone.PositionType == PositionTypeEnum.Bar {
			return true
		}
	}
	return false
}

// Kirik tasi girebilecek mi? Girerse hangi zar veya zarlar ile girebilecek.
/*func CanEnterFromBar(stones []*LogicalCoordinate, player int, dice []int) (bool, []int) {
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
}*/

// Kirik butun taslari girebilecek mi? Girerse hangi zar veya zarlar ile girilebilecek. Double(Cift) zar destegi icin ExpandDice() function kullan..
func CanAllBarStonesEnter(stones []*LogicalCoordinate, player int, dice []int) (bool, []int) {
	var usedDice []int
	remainingDice := append([]int(nil), dice...) // Zarları kopyala
	barStonesCount := 0

	// Bar'daki kendi kirik taşlarını say
	for _, stone := range stones {
		if stone.Player == player && stone.PositionType == PositionTypeEnum.Bar {
			barStonesCount++
		}
	}

	if barStonesCount == 0 {
		return true, []int{} // Zaten kirik bar taşı yok
	}

	// Her bar taşı için bir zar bulmaya çalış
	for i := 0; i < barStonesCount; i++ {
		found := false
		for j, die := range remainingDice {
			entryPoint := GetEntryPoint(player, die)
			if CanMoveToPoint(stones, player, entryPoint) {
				usedDice = append(usedDice, die) //Ise yarar zar, kirik bir tas icin kullanilir.
				// Bu zarı kullan, listeden çıkar
				remainingDice = append(remainingDice[:j], remainingDice[j+1:]...)
				found = true
				break
			}
		}
		if !found {
			return false, usedDice // Bu taş için zar yok, işlem başarısız
		}
	}

	return true, usedDice // Her taş için zar bulundu
}

// Player'a gore atilan zarin, tavlada PontIndex karsiligi bulunur.
// PointIndex 0'dan baslar zarlarin PointIndex karsiligi (-1) 1 eksigidir.
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
		//canEnter, enterableDice := CanEnterFromBar(stones, player, dice)
		canEnter, enterableDice := CanAllBarStonesEnter(stones, player, dice)
		result.CanEnterFromBar = canEnter
		result.EnterableDice = enterableDice
		result.UsedDice = enterableDice
		result.Allowed = canEnter

		//Kullanilmayan geride kalan zarlar hesaplanir
		if canEnter && len(enterableDice) > 0 {
			result.RemainingDice = CalculateRemainingDice(dice, enterableDice)
		} else {
			result.RemainingDice = dice // veya boş liste []int{}, tercihe göre
		}

		return result
	}

	// Bar’da taş yok, izin yok demiyoruz, sadece bar durumu değil. Yani Kirik tasi yok..
	result.FromBar = false
	//result.Allowed = false
	result.Allowed = true
	result.RemainingDice = dice
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
// Cift zar var ise 3lu ve 4lu hareketi kontrol ediliyor.
func IsNormalMoveAllowed(stones []*LogicalCoordinate, player int, fromPointIndex int, toPointIndex int, dice []int) models.MoveCheckResult {
	result := models.MoveCheckResult{}
	usedDie := []int{}

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
				usedDie = []int{d}
				break
			}
		}
	}

	// İki farklı zarla (normal durum)
	if !canMove && len(dice) == 2 {
		d1, d2 := dice[0], dice[1]

		// Önce d1 sonra d2 ile hareket dene
		posAfterFirst := fromPointIndex + direction*d1
		posAfterSecond := posAfterFirst + direction*d2
		if CanMoveToPoint(stones, player, posAfterFirst) && CanMoveToPoint(stones, player, posAfterSecond) && posAfterSecond == toPointIndex {
			usedDie = []int{d1, d2}
			canMove = true
		}

		if !canMove {
			// Önce d2 sonra d1 ile hareket dene
			posAfterFirst = fromPointIndex + direction*d2
			posAfterSecond = posAfterFirst + direction*d1
			if CanMoveToPoint(stones, player, posAfterFirst) && CanMoveToPoint(stones, player, posAfterSecond) && posAfterSecond == toPointIndex {
				usedDie = []int{d1, d2}
				canMove = true
			}
		}
	}

	// *** 3 zarla hareket kontrolü buraya eklendi ***
	if !canMove && len(dice) == 4 {
		// 3 zar kullanarak hedefe varma kontrolü (örn: ilk 3 zarla)
		sum := 0
		canReach := true
		for i := 1; i <= 3; i++ {
			sum += dice[0] // hepsi aynı zar (double) olduğu için dice[0]
			intermediate := fromPointIndex + direction*sum

			if i < 3 && !CanMoveToPoint(stones, player, intermediate) { // ara noktalar engelli mi?
				canReach = false
				break
			}
		}
		if canReach && (fromPointIndex+direction*sum) == toPointIndex && CanMoveToPoint(stones, player, toPointIndex) {
			usedDie = []int{dice[0], dice[1], dice[2]} // ilk 3 zar kullanıldı
			canMove = true
		}
	}

	// Double zar durumunda 4 adımda hedefe ulaşabiliyor muyuz?
	if !canMove && len(dice) == 4 {
		sum := 0
		for i := 1; i <= 4; i++ {
			sum += dice[0] // Hepsi aynı olduğundan dice[0] yeterlidir
			intermediate := fromPointIndex + direction*sum

			// Ara noktalarda rakip taşlar varsa hareket geçersiz
			if i < 4 && !CanMoveToPoint(stones, player, intermediate) {
				break
			}

			// Dördüncü adımda hedefe ulaşılmış ve taş konulabilir mi?
			if i == 4 && intermediate == toPointIndex && CanMoveToPoint(stones, player, toPointIndex) {
				usedDie = []int{dice[0], dice[1], dice[2], dice[3]} // ya da 4x dice[0]
				canMove = true
			}
		}
	}

	//Kullanilip geride kalan kullanilmamis zarlar burada tanimlanir.
	if canMove && len(usedDie) > 0 {
		result.RemainingDice = CalculateRemainingDice(dice, usedDie)
		result.UsedDice = usedDie
	} else {
		result.RemainingDice = dice
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

// Sadece verilen noktaların taşlarını günceller. Hem PointIndex hem de StackIndex guncellenir.
func UpdateStacks(stones []*LogicalCoordinate, pointsToUpdate []int) []*LogicalCoordinate {
	// pointsToUpdate'deki her nokta için işlemi yap
	for _, pointIndex := range pointsToUpdate {
		// 24 (OffBoard) için stack güncellemesi yapma
		if pointIndex == 24 {
			continue
		}

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

func RollDice_Old() (int, int, error) {
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

func RollDice() ([]int, error) {
	d1, err := rollDie()
	if err != nil {
		return []int{0, 0}, err
	}
	d2, err := rollDie()
	if err != nil {
		return []int{0, 0}, err
	}
	return ExpandDice([]int{d1, d2}), err
}

// Double gelen zarlari 4ler
// originalDice := []int{6, 6}            // double
// expandedDice := ExpandDice(originalDice) // => [6, 6, 6, 6]
func ExpandDice(dice []int) []int {
	if len(dice) == 2 && dice[0] == dice[1] {
		// Double zar atılmış, 4 kere kullanılmalı
		return []int{dice[0], dice[0], dice[0], dice[0]}
	}
	// Normal zar
	return dice
}

// Zarlardan ise yararlar kullanildiktan sonra geri kalan zarlar..
// dice = [4,4,4,4], used = [4,4,4] → kalan: [4]
func CalculateRemainingDice(dice []int, used []int) []int {
	remaining := make([]int, len(dice))
	copy(remaining, dice)

	for _, usedDie := range used {
		for i, remDie := range remaining {
			if remDie == usedDie {
				// Bu zarı kullandık, listeden çıkar
				remaining = append(remaining[:i], remaining[i+1:]...)
				break
			}
		}
	}

	return remaining
}

// Oynanabilir zarlara gore belirlenen PointIndex'deki tas ile gidilebilecek PointIndexler yani hamleler hesaplanir.
func GetPossibleMovePoints(
	stones []*LogicalCoordinate,
	player int,
	fromPointIndex int,
	remainingDice []int,
) []int {
	direction := 1
	if player == 2 {
		direction = -1
	}

	var possiblePoints []int

	if !PlayerHasTopStoneAt(stones, player, fromPointIndex) {
		return possiblePoints
	}

	for _, die := range remainingDice {
		toPointIndex := fromPointIndex + direction*die

		if toPointIndex < 0 || toPointIndex >= 24 {
			continue
		}

		if CanMoveToPoint(stones, player, toPointIndex) {
			possiblePoints = append(possiblePoints, toPointIndex)
		}
	}

	return possiblePoints
}
func getBearoffRangeForPlayer(player int) []int {
	if player == 1 {
		return []int{18, 19, 20, 21, 22, 23}
	}
	return []int{0, 1, 2, 3, 4, 5}
}

// Toplamak icin tum taslar olmasi gereken yerde mi ?
func AreAllStonesInBearOffArea(stones []*LogicalCoordinate, player int) bool {
	bearOffRange := getBearoffRangeForPlayer(player)

	pointSet := make(map[int]bool)
	for _, p := range bearOffRange {
		pointSet[p] = true
	}

	for _, s := range stones {
		if s.Player != player {
			continue
		}

		// Toplanmış taş (PointIndex == 24) her zaman geçerli
		if s.PointIndex == 24 {
			continue
		}

		//Kirik tasi varsa toplama yapilamaz
		if s.PointIndex == -1 {
			return false
		}

		if !pointSet[s.PointIndex] {
			return false
		}
	}
	return true
}

func removeDieAtIndex(dice []int, index int) []int {
	return append(append([]int{}, dice[:index]...), dice[index+1:]...)
}

// Bu Fonksiyonun sonrasinda => "MoveTopStoneAndUpdate(stones, player, fromIndex, 24)" tasi toplamak icin cagrilir...
func CanBearOffStone(stones []*LogicalCoordinate, player int, pointIndex int, dice []int) (result bool, remainingDice []int, usedDice []int) {
	// 1. Tüm taşlar Player icin toplama alanında mı?
	if !AreAllStonesInBearOffArea(stones, player) {
		return false, dice, nil
	}

	//Gerek yok ama gene de koydum..
	// 2. Toplama alanı dışında bir taş işaretlenmişse izin verilmez
	bearOffRange := getBearoffRangeForPlayer(player)
	if !slices.Contains(bearOffRange, pointIndex) {
		return false, dice, nil
	}

	// 3. İlgili noktada oyuncunun top taşı var mı?
	if !PlayerHasTopStoneAt(stones, player, pointIndex) {
		return false, dice, nil
	}

	// Mesafe hesabı => Zar karsiligi
	var distance int
	if player == 1 {
		distance = 23 - pointIndex + 1
	} else {
		distance = pointIndex + 1
	}

	// En küçük geçerli zarı bul
	for i, die := range dice {
		if die == distance {
			// Tam zarla çıkabilir
			remaining := removeDieAtIndex(dice, i)
			return true, remaining, []int{die}
		}
	}

	// Daha büyük zar var mı? Atilan zardan ornek 4'den daha geride 5-6'da tas var mi diye bakilir..
	for i, die := range dice {
		if die > distance {
			// Daha geride taş var mı kontrol et
			for _, s := range stones {
				if s.Player != player || s.PointIndex == 24 || s.PointIndex == pointIndex {
					continue
				}
				if player == 1 && s.PointIndex < pointIndex {
					return false, dice, nil
				}
				if player == 2 && s.PointIndex > pointIndex {
					return false, dice, nil
				}
			}
			remaining := removeDieAtIndex(dice, i)
			return true, remaining, []int{die}
		}
	}

	return false, dice, nil
}

func IsFinishedForPlayer(stones []*LogicalCoordinate, player int) bool {
	for _, s := range stones {
		if s.Player == player && s.PointIndex != 24 {
			return false
		}
	}
	return true
}
func IsFinishedForPlayer_MoreCheck(stones []*LogicalCoordinate, player int) bool {
	count := 0
	for _, s := range stones {
		if s.Player == player {
			if s.PointIndex != 24 {
				return false
			}
			count++
		}
	}
	return count == 15 // Tavlada bir oyuncunun 15 taşı vardır
}
