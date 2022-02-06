package main

import (
	"fmt"
	"log"
	"os"
	"sort"
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

func checkStrats() {
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

func sim(a, b map[Result]bool) (int, int) {
	i := 0
	u := 0
	for x := range a {
		if b[x] {
			u++
			i++
		} else {
			u++
		}
	}
	for x := range b {
		if !a[x] {
			u++
		}
	}
	return i, u
}

func answerSets() {
	worst, worstScore := "ZZZZ", 100000
	gots := map[string]map[Result]bool{}
	for _, target := range Answers {
		got := map[Result]bool{}
		for _, guess := range AllWords {
			got[Score(target, guess)] = true
		}
		if len(got) < worstScore {
			worst, worstScore = target, len(got)
		}
		gots[target] = got
	}
	fmt.Println("worst:", worst, worstScore)
	type wwiu struct {
		w1, w2 string
		i, u   int
	}
	all := []wwiu{}
	for i := range Answers {
		for j := i + 1; j < len(Answers); j++ {
			i, u := sim(gots[Answers[i]], gots[Answers[j]])
			all = append(all, wwiu{Answers[i], Answers[j], i, u})
		}
	}
	sort.Slice(all, func(i, j int) bool {
		s1 := float64(all[i].i) / float64(all[i].u)
		s2 := float64(all[j].i) / float64(all[j].u)
		return s1 > s2
	})
	for i := 0; i < 10; i++ {
		fmt.Println(all[i])
	}
}

func mprs(s string) Result {
	r, err := ResultFromString(s)
	if err != nil {
		panic(err)
	}
	return r
}

func mustParseResults(s ...string) []Result {
	var r []Result
	for _, x := range s {
		r = append(r, mprs(x))
	}
	return r
}

func findAnswer(wants []Result) {
	for _, target := range Answers {
		got := map[Result]bool{}
		for _, guess := range AllWords {
			got[Score(target, guess)] = true
		}
		ok := true
		for _, w := range wants {
			if !got[w] {
				ok = false
				break
			}
		}
		if ok {
			fmt.Println(target)
		}
	}
}

func main() {
	checkStrats()
	// answerSets()
}
