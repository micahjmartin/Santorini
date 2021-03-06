package main

import (
	"fmt"
	"os"
	"santorini/bots"
	santorini "santorini/pkg"
	"strconv"
	"sync"

	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

var knownbots = []santorini.BotInitializer{bots.NewBasicBot, bots.NewKyleBot, bots.NewRandomBot}

func listBots() {
	for i, b := range knownbots {
		bot := b(0, &santorini.Board{}, nil)
		fmt.Printf("%d. %s\n", i, bot.Name())
	}
}

type options struct {
	threadCount int `help:"Number of threads to use"`
	simCount    int
}

type overallstats struct {
	bot1Wins int
	bot2Wins int
	// Calculate average round count
	sumRounds  int
	loseBoards []*santorini.Board
	pb         *progressbar.ProgressBar
}

func (stats *overallstats) update(sim *santorini.Simulation) {
	// The first bot can be team 1 or team 2 depending on the round number
	if sim.Board.Victor == sim.Number%2+1 {
		stats.bot1Wins++
	} else {
		stats.bot2Wins++
		// Keep track of the losses
		stats.loseBoards = append(stats.loseBoards, sim.Board)
	}
	stats.sumRounds += len(sim.Board.Moves) / 2
	if stats.pb != nil {
		stats.pb.Describe(fmt.Sprintf("%03d / %03d", stats.bot1Wins, stats.bot2Wins))
		stats.pb.Add(1)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Chose two bots to simulate. Bots will alternate going first. Deterministic bots will only run 1 game each.")
		fmt.Printf("USAGE: %s bot1 bot2 [numRounds]\n", os.Args[0])
		listBots()
		os.Exit(1)
	}

	var bot1 santorini.BotInitializer
	var bot2 santorini.BotInitializer
	for _, b := range knownbots {
		bot := b(0, &santorini.Board{}, nil)
		if bot.Name() == os.Args[1] {
			bot1 = b
		}
		if bot.Name() == os.Args[2] {
			bot2 = b
		}
	}

	if bot1 == nil {
		fmt.Printf("%s is not a known bot\n", os.Args[1])
		os.Exit(1)
	}
	if bot2 == nil {
		fmt.Printf("%s is not a known bot\n", os.Args[2])
		os.Exit(1)
	}

	opts := &options{
		threadCount: 10,
		simCount:    1000,
	}

	//logrus.SetLevel(logrus.DebugLevel)
	// Deterministic bots dont need to be run many times (unless explicitly told to)
	b1 := bot1(0, &santorini.Board{}, nil)
	b2 := bot2(0, &santorini.Board{}, nil)

	if b1.IsDeterministic() && b2.IsDeterministic() {
		opts.simCount = 2
	}
	if len(os.Args) > 3 {
		if i, err := strconv.ParseInt(os.Args[3], 10, 64); err == nil {
			opts.simCount = int(i)
		} else {
			fmt.Println("Cannot parse integer", os.Args[3])
			os.Exit(1)
		}
	}

	logrus.Infof("Running %d simulations between %s and %s", opts.simCount, b1.Name(), b2.Name())
	stats := &overallstats{
		loseBoards: make([]*santorini.Board, 0, opts.simCount),
		pb:         progressbar.Default(int64(opts.simCount), "0 / 0"),
	}

	wg := new(sync.WaitGroup)
	wg2 := new(sync.WaitGroup)
	sims := make(chan *santorini.Simulation)
	completedSims := make(chan *santorini.Simulation)

	logrus.Debugf("Starting %d workers", opts.threadCount)
	for i := 0; i < opts.threadCount; i++ {
		wg.Add(1)
		go runner(wg, sims, completedSims)
	}
	wg2.Add(1)
	go statistician(wg2, completedSims, stats)

	// run all the sim
	for i := 0; i < opts.simCount; i++ {
		var sim *santorini.Simulation
		if i%2 == 0 {
			sim = santorini.NewSimulator(i, logrus.StandardLogger(), bot1, bot2)
		} else {
			sim = santorini.NewSimulator(i, logrus.StandardLogger(), bot2, bot1)
		}
		sims <- sim
	}

	// Wait for all the sims to finish
	logrus.Debug("Waiting for runners to finish")
	close(sims)
	wg.Wait()
	// Wait for the stats to finish
	logrus.Debug("Waiting for stats to finish")
	close(completedSims)
	wg2.Wait()

	logrus.WithFields(map[string]interface{}{
		"bot1":             b1.Name(),
		"bot1_wins":        stats.bot1Wins,
		"bot2":             b2.Name(),
		"bot2_wins":        stats.bot2Wins,
		"avg_round_length": stats.sumRounds / opts.simCount,
		"num_rounds":       opts.simCount,
	}).Info("Simulation Complete")
}

func runner(wg *sync.WaitGroup, sims chan *santorini.Simulation, results chan *santorini.Simulation) {
	defer wg.Done()
	defer logrus.Debug("Runner finished")
	for sim := range sims {
		sim.Run()
		results <- sim
	}
}

func statistician(wg *sync.WaitGroup, results chan *santorini.Simulation, stats *overallstats) {
	defer wg.Done()
	for sim := range results {
		stats.update(sim)
	}
}
