package main

import (
	"backgammon/core"
	"fmt"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	core.ResetMoveOrder()
	//Taslari diz
	var stones = core.GetInitialStones()
	// PointIndex'e göre artan şekilde sırala
	core.SortStonesByPlayerPointAndStackDesc(stones)

	// Console'a bas
	fmt.Println("Başlangıç taşları:")
	for _, stone := range stones {
		fmt.Printf("PointIndex: %d, Player: %d, StackIndex: %d, IsTop: %v\n, MoveOrder: %d\n",
			stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
	}

	//Zar At...
	dices, err := core.RollDice()
	if err != nil {
		fmt.Println("Zar atılırken bir hata oluştu:", err)
		return
	}

	var strDice string
	for i, d := range dices {
		if i > 0 {
			strDice += ","
		}
		strDice += fmt.Sprintf("%d", d)
	}

	fmt.Printf("Zarlar: %s\n", strDice)

	//stones, moved := core.MoveStoneAndUpdate(stones, 0, 11, 1)
	stones, moved, brokenStones := core.MoveTopStoneAndUpdate(stones, 1, 0, 11)
	core.SortStonesByPlayerPointAndStackDesc(stones)
	if moved {
		fmt.Println("Taş başarıyla hareket etti.")
		fmt.Println("Taşların Son Durumu:")
		for _, stone := range stones {
			fmt.Printf("PointIndex: %d, Player: %d, StackIndex: %d, IsTop: %v\n, MoveOrder: %d\n",
				stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
		}
		for _, stone := range brokenStones {
			fmt.Printf("Kirilan Tas PointIndex: %d, Player: %d, StackIndex: %d, IsTop: %v\n, MoveOrder: %d\n",
				stone.PointIndex, stone.Player, stone.StackIndex, stone.IsTop, stone.MoveOrder)
		}
	} else {
		fmt.Println("Taş hareket edemedi.")
	}
}
