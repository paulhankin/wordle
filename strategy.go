package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Result uint8

type ResultLetter int

const (
	Black  = ResultLetter(0)
	Green  = ResultLetter(1)
	Yellow = ResultLetter(2)
)

const AllGreen Result = 121
const MaxResult Result = 242

// Score returns the result of matching a guess with the target string.
func Score(target, guess string) Result {
	var r [5]ResultLetter
	var used [5]bool
	for i := 0; i < 5; i++ {
		if target[i] == guess[i] {
			r[i] = Green
			used[i] = true
		}
	}
	for i := 0; i < 5; i++ {
		if r[i] != Black {
			continue
		}
		for j := 0; j < 5; j++ {
			if used[j] {
				continue
			}
			if target[j] == guess[i] {
				r[i] = Yellow
				used[j] = true
				break
			}
		}
	}
	return NewResult(r)
}

func (r ResultLetter) String() string {
	switch r {
	case Black:
		return "B"
	case Yellow:
		return "Y"
	case Green:
		return "G"
	default:
		return "?"
	}
}

func NewResult(ls [5]ResultLetter) Result {
	var r Result
	for _, rl := range ls {
		r = r*3 + Result(rl)
	}
	return r
}

func ResultFromString(s string) (Result, error) {
	if len(s) == 6 && s[5] >= '1' && s[5] <= '9' {
		// Allow a guess-count to be present at the end of the result string.
		// We ignore it.
		s = s[:5]
	}
	if len(s) != 5 {
		return 0, fmt.Errorf("result string %q must be of the form [GBY]^5", s)
	}
	var r Result
	for i := 0; i < 5; i++ {
		switch s[i] {
		case 'B':
			r = (r * 3) + 0
		case 'Y':
			r = (r * 3) + 2
		case 'G':
			r = (r * 3) + 1
		default:
			return 0, fmt.Errorf("error in result string %q at index %d: illegal char %c", s, i, s[i])
		}
	}
	return r, nil
}

func (res Result) Letters() [5]ResultLetter {
	var r [5]ResultLetter
	for i := 0; i < 5; i++ {
		r[4-i] = ResultLetter(res % 3)
		res /= 3
	}
	return r
}

func (res Result) String() string {
	var b strings.Builder
	for _, l := range res.Letters() {
		b.WriteString(l.String())
	}
	return b.String()
}

type Choice struct {
	W string
	N map[Result]*Choice
}

var spaces = strings.Repeat(" ", 200)

func (c *Choice) format(b *strings.Builder, indent int) {
	b.WriteString(c.W)
	b.WriteString(" ")
	first := true
	for i := Result(0); i <= MaxResult; i++ {
		c2, ok := c.N[i]
		if !ok {
			continue
		}
		if !first {
			b.WriteByte('\n')
			b.WriteString(spaces[:indent+6])
		}
		first = false
		b.WriteString(i.String())
		b.WriteByte("123456789"[indent/13])
		if c2 != nil {
			b.WriteByte(' ')
			c2.format(b, indent+13)
		}
	}
}

func (c *Choice) String() string {
	var b strings.Builder
	c.format(&b, 0)
	return b.String()
}

func (c *Choice) Guess(prev []string) string {
	return c.W
}

func (c *Choice) Next(r Result) (*Choice, error) {
	c2, ok := c.N[r]
	if !ok {
		return nil, fmt.Errorf("result %s invalid", r)
	}
	return c2, nil
}

func LoadSelbyStrategy(r io.Reader) (*Choice, error) {
	var stack []*Choice
	scanner := bufio.NewScanner(r)
	colpos := func(i int) int {
		r := 13 * (i / 2)
		r += 6 * (i % 2)
		return r
	}
	collen := func(i int) int {
		return 5 + (i % 2)
	}
	for scanner.Scan() {
		line := scanner.Text()
		var cols []string
		for i := 0; colpos(i) < len(line); i++ {
			s := strings.TrimSpace(line[colpos(i) : colpos(i)+collen(i)])
			cols = append(cols, s)
		}
		if r, err := ResultFromString(cols[len(cols)-1]); err != nil || r != AllGreen {
			return nil, fmt.Errorf("line %q does not end with GGGGG", line)
		}
		for i := 0; i < len(cols); i += 2 {
			if cols[i] == "" {
				if i > 0 && cols[i-1] != "" {
					return nil, fmt.Errorf("found result but no word at col %d in %q", i, line)
				}
				continue
			}
			if !WordDict[cols[i]] {
				return nil, fmt.Errorf("strategy guesses %q which is not an approved word", cols[i])
			}
			if i == 0 {
				if len(stack) > 0 {
					return nil, fmt.Errorf("found two initial guesses (second on line %q)", line)
				}
				stack = append(stack, &Choice{W: cols[i]})
			} else {
				r, err := ResultFromString(cols[i-1])
				if err != nil {
					return nil, fmt.Errorf("no valid result found for word %q on line %q", cols[i], line)
				}
				if i/2 > len(stack) {
					return nil, fmt.Errorf("found no parent for word %q on line %q", cols[i], line)
				}
				stack = stack[:i/2]
				top := stack[i/2-1]
				stack = append(stack, &Choice{W: cols[i]})
				if top.N == nil {
					top.N = map[Result]*Choice{}
				}
				if _, ok := top.N[r]; ok {
					return nil, fmt.Errorf("found duplicate results %s for %q", r, top.W)
				}
				top.N[r] = stack[i/2]
			}
		}
		top := stack[len(cols)/2-1]
		if top.N == nil {
			top.N = map[Result]*Choice{}
		}
		if _, ok := top.N[AllGreen]; ok {
			return nil, fmt.Errorf("found duplicate GGGGG for %q", top.W)
		}
		top.N[AllGreen] = nil

	}
	if len(stack) == 0 {
		return nil, fmt.Errorf("file empty?")
	}
	return stack[0], nil
}
