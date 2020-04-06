package codenames

import (
	"math"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

const (
	Red     = iota // 0
	Blue           // 1
	Neutral        // 2
	Black          // 3

	identifierLength = 5  // length of a game identifier
	numWords         = 25 // number of words in a game

	// dictionary used to create game identifiers
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	// default word assignments, missing extra word for starting team
	defaultAssignments = createWordAssignments()
)

// contains information about a single word on the board
type GameWord struct {
	Word     string `json:"word"`
	Color    int    `json:"color"`
	Revealed bool   `json:"revealed"`
}

// contains information about the entire state of the game
type GameBoard struct {
	Words         [numWords]*GameWord `json:"words"`
	Seed          int64               `json:"seed"` // permanent seed for a given game
	StartingColor int                 `json:"startingColor"`
}

// create a new initialized game board
func CreateGame(words []string, existingIdentifiers map[string]bool) (string, *GameBoard) {
	var identifier string
	for ok := true; ok; _, ok = existingIdentifiers[identifier] {
		identifier = createGameIdentifier()
	}
	log.Infof("created game identifier %s", identifier)

	// generate random seed for this game
	gameSeed := rand.Int63()
	randSeed := rand.New(rand.NewSource(gameSeed))

	// select words as permutation from dictionary
	wordsIndex := randSeed.Perm(len(words))
	gameWords := [numWords]*GameWord{}

	// choose starting color at random, clone default word assignments
	startingColor := randSeed.Intn(2)
	gameAssignments := append(append(defaultAssignments[:0:0], startingColor), defaultAssignments...)

	// shuffle game assignments
	for i := 0; i < 10; i++ {
		randSeed.Shuffle(len(gameAssignments), func(i, j int) {
			gameAssignments[i], gameAssignments[j] = gameAssignments[j], gameAssignments[i]
		})
	}

	// initialize words
	for index, permIndex := range wordsIndex[:numWords] {
		gameWords[index] = &GameWord{
			Word:     words[permIndex],
			Color:    gameAssignments[index],
			Revealed: false,
		}
	}

	// create game board
	return identifier, &GameBoard{
		Words:         gameWords,
		Seed:          gameSeed,
		StartingColor: startingColor,
	}
}

// create default word assignments that will be shuffled per game
func createWordAssignments() []int {
	colorCards := int(math.Floor(numWords / 3))
	var assignments []int

	assignments = append(assignments, duplicateElement(Red, colorCards)...)
	assignments = append(assignments, duplicateElement(Blue, colorCards)...)
	assignments = append(assignments, duplicateElement(Neutral, colorCards-1)...)
	assignments = append(assignments, Black)

	return assignments
}

// create an array with value repeated n times
func duplicateElement(value, n int) []int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = value
	}

	return arr
}

func createGameIdentifier() string {
	b := make([]byte, identifierLength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
