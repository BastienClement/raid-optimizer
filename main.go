package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/MaxHalford/eaopt"
)

const RMAX = 10
const CMAX = 100

var strategy Strategy
var healerMinRatio, healerMaxRatio float64
var raidSize int

var roster []Character
var players []string

var playerCharacters [][]int

type RoleIndex struct {
	Chars []int
	Mains []int
	Alts  []int
}

var roleIndex = struct {
	Tank RoleIndex
	Heal RoleIndex
	Dps  RoleIndex
}{}

var minRaids, maxRaids int
var checkViability bool

var prepareFns []func()

func IndexRoster() {
	playerCharacters = make([][]int, len(players))
	for cid, char := range roster {
		player := char.Player
		playerCharacters[player] = append(playerCharacters[player], cid)

		var ridx *RoleIndex
		switch char.Role {
		case Tank:
			ridx = &roleIndex.Tank
		case Healer:
			ridx = &roleIndex.Heal
		case Melee, Ranged:
			ridx = &roleIndex.Dps
		}

		ridx.Chars = append(ridx.Chars, cid)
		if char.Main {
			ridx.Mains = append(ridx.Mains, cid)
		} else {
			ridx.Alts = append(ridx.Alts, cid)
		}
	}
}

func ComputeBounds() {
	// Tanks-related bounds
	minRaids = Max(minRaids, len(roleIndex.Tank.Mains)/2) // Using only mains
	maxRaids = Min(maxRaids, len(roleIndex.Tank.Chars)/2) // Using every tanks

	mainCount := float64(len(roleIndex.Tank.Mains) + len(roleIndex.Heal.Mains) + len(roleIndex.Dps.Mains))
	charCount := float64(len(roleIndex.Tank.Chars) + len(roleIndex.Heal.Chars) + len(roleIndex.Dps.Chars))

	// Roster-related bounds
	minRaids = Max(minRaids, int(math.Ceil(mainCount/float64(raidSize)))) // Packing mains in the minimum number of raids
	maxRaids = Min(maxRaids, int(math.Ceil(charCount/10)))                // Spreading every char in the smallest possible raids

	// Healers-related bounds
	// TODO: ensure bounds are not broken when taking healer ratio in consideration

	log.Printf("Raid count bounds: %d-%d", minRaids, maxRaids)
}

var cpuprofile *string

func ParseOpts(ga *eaopt.GA) {
	optStrategy := flag.String("strategy", "armor", "optimization strategy")

	flag.IntVar(&raidSize, "size", 30, "maximum raid size")
	flag.IntVar(&minRaids, "min", 2, "minimum number of raids")
	flag.IntVar(&maxRaids, "max", RMAX, "maximum number of raids")

	flag.Float64Var(&healerMinRatio, "healer-min", 0.175, "minimum ratio of healer in raid")
	flag.Float64Var(&healerMaxRatio, "healer-max", 0.5, "maximum ratio of healer in raid")

	flag.UintVar(&ga.NPops, "npops", 12, "number of populations")
	flag.UintVar(&ga.PopSize, "popsize", 3000, "number of size of populations")
	flag.UintVar(&ga.NGenerations, "gen", 2000, "number of generation")

	model := flag.String("model", "default", "the EA model to use")
	noCheck := flag.Bool("no-check", false, "check raid viability at each steps")

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	strategy = ParseStrategy(*optStrategy)
	checkViability = !*noCheck

	switch *model {
	case "mutonly":
		ga.Model = eaopt.ModMutationOnly{Strict: true}
	case "mutonly-nonstrict":
		ga.Model = eaopt.ModMutationOnly{Strict: false}
	case "default":
		ga.Model = eaopt.ModGenerational{Selector: eaopt.SelTournament{NContestants: 3}, MutRate: 1, CrossRate: 1}
	default:
		log.Fatalf("Unknown model: %s", *model)
	}
}

func main() {
	log.SetFlags(0)
	ga, err := eaopt.NewDefaultGAConfig().NewGA()
	if err != nil {
		log.Fatal(err)
	}

	ParseOpts(ga)
	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	log.Printf("Using strategy: %s", strategy)
	log.Printf("Checking viability: %v\n", checkViability)
	log.Printf("Model: %+v\n\n", ga.Model)

	log.Printf("Loading roster...")
	roster, players = LoadRoster()

	log.Printf("Indexing roster...")
	IndexRoster()
	ComputeBounds()
	fmt.Fprint(os.Stderr, "\n")

	strategy.Prepare()
	for _, fn := range prepareFns {
		fn()
	}

	var nextPercent uint = 0
	var stop bool = false
	var start int64

	ga.Callback = func(ga *eaopt.GA) {
		percent := ga.Generations * 100 / ga.NGenerations
		if percent >= nextPercent {
			nextPercent = percent + 5
			var eta string
			if start > 0 && percent < 100 {
				d := time.Now().UnixMilli() - start
				eta = fmt.Sprintf("  eta=%ds", (time.Duration(int64(100-percent)*d/int64(percent))*time.Millisecond).Truncate(time.Second)/time.Second)
			} else {
				start = time.Now().UnixMilli()
			}

			log.Printf("Best fitness after %3d%%: %f%s", percent, ga.HallOfFame[0].Fitness, eta)
		}
	}

	ga.EarlyStop = func(ga *eaopt.GA) bool {
		return stop
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for {
			<-c
			if stop {
				os.Exit(1)
			}
			stop = true
		}
		//log.Printf("Interrupted")
	}()

	ga.MigFrequency = ga.NGenerations / 5
	ga.Migrator = eaopt.MigRing{NMigrants: ga.PopSize / 4}
	ga.HofSize = 1

	if minRaids != maxRaids {
		ga.Speciator = Speciator{}
	}

	//rng.Seed(time.Now().UnixNano())
	//ga.RNG = rand.New(&rng)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	fmt.Fprint(os.Stderr, "\n")
	log.Printf("Starting...")
	err = ga.Minimize(func(rng *rand.Rand) eaopt.Genome {
		return MakeRaid(rng)
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	for i := 0; i < int(ga.HofSize); i++ {
		PrintRaid(ga.HallOfFame[i].Genome.(*Genome))
	}
}
