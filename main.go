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

func Play(c *Choice, target string, hard bool) int {
	var prev []string
	for n := 1; ; n++ {
		g := c.Guess(prev)
		sc := Score(target, g)
		if sc == AllGreen {
			return n
		}
		var err error
		c, err = c.Next(sc)
		if err != nil {
			panic(err)
		}
		prev = append(prev, g)
	}
}

func main() {
	for hard := 0; hard <= 1; hard++ {
		s, err := LoadStrategy([]string{"strategy_normal.txt", "strategy_hard.txt"}[hard])
		if err != nil {
			log.Fatal(err)
		}
		sum := 0
		for _, w := range Answers {
			sum += Play(s, w, hard == 1)
		}
		fmt.Println("hard =", hard, "sum =", sum, "mean =", float64(sum)/float64(len(Answers)))
	}
}
