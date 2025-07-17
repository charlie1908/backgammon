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

// 1-) Kirik tasi var mi "IsAllBarEntryAllowed()" ve Taslar istenen PointIndex'e hareket edebiliyor mu "IsNormalMoveAllowed()" bakildiktan sonra cagrilir.
// 2-) Taslarin PointIndexlerini gercekten yer degistirerek ayrica ayrildigi ve tasindigi guruplarin Stackindexlerini degistirerek taslari gerckten hareket ettirir.
// 3-) Arada rakibin kirilan tasi var ise kirar..
func MoveTopStoneAndUpdate(stones []*LogicalCoordinate, player int, fromPointIndex int, toPointIndex int) ([]*LogicalCoordinate, bool, []*LogicalCoordinate) {

	var brokenStones []*LogicalCoordinate //Varsa kirilan taslar
	//Normalde IsAllBarEntryAllowed(), IsNormalMoveAllowed() functionlari bu Methodun basinda cagrilip bakilmali. Ama gene de bir ekstra 3 kontrol koyma ihtiyaci duydum.

	// 1. Gecerli oyuncun belirtilen yerde en ustte tasi var mi ?
	if !playerHasTopStoneAt(stones, player, fromPointIndex) {
		return stones, false, brokenStones // Oyuncunun bu noktada üstte taşı yok, hareket edemez
	}

	// 2. Hedef nokta geçerli mi?
	//if toPointIndex < 0 || toPointIndex >= 24 {
	//Taslar Toplana da bilir..
	if toPointIndex < 0 || toPointIndex > 24 {
		return stones, false, brokenStones
	}

	// toPointIndex 24 değilse karsi rakibin taslari ile [len(stone)>1] blokaj ve kırma kontrollerini yap
	if toPointIndex != 24 {
		//Karsi rakibin birden fazla tasi var mi ? ilgili PointIndex'inde..
		// 3. Hedef nokta rakip tarafindan blokaj altinda mi ?
		if !canMoveToPoint(stones, player, toPointIndex) {
			return stones, false, brokenStones // Hareket yasak
		}

		// 💥 Kırma kontrolü — hedef noktada 1 adet rakip taşı varsa kır ve bar'a gönder
		if countOpponentStonesAtPoint(stones, player, toPointIndex) == 1 {
			var brokenPlayer int
			for i, stone := range stones {
				if stone.PointIndex == toPointIndex && stone.Player != player {
					// Kırılan taşın oyuncusu
					brokenPlayer = stone.Player

					// 💾 Kırılan taşı brokenStones içine ekle
					original := *stones[i] // struct değer kopyası
					brokenStones = append(brokenStones, &original)

					stones[i].PointIndex = -1
					stones[i].PositionType = PositionTypeEnum.Bar
					stones[i].StackIndex = 0
					stones[i].IsTop = true

					globalMoveOrder++
					stones[i].MoveOrder = globalMoveOrder

					break
				}
			}
			// Sadece kırılan oyuncunun barındaki taşların stack bilgisi güncellenir
			stones = updateStacks(stones, []int{-1}, brokenPlayer)
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
		return stones, false, brokenStones
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
	/*stones = updateStacks(stones, []int{oldPointIndex, toPointIndex})*/
	// StackIndex Güncellemelerini yap
	if fromPointIndex == -1 {
		// Oyuncu bar'dan çıktıysa, hem kendi bar'ı hem de hedefi günceller. Player bazli
		stones = updateStacks(stones, []int{-1, toPointIndex}, player)
	} else {
		stones = updateStacks(stones, []int{oldPointIndex, toPointIndex})
	}

	return stones, true, brokenStones
}

// Oynayacak Player'in belirtilen from noktasindaki taş dilimlerinde en üstte taşa sahip olup olmadığını kontrol eder.
func playerHasTopStoneAt(stones []*LogicalCoordinate, player int, pointIndex int) bool {
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
func canMoveToPoint(stones []*LogicalCoordinate, player int, toPointIndex int) bool {
	// 24 = OffBoard (toplama alanı), buralara doğrudan gidilebilir
	if toPointIndex == 24 {
		return true
	}

	//Tasin oynanacagi yerde rakip tas var mi ve en fazla 1 tane mi ?
	opponentCount := countOpponentStonesAtPoint(stones, player, toPointIndex)

	// Eğer rakip taş sayısı 0 veya 1 ise hareket mümkün
	// 2 veya daha fazla rakip taş varsa hareket yasak
	return opponentCount <= 1
}

// Rakip oyuncunun belirtilen noktada (PointIndex) kac tasi var onu hesaplar..
func countOpponentStonesAtPoint(stones []*LogicalCoordinate, player int, pointIndex int) int {
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
		if canMoveToPoint(stones, player, entryPoint) {
			enterableDice = append(enterableDice, die)
		}
	}
	return len(enterableDice) > 0, enterableDice
}*/

// Kirik butun taslari girebilecek mi? Girerse hangi zar veya zarlar ile girilebilecek. Double(Cift) zar destegi icin ExpandDice() function kullan..
// Her bir kırık taş için zar bulabilir miyim?
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
			if canMoveToPoint(stones, player, entryPoint) {
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

// Kirik taslari girebilecegin TUM zarlari geri doner.
func GetEnterableBarDice(stones []*LogicalCoordinate, player int, dice []int) (enterableDice []int) {
	for _, die := range dice {
		entryPoint := GetEntryPoint(player, die)
		if canMoveToPoint(stones, player, entryPoint) {
			enterableDice = append(enterableDice, die)
		}
	}
	return
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
// Tum kiriklar girilebiliyor mu ve kullanilmayan zarlar hangisi
// Tum kiriklar girilmez ise Allowed ve CanEnterFromBar = false olur..
func IsAllBarEntryAllowed(stones []*LogicalCoordinate, player int, dice []int) models.MoveCheckResult {
	result := models.MoveCheckResult{}

	if PlayerMustEnterFromBar(stones, player) {
		result.FromBar = true
		//canEnter, enterableDice := CanEnterFromBar(stones, player, dice)
		canEnter, enterableDice := CanAllBarStonesEnter(stones, player, dice) // Burada kirigi iceri sokabilen sadece ilk bulunan zarlar her bir kirik icin bulunup butun taslar girilebiliyor mu ve geride kalan kullanilmayan zarlar hangisi diye bakiliyor.

		result.CanEnterFromBar = canEnter
		result.EnterableDice = enterableDice
		result.UsedDice = enterableDice
		result.Allowed = canEnter

		//Kullanilmayan geride kalan zarlar hesaplanir
		if canEnter && len(enterableDice) > 0 {
			result.RemainingDice = calculateRemainingDice(dice, enterableDice)
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

// Önce bar girişi kontrol edilir. Kirik tas var mi ? Varsa kirigi iceri sokabilen tum zarlar ve hicbir kirigi iceri sokamayan zarlar belirlenir.
func IsAnyBarEntryAllowed(stones []*LogicalCoordinate, player int, dice []int) models.MoveCheckResult {
	result := models.MoveCheckResult{}

	if PlayerMustEnterFromBar(stones, player) {
		result.FromBar = true
		enterableDice := GetEnterableBarDice(stones, player, dice) //Kirik veya kiriklari iceri sokabilecek tum zarlar geri donulur. RemaningDice ile hicbir kirigi iceri sokamayan zarlar belli olur.
		canEnter := len(enterableDice) > 0

		result.CanEnterFromBar = canEnter
		result.EnterableDice = enterableDice
		result.UsedDice = enterableDice
		result.Allowed = canEnter

		//Kullanilmayan geride kalan zarlar hesaplanir
		if canEnter && len(enterableDice) > 0 {
			result.RemainingDice = calculateRemainingDice(dice, enterableDice)
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
	canMove := canMoveToPoint(stones, player, toPointIndex)
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
			if d == distance && canMoveToPoint(stones, player, fromPointIndex+direction*d) {
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
		if canMoveToPoint(stones, player, posAfterFirst) && canMoveToPoint(stones, player, posAfterSecond) && posAfterSecond == toPointIndex {
			usedDie = []int{d1, d2}
			canMove = true
		}

		if !canMove {
			// Önce d2 sonra d1 ile hareket dene
			posAfterFirst = fromPointIndex + direction*d2
			posAfterSecond = posAfterFirst + direction*d1
			if canMoveToPoint(stones, player, posAfterFirst) && canMoveToPoint(stones, player, posAfterSecond) && posAfterSecond == toPointIndex {
				usedDie = []int{d1, d2}
				canMove = true
			}
		}
	}

	// *** Double'dan geriye 3 zar kaldıysa (örnek: 1,1,1), 3 zarla hareket kontrolü ***
	if !canMove && len(dice) == 3 {
		sum := 0
		canReach := true
		for i := 1; i <= 3; i++ {
			sum += dice[0] // hepsi aynı zar (double)
			intermediate := fromPointIndex + direction*sum

			if i < 3 && !canMoveToPoint(stones, player, intermediate) {
				canReach = false
				break
			}
		}
		if canReach && (fromPointIndex+direction*sum) == toPointIndex && canMoveToPoint(stones, player, toPointIndex) {
			usedDie = []int{dice[0], dice[1], dice[2]} // 3 zar kullanıldı
			canMove = true
		}
	}

	// *** Double geldiği zaman 2 zarla hareket kontrolü buraya eklendi ***
	if !canMove && len(dice) == 4 {
		// 2 zar kullanarak hedefe varma kontrolü (örn: ilk 2 zarla)
		sum := 0
		canReach := true
		for i := 1; i <= 2; i++ {
			sum += dice[0] // hepsi aynı zar (double) olduğu için dice[0]
			intermediate := fromPointIndex + direction*sum

			if i < 2 && !canMoveToPoint(stones, player, intermediate) { // ara noktalar engelli mi?
				canReach = false
				break
			}
		}
		if canReach && (fromPointIndex+direction*sum) == toPointIndex && canMoveToPoint(stones, player, toPointIndex) {
			usedDie = []int{dice[0], dice[1]} // ilk 2 zar kullanıldı
			canMove = true
		}
	}

	// *** Double geldiği zaman 3 zarla hareket kontrolü buraya eklendi ***
	if !canMove && len(dice) == 4 {
		// 3 zar kullanarak hedefe varma kontrolü (örn: ilk 3 zarla)
		sum := 0
		canReach := true
		for i := 1; i <= 3; i++ {
			sum += dice[0] // hepsi aynı zar (double) olduğu için dice[0]
			intermediate := fromPointIndex + direction*sum

			if i < 3 && !canMoveToPoint(stones, player, intermediate) { // ara noktalar engelli mi?
				canReach = false
				break
			}
		}
		if canReach && (fromPointIndex+direction*sum) == toPointIndex && canMoveToPoint(stones, player, toPointIndex) {
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
			if i < 4 && !canMoveToPoint(stones, player, intermediate) {
				break
			}

			// Dördüncü adımda hedefe ulaşılmış ve taş konulabilir mi?
			if i == 4 && intermediate == toPointIndex && canMoveToPoint(stones, player, toPointIndex) {
				usedDie = []int{dice[0], dice[1], dice[2], dice[3]} // ya da 4x dice[0]
				canMove = true
			}
		}
	}

	//Kullanilip geride kalan kullanilmamis zarlar burada tanimlanir.
	if canMove && len(usedDie) > 0 {
		result.RemainingDice = calculateRemainingDice(dice, usedDie)
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

// Sadece verilen noktaların taşlarını günceller. Hem PointIndex hem de StackIndex guncellenir. Eski yeri ve gittigi yer, her ikisi de guncellenir.
// Tas kiriliyor ise playerFilter gelir. Kirilan tas kimin filter'i. Geri kalan guncellemelerde Player filitresine ihtiyac yok.
func updateStacks(stones []*LogicalCoordinate, pointsToUpdate []int, playerFilter ...int) []*LogicalCoordinate {
	//Barda kimin tasi kirilmis..
	//Bu sadece kirik taslarin guncellenmesinde player bazli filter eklemek icin kullanilir. Diger Pointindexlerde tek player olacagi icin kullanilmaz..
	var filterByPlayer bool
	var player int
	if len(playerFilter) > 0 {
		filterByPlayer = true
		player = playerFilter[0]
	}

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
				//Sadece kendisinin kirik taslarinin StackIndexini guncelleyecek..
				if filterByPlayer && stone.Player != player {
					continue
				}
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
func calculateRemainingDice(dice []int, used []int) []int {
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
func GetPossibleMovePoints_old(
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

	if !playerHasTopStoneAt(stones, player, fromPointIndex) {
		return possiblePoints
	}

	for _, die := range remainingDice {
		toPointIndex := fromPointIndex + direction*die

		if toPointIndex < 0 || toPointIndex >= 24 {
			continue
		}

		if canMoveToPoint(stones, player, toPointIndex) {
			possiblePoints = append(possiblePoints, toPointIndex)
		}
	}

	return possiblePoints
}

// Oynanabilir zarlara gore belirlenen PointIndex'deki tas ile gidilebilecek PointIndexler yani hamleler hesaplanir.
func GetPossibleMovePoints_Old2(
	stones []*LogicalCoordinate,
	player int,
	fromPointIndex int,
	dice []int,
) []int {
	direction := 1
	if player == 2 {
		direction = -1
	}

	resultSet := make(map[int]bool)

	if !playerHasTopStoneAt(stones, player, fromPointIndex) {
		return nil
	}

	// Tek zarla
	for _, d := range dice {
		to := fromPointIndex + direction*d
		if to >= 0 && to < 24 && canMoveToPoint(stones, player, to) {
			resultSet[to] = true
		}
	}

	// İki farklı zarla (normal zar)
	if len(dice) == 2 {
		d1, d2 := dice[0], dice[1]

		// d1 sonra d2
		intermediate1 := fromPointIndex + direction*d1
		final1 := intermediate1 + direction*d2
		if canMoveToPoint(stones, player, intermediate1) &&
			final1 >= 0 && final1 < 24 &&
			canMoveToPoint(stones, player, final1) {
			resultSet[final1] = true
		}

		// d2 sonra d1
		intermediate2 := fromPointIndex + direction*d2
		final2 := intermediate2 + direction*d1
		if canMoveToPoint(stones, player, intermediate2) &&
			final2 >= 0 && final2 < 24 &&
			canMoveToPoint(stones, player, final2) {
			resultSet[final2] = true
		}
	}

	// Double zar: 3 adım (d, d, d)
	if len(dice) == 4 {
		d := dice[0]
		sum := 0
		valid := true
		for i := 1; i <= 3; i++ {
			sum += d
			intermediate := fromPointIndex + direction*sum
			if i < 3 && !canMoveToPoint(stones, player, intermediate) {
				valid = false
				break
			}
		}
		final := fromPointIndex + direction*sum
		if valid && final >= 0 && final < 24 && canMoveToPoint(stones, player, final) {
			resultSet[final] = true
		}
	}

	// Double zar: 4 adım
	if len(dice) == 4 {
		d := dice[0]
		sum := 0
		valid := true
		for i := 1; i <= 4; i++ {
			sum += d
			intermediate := fromPointIndex + direction*sum
			if i < 4 && !canMoveToPoint(stones, player, intermediate) {
				valid = false
				break
			}
		}
		final := fromPointIndex + direction*sum
		if valid && final >= 0 && final < 24 && canMoveToPoint(stones, player, final) {
			resultSet[final] = true
		}
	}

	// Set → slice
	var possiblePoints []int
	for p := range resultSet {
		possiblePoints = append(possiblePoints, p)
	}

	sort.Ints(possiblePoints)
	return possiblePoints
}
func GetPossibleMovePoints_NotSupport24BearOff(
	stones []*LogicalCoordinate,
	player int,
	fromPointIndex int,
	dice []int,
) []int {
	direction := 1
	if player == 2 {
		direction = -1
	}

	resultSet := make(map[int]bool)

	if !playerHasTopStoneAt(stones, player, fromPointIndex) {
		return nil
	}

	// Tek zarla
	for _, d := range dice {
		to := fromPointIndex + direction*d
		if to >= 0 && to < 24 && canMoveToPoint(stones, player, to) {
			resultSet[to] = true
		}
	}

	// İki farklı zarla (normal zar)
	if len(dice) == 2 {
		d1, d2 := dice[0], dice[1]

		// d1 sonra d2
		intermediate1 := fromPointIndex + direction*d1
		final1 := intermediate1 + direction*d2
		if intermediate1 >= 0 && intermediate1 < 24 && canMoveToPoint(stones, player, intermediate1) &&
			final1 >= 0 && final1 < 24 && canMoveToPoint(stones, player, final1) {
			resultSet[final1] = true
		}

		// d2 sonra d1
		intermediate2 := fromPointIndex + direction*d2
		final2 := intermediate2 + direction*d1
		if intermediate2 >= 0 && intermediate2 < 24 && canMoveToPoint(stones, player, intermediate2) &&
			final2 >= 0 && final2 < 24 && canMoveToPoint(stones, player, final2) {
			resultSet[final2] = true
		}
	}

	// Double zar: 1 ila 4 adım (d, d, d, d)
	if len(dice) == 4 && dice[0] == dice[1] && dice[1] == dice[2] && dice[2] == dice[3] {
		d := dice[0]
		for steps := 1; steps <= 4; steps++ {
			valid := true
			for i := 1; i < steps; i++ {
				intermediate := fromPointIndex + direction*d*i
				if intermediate < 0 || intermediate >= 24 || !canMoveToPoint(stones, player, intermediate) {
					valid = false
					break
				}
			}
			target := fromPointIndex + direction*d*steps
			if valid && target >= 0 && target < 24 && canMoveToPoint(stones, player, target) {
				resultSet[target] = true
			}
		}
	}

	// Set → slice
	var possiblePoints []int
	for p := range resultSet {
		possiblePoints = append(possiblePoints, p)
	}

	//Siraya dizip oyle versin
	sort.Ints(possiblePoints)
	return possiblePoints
}

// Bir tas secilince PointIndex ,Zar ve Player bilgisi ile gidilebilecek PointIndex noktalari verilir. Bir cesit Bot Helper..
// [24 bear off] tas playera gore toplama alani icinde ise 24 PointIndex sunulur ama tas disardan geliyor ise 24 PointIndex sunulmaz. Disardan dogrudan gelen tasin Bear Off olmasini bu functionda desteklenmiyor.
func GetPossibleMovePoints(
	stones []*LogicalCoordinate,
	player int,
	fromPointIndex int,
	dice []int,
) []int {
	direction := 1
	if player == 2 {
		direction = -1
	}

	resultSet := make(map[int]bool)

	//Kirik Tasi var ve baska tasi oynamak istiyor..Izin verilmez..
	isPlayerMustEnterFromBar := PlayerMustEnterFromBar(stones, player)
	if isPlayerMustEnterFromBar && fromPointIndex != -1 {
		return nil // bar'dan inmediği sürece başka taş oynayamaz
	} else if isPlayerMustEnterFromBar && fromPointIndex == -1 { //Kirik tasi sokmasi durumunda girilebilecek PointIndexleri Gosterir..
		barResult := IsAnyBarEntryAllowed(stones, player, dice) //1-) barResult.EnterableDice kullanilabilecek zarlari ifade eder.
		if !barResult.Allowed {
			return nil
		}
		for _, die := range barResult.EnterableDice {
			entryPoint := GetEntryPoint(player, die)
			resultSet[entryPoint] = true
		}

		//Eger kirik tas girilmeye calisiliyor ise diger kosullara bakilmasina gerek yok..
		// Set → slice
		return mapKeysToSortedSlice(resultSet)

		/*	var possiblePoints []int
			for p := range resultSet {
				possiblePoints = append(possiblePoints, p)
			}

			sort.Ints(possiblePoints)
			return possiblePoints*/
	}

	if !playerHasTopStoneAt(stones, player, fromPointIndex) {
		return nil
	}

	/*isBearOffTarget := func(to int) bool {
		return to >= 24
	}*/
	//Tas toplanirken Player yonune gore Player 1 > 24 ve Player 2 < 0 "to" degerine bakilir ama toplanan her tasin PointIndex'i 24 dur!
	isBearOffTarget := func(to int) bool {
		if player == 1 {
			return to >= 24
		} else if player == 2 {
			return to < 0
		}
		return false
	}

	// Tek zarla
	for _, d := range dice {
		to := fromPointIndex + direction*d
		if to >= 0 && to < 24 {
			if canMoveToPoint(stones, player, to) {
				resultSet[to] = true
			}
		} else if isBearOffTarget(to) {
			if allowed, _, _ := CanBearOffStone(stones, player, fromPointIndex, []int{d}); allowed {
				resultSet[24] = true
			}
		}
	}

	// İki farklı zarla (normal zar)
	if len(dice) == 2 {
		d1, d2 := dice[0], dice[1]

		// d1 sonra d2
		intermediate1 := fromPointIndex + direction*d1
		final1 := intermediate1 + direction*d2
		if intermediate1 >= 0 && intermediate1 < 24 && canMoveToPoint(stones, player, intermediate1) {
			if final1 >= 0 && final1 < 24 && canMoveToPoint(stones, player, final1) {
				resultSet[final1] = true
			} else if isBearOffTarget(final1) {
				if allowed, _, _ := CanBearOffStone(stones, player, fromPointIndex, []int{d1, d2}); allowed {
					resultSet[24] = true
				}
			}
		}

		// d2 sonra d1
		intermediate2 := fromPointIndex + direction*d2
		final2 := intermediate2 + direction*d1
		if intermediate2 >= 0 && intermediate2 < 24 && canMoveToPoint(stones, player, intermediate2) {
			if final2 >= 0 && final2 < 24 && canMoveToPoint(stones, player, final2) {
				resultSet[final2] = true
			} else if isBearOffTarget(final2) {
				if allowed, _, _ := CanBearOffStone(stones, player, fromPointIndex, []int{d2, d1}); allowed {
					resultSet[24] = true
				}
			}
		}
	}

	// Double zar: 3 adım. 1'i kullanilmis. (d, d, d) 1'den 3 adıma kadar
	if len(dice) == 3 && dice[0] == dice[1] {
		d := dice[0]
		for steps := 1; steps <= 3; steps++ {
			sum := d * steps
			valid := true
			//1, 2 ve 3 adimlik tum kademeler denenir.
			for i := 1; i < steps; i++ {
				intermediate := fromPointIndex + direction*d*i
				if intermediate < 0 || intermediate >= 24 || !canMoveToPoint(stones, player, intermediate) {
					valid = false
					break
				}
			}
			target := fromPointIndex + direction*sum
			if valid {
				if target >= 0 && target < 24 && canMoveToPoint(stones, player, target) {
					resultSet[target] = true
				} else if isBearOffTarget(target) {
					if allowed, _, _ := CanBearOffStone(stones, player, fromPointIndex, dice[:steps]); allowed {
						resultSet[24] = true
					}
				}
			}
		}
	}

	// Double zar: 4 adım (d, d, d, d) 1'den 4 adıma kadar
	if len(dice) == 4 && dice[0] == dice[1] {
		d := dice[0]
		for steps := 1; steps <= 4; steps++ {
			sum := d * steps
			valid := true
			//1, 2, 3 ve 4 adimlik tum kademeler denenir.
			for i := 1; i < steps; i++ {
				intermediate := fromPointIndex + direction*d*i
				if intermediate < 0 || intermediate >= 24 || !canMoveToPoint(stones, player, intermediate) {
					valid = false
					break
				}
			}
			target := fromPointIndex + direction*sum
			if valid {
				if target >= 0 && target < 24 && canMoveToPoint(stones, player, target) {
					resultSet[target] = true
				} else if isBearOffTarget(target) {
					if allowed, _, _ := CanBearOffStone(stones, player, fromPointIndex, dice[:steps]); allowed {
						resultSet[24] = true
					}
				}
			}
		}
	}
	/*if len(dice) == 4 {
		d := dice[0]
		sum := 0
		valid := true
		for i := 1; i <= 4; i++ {
			sum += d
			intermediate := fromPointIndex + direction*sum
			if i < 4 && !canMoveToPoint(stones, player, intermediate) {
				valid = false
				break
			}
		}
		final := fromPointIndex + direction*sum
		if valid {
			if final >= 0 && final < 24 && canMoveToPoint(stones, player, final) {
				resultSet[final] = true
			} else if isBearOffTarget(final) {
				if allowed, _, _ := CanBearOffStone(stones, player, fromPointIndex, []int{d, d, d, d}); allowed {
					resultSet[24] = true
				}
			}
		}
	}*/

	// Set → slice
	return mapKeysToSortedSlice(resultSet)

	/*var possiblePoints []int
	for p := range resultSet {
		possiblePoints = append(possiblePoints, p)
	}

	sort.Ints(possiblePoints)
	return possiblePoints*/
}

func mapKeysToSortedSlice(m map[int]bool) []int {
	var result []int
	for k := range m {
		result = append(result, k)
	}
	sort.Ints(result)
	return result
}

// Oyuncuya gore Tas Toplama PointIndex araliklari belirleniyor.
func getBearoffRangeForPlayer(player int) []int {
	if player == 1 {
		return []int{18, 19, 20, 21, 22, 23}
	}
	return []int{0, 1, 2, 3, 4, 5}
}

// Toplamak icin(Bear_Off) tum taslar olmasi gereken yerde mi ?
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
// [24 bear off] tas playera gore toplama alani icinde olmadan olmuyor. Disardan dogrudan gelen tasin Bear Off olmasini bu function desteklenmiyor.
// Tas Toplamaya Uygun mu, kullanilan zar(usedDice) ve geriye kullanilabilecek zar(remainingDice) da donulur.
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
	if !playerHasTopStoneAt(stones, player, pointIndex) {
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
			// Daha geride taş var mı kontrol et varsa o pointIndexdeki tasi alamassin => return false
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

// Kazanan Player'i belirlemek icin kullanilir.
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

/*TryMoveStone, bir oyuncunun taşını belirtilen noktadan başka bir noktaya taşımayı dener.
Bu fonksiyon, verilen zar(lar) ve mevcut taş durumu üzerinden geçerli bir hamle olup olmadığını kontrol eder ve uygular.
Yani once PointIndex'ini degistirir. Sonra gitti yiginda StackIndex'ini degistirip isTop= true yapar ve
sort isleminde en tepede ciksin diye global artan MoveOrder degeri kullanilir.

Parametreler:
  - stones: Oyun tahtasındaki mevcut tüm taşların listesi.
  - player: Hamleyi yapan oyuncunun numarası (1 veya 2).
  - fromPoint: Taşın bulunduğu noktanın indeksidir. -1 ise oyuncunun bar’daki (kırık) taşını içeri soktuğu anlamına gelir.
  - toPoint: Taşın gitmek istediği hedef noktanın indeksidir. 24 ise taş toplanmak (bearing off / Africa) isteniyor demektir.
  - dice: Oyuncunun o turda sahip olduğu kullanılabilir zarlar.

Geri Dönenler:
  - newStones: Taşların güncellenmiş hali (hamle başarılıysa).
  - ok: Hamlenin kurallara uygun olup olmadığını belirten bool değer.
  - usedDice: Hamlede kullanılan zar(lar).
  - remainingDice: Kullanılmayan zarlar (hamleden sonra kalır).
  - brokenStones: Eğer rakip taş kırıldıysa, kırılan taş(lar)ın orijinal (eski) hali.

Açıklama:
  - Bar’dan giriş: Eğer fromPoint == -1 ise, oyuncunun bar’daki taşını zar değerine göre uygun giriş noktasına sokulup sokamayacağı kontrol edilir.
    Rakibin taşı varsa ve tekse kırılır, bar’a gönderilir.
  - Normal hamle: Oyuncunun tahtadaki taşlarıyla, zar değeri kadar ileriye hamle yapması kontrol edilir. Rakip taş varsa kırma veya blokaj kontrolü yapılır.
  - BearOff (taş toplama): Eğer toPoint == 24 ise, taş toplanmak isteniyordur. Oyuncunun tüm taşları son bölgede mi ve zar uygun mu diye kontrol edilir.

Kurallar:
  - Oyuncunun bar’da kırık taşı varsa, önce onu içeri sokması gerekir (öncelikli hamledir).
  - Hedef noktada 1 adet rakip taşı varsa, kırma işlemi yapılır ve rakip taş bar’a gönderilir.
  - Bar’a gönderilen taşlar ilgili oyuncuya göre güncellenir ve bar'daki yığılma (StackIndex) sırası yeniden hesaplanır.

Not:
  - Bu fonksiyon yalnızca taşların hareketini ve geçerliliğini yönetir. Görsellik veya animasyon içermez.
  - Gerekli kuralların kontrolü için `MoveTopStoneAndUpdate`, `IsAnyBarEntryAllowed`, `calculateRemainingDice` gibi yardımcı fonksiyonlar içeride çağrılır.*/

// Son Kullanicinin hersey icin kullanacagi Function => Test : TestFullSmilation()
func TryMoveStone(
	stones []*LogicalCoordinate,
	player int,
	fromPoint int,
	toPoint int,
	dice []int,
) (newStones []*LogicalCoordinate, ok bool, usedDice []int, remainingDice []int, brokenStones []*LogicalCoordinate) {

	// 1. Bar'dan giriş durumu. Kirik tas var mi ?
	if fromPoint == -1 {
		barResult := IsAnyBarEntryAllowed(stones, player, dice) //1-) barResult.RemainingDice kullanilabilecek zarlardan sonra geri kalan kirigi sokamayacagin zarlari ifade eder.
		if !barResult.FromBar || !barResult.Allowed {
			return stones, false, dice, nil, nil
		}

		var usedDie int = -1
		// Giriş yapılabilecek zar var mı? toPoint'i kontrol etmiyoruz, zarla gelen yere giriyoruz
		for _, die := range barResult.EnterableDice {
			entryPoint := GetEntryPoint(player, die)
			if entryPoint == toPoint && canMoveToPoint(stones, player, toPoint) {
				usedDie = die
				break
			}
		}

		if usedDie == -1 {
			// toPoint için uygun zar yok
			return stones, false, dice, nil, nil
		}

		newStones, moved, brokenStones := MoveTopStoneAndUpdate(stones, player, -1, toPoint)
		if !moved {
			return stones, false, dice, nil, nil
		}

		used := []int{usedDie}
		remainingDice := calculateRemainingDice(dice, used)
		return newStones, true, used, remainingDice, brokenStones
	}

	// 2. Hedef toplama alanı mı?
	isBearOff := func(to int) bool {
		/*if player == 1 {
			return to >= 24
		} else if player == 2 {
			return to < 0
		}
		return false*/
		return to >= 24
	}

	// 3. Oyuncunun fromPoint noktasında en üst taşı var mı?
	if !playerHasTopStoneAt(stones, player, fromPoint) {
		return stones, false, dice, nil, nil
	}

	// 4. Bear-off kontrolü
	if isBearOff(toPoint) {
		canBearOff, remaining, used := CanBearOffStone(stones, player, fromPoint, dice)
		if !canBearOff {
			return stones, false, dice, nil, nil
		}
		newStones, ok, brokenStones := MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
		return newStones, ok, used, remaining, brokenStones
	}

	// 5. Normal taş hareketi
	moveResult := IsNormalMoveAllowed(stones, player, fromPoint, toPoint, dice)
	if !moveResult.Allowed {
		return stones, false, dice, nil, nil
	}

	newStones, ok, brokenStones = MoveTopStoneAndUpdate(stones, player, fromPoint, toPoint)
	return newStones, ok, moveResult.UsedDice, moveResult.RemainingDice, brokenStones
}

// Player Bazinda toplanan toplam tas(Bear_Off) sayisini verir..
func CountCollectedStones(stones []*LogicalCoordinate, player int) int {
	count := 0
	for _, s := range stones {
		if s.Player == player && s.PointIndex == 24 {
			count++
		}
	}
	return count
}

/*func updateStacksForPlayerBar(stones []*LogicalCoordinate, player int) []*LogicalCoordinate {
	// Bar noktası: -1
	barIndex := -1

	// Bar'daki sadece ilgili oyuncunun taşlarını filtrele
	group := []*LogicalCoordinate{}
	for _, stone := range stones {
		if stone.PointIndex == barIndex && stone.Player == player {
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

	return stones
}*/
