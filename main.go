package main

import (
	"fmt"
	"log"
	"os"
)

func LoadStrategy(name string) (*Choice, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s, err := LoadSelbyStrategy(f)
	return s, err
}

type wordResult struct {
	W string
	R Result
}

func ValidGuessNormal(g string, prev []wordResult) bool {
	return WordDict[g]
}

// ValidGuessHard checks if guess g is compatible with
// previous guesses.
// Any green letters previously found must be in the
// same place in the guess, and any yellow letters
// must appear somewhere in the guess.
func ValidGuessHard(g string, prev []wordResult) bool {
	if !WordDict[g] {
		return false
	}
	for _, p := range prev {
		var used [5]bool
		for i, lc := range p.R.Letters() {
			if lc == Green {
				if g[i] != p.W[i] {
					return false
				}
				used[i] = true
			}
		}
		for i, lc := range p.R.Letters() {
			if lc == Yellow {
				found := false
				for j := 0; j < 5; j++ {
					if used[j] {
						continue
					}
					if g[j] == p.W[i] {
						used[j] = true
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}
	return true
}

func Play(c *Choice, target string, valid func(string, []wordResult) bool) (int, error) {
	var prev []wordResult
	var prevWords []string
	for n := 1; ; n++ {
		g := c.Guess(prevWords)
		if !valid(g, prev) {
			return 0, fmt.Errorf("bad guess %q when previous guesses are %v", g, prev)
		}
		sc := Score(target, g)
		if sc == AllGreen {
			return n, nil
		}
		var err error
		c, err = c.Next(sc)
		if err != nil {
			return 0, err
		}
		prev = append(prev, wordResult{g, sc})
		prevWords = append(prevWords, g)
	}
}

func main() {
	for hard := 0; hard <= 2; hard++ {
		validf := ValidGuessNormal
		if hard >= 1 {
			validf = ValidGuessHard
		}
		harddesc := []string{"normal", "hard", "hard(max5)"}[hard]
		s, err := LoadStrategy([]string{"strategy_normal.txt", "strategy_hard.txt", "strategy_hard5.txt"}[hard])
		if err != nil {
			log.Fatal(err)
		}
		sum := 0
		for _, w := range Answers {
			score, err := Play(s, w, validf)
			if err != nil {
				log.Fatalf("bad play when guessing %q: %v", w, err)
			}
			sum += score
		}
		mean := float64(sum) / float64(len(Answers))
		fmt.Printf("%-11s sum=%d mean=%.4f\n", harddesc+":", sum, mean)
	}
}
