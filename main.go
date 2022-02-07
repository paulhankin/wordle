package main

import (
	"fmt"
	"log"
	"math/bits"
	"os"
	"sort"
	"strings"
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

var allWords = []string{
	"bemix", "bling", "blunk", "brick", "brung", "chunk", "cimex", "clipt", "clunk", "cylix", "fjord", "glent", "grypt", "gucks", "gymps", "jumby", "jumpy", "kempt", "kreng", "nymph", "pling", "prick", "treck", "vibex", "vozhd", "waltz", "waqfs", "xylic",
}

func printSol(res []uint32, masks map[uint32][]string) {
	var b strings.Builder
	for i, r := range res {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(strings.Join(masks[r], "|"))
	}
	fmt.Println(b.String())
}

func find5(w []uint32, mask uint32, n int, res []uint32, masks map[uint32][]string) {
	if n == 5 {
		printSol(res, masks)
		return
	}
	sub := []uint32{}
	for _, x := range w {
		if x&mask != 0 {
			continue
		}
		sub = append(sub, x)
	}
	for i, x := range sub {
		res[n] = x
		find5(sub[i+1:], mask|x, n+1, res, masks)
	}
}

func find5clique() {
	masks := map[uint32][]string{}
	for _, x := range allWords {
		m := uint32(0)
		for _, c := range x {
			m |= 1 << (c - 'a')
		}
		if bits.OnesCount32(m) == 5 {
			masks[m] = append(masks[m], x)
		}
	}
	maskSlice := []uint32{}
	for m := range masks {
		maskSlice = append(maskSlice, m)
	}
	sort.Slice(maskSlice, func(i, j int) bool {
		return maskSlice[i] < maskSlice[j]
	})
	find5(maskSlice, uint32(0), 0, make([]uint32, 5, 5), masks)
}

func main() {
	checkStrats()
}
