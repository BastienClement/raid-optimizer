package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode/utf8"
)

var longestCharName int

type Character struct {
	Player     int
	Name       string
	Class      Class
	Role       Role
	Main       bool
	TokenSlots TokenSlotSet
}

func (char Character) String() string {
	if char.Class < Warrior {
		return fmt.Sprintf("%-"+strconv.Itoa(longestCharName)+"s", "")
	}

	r, g, b := ClassColor(char.Class)
	var format string
	if char.Main {
		format = fmt.Sprintf("\x1b[48;2;%d;%d;%dm\x1b[38;5;0m", r, g, b)
	} else {
		format = fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
	}

	return fmt.Sprintf("%s%-"+strconv.Itoa(longestCharName)+"s\x1b[0m", format, char.Name)
}

func LoadRoster() ([]Character, []string) {
	f, err := os.Open(flag.Arg(0))
	if err != nil  {
		log.Fatalf("%s", err)
	}
	reader := csv.NewReader(f)

	records, err := reader.ReadAll()
	if err != nil  {
		log.Fatalf("%s", err)
	}

	if len(records) > CMAX {
		log.Fatalf("Number of record (%d) > CMAX (%d)", len(records), CMAX)
	}

	roster := make([]Character, len(records))
	players := make([]string, 0)
	playerIndex := make(map[string]int)
		
	for i, record := range records {
		player := record[0]
		if _, found := playerIndex[player]; !found {
			playerIndex[player] = len(players)
			players = append(players, player)
		}

		name := record[2]
		if len := utf8.RuneCountInString(name) + 1; len > longestCharName {
			longestCharName = len
		}

		char := Character{
			Player: playerIndex[player],
			Name:   name,
			Class:  ParseClass(record[3]),
			Role:   ParseRole(record[1]),
			Main:   record[4] == "True",
		}


		strategy.LoadChar(&char, record)
		roster[i] = char
	}

	return roster, players
}
