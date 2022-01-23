package main

import (
	"fmt"
	"log"
	"os"
)

func LoadStrategy(name string) (*Choice, error) {
	f, err := os.Open("strategy.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s, err := LoadSelbyStrategy(f)
	return s, err
}

func Play(c *Choice, target string) int {
	for n := 1; ; n++ {
		g := c.Guess()
		sc := Score(target, g)
		if sc == AllGreen {
			return n
		}
		var err error
		c, err = c.Next(sc)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	s, err := LoadStrategy("strategy.txt")
	if err != nil {
		log.Fatal(err)
	}
	sum := 0
	for _, w := range Answers {
		sum += Play(s, w)
	}
	fmt.Println("sum =", sum, "mean =", float64(sum)/float64(len(Answers)))
}
