package main

import (
	"bufio"
	"fmt"
	"os"
	"santorini/bots"
	santorini "santorini/pkg"
	"santorini/pkg/color"
	"santorini/ui"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type TurnSelector interface {
	SelectTurn() santorini.Turn
}

func main() {
	game := ui.NewGame(1, bots.NewRandomBot)
	game.Run()

}
func main2() {
	// Initialize a new board
	board := santorini.NewBoard()

	// Place Workers
	board.PlaceWorker(1, 1, 2, 1)
	board.PlaceWorker(1, 2, 2, 3)
	board.PlaceWorker(2, 1, 1, 2)
	board.PlaceWorker(2, 2, 3, 2)

	// Initialize RNG Team 1
	team1 := bots.NewBasicBot(1, board, logrus.StandardLogger())
	// Initialize Team 2
	team2 := bots.NewPlayerBot(2, board, logrus.StandardLogger())

	fmt.Println("Team 1 -", team1.Name())
	fmt.Printf("%sTeam 1 - Worker 1%s\n", color.GetWorkerColor(1, 1), color.Reset)
	fmt.Printf("%sTeam 1 - Worker 2%s\n", color.GetWorkerColor(1, 2), color.Reset)
	fmt.Printf("%sTeam 2 - Worker 1%s\n", color.GetWorkerColor(2, 1), color.Reset)
	fmt.Printf("%sTeam 2 - Worker 2%s\n", color.GetWorkerColor(2, 2), color.Reset)

	// REPL
	reader := bufio.NewReader(os.Stdin)
	for round := 0; round < 1000; round++ {
		fmt.Printf("\nStarting Round %d\n\n", round+1)
		// Print the board
		fmt.Println(board)

		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if text == "exit" {
			return
		}
		if strings.Contains(text, ",") {
			parts := strings.Split(text, ",")
			x, err := strconv.ParseInt(parts[0], 10, 32)
			if err != nil {
				panic(err)
			}
			y, err := strconv.ParseUint(parts[1], 10, 8)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Information for Tile %d,%d:\n", x, y)
			tile := board.GetTile(int(x), int(y))
			fmt.Printf("%+v\n", tile)
		}

		// Team 1 Select
		turn1 := team1.SelectTurn()
		if turn1 == nil {
			fmt.Printf("Team 2 Wins! Team 1 has no remaining moves\n")
			break
		}

		//turn1Data, _ := json.Marshal(turn1)
		//fmt.Printf("Turn JSON: %s\n", turn1Data)
		fmt.Printf("Team 1 moves %sWorker %d%s to %d,%d and builds %d,%d\n",
			color.GetWorkerColor(turn1.Team, turn1.Worker),
			turn1.Worker,
			color.Reset,
			turn1.MoveTo.GetX(),
			turn1.MoveTo.GetY(),
			turn1.Build.GetX(),
			turn1.Build.GetY(),
		)
		if board.PlayTurn(*turn1) {
			fmt.Printf("Team 1 Wins!\n")
			break
		}
		fmt.Printf("\n%s\n\n", board)

		// Team 2 Select
		turn2 := team2.SelectTurn()
		if turn2 == nil {
			fmt.Printf("Team 1 Wins! Team 2 has no remaining moves\n")
			break
		}
		//turn2Data, _ := json.Marshal(turn2)
		//fmt.Printf("Turn JSON: %s\n", turn2Data)
		fmt.Printf("Team 2 moves %sWorker %d%s to %d,%d and builds %d,%d\n",
			color.GetWorkerColor(turn2.Team, turn2.Worker),
			turn2.Worker,
			color.Reset,
			turn2.MoveTo.GetX(),
			turn2.MoveTo.GetY(),
			turn2.Build.GetX(),
			turn2.Build.GetY(),
		)
		if board.PlayTurn(*turn2) {
			fmt.Printf("Team 2 Wins!\n")
			break
		}
		fmt.Printf("\n%s\n\n", board)
	}

	fmt.Printf("Final Board:\n%s\n", board)
}
