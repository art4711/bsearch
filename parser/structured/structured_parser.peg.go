package structured

import (
	/*"bytes"*/
	"fmt"
	"math"
	"sort"
	"strconv"
	"bsearch/parser/opers"
)

const END_SYMBOL rune = 4

/* The rule types inferred from the grammar are below. */
type Rule uint8

const (
	RuleUnknown Rule = iota
	RuleQuery
	RuleOperation
	RuleOffset
	RuleLimit
	RuleCountAll
	RuleIntersection
	RuleUnion
	RuleAttr
	RuleName
	RuleIntValue
	Rulenumber
	Rules
	Rulegeneric_name
	RuleAction0
	RuleAction1
	RuleAction2
	RuleAction3
	RuleAction4
	RuleAction5
	RuleAction6
	RuleAction7
	RuleAction8
	RuleAction9
	RuleAction10
	RulePegText

	RulePre_
	Rule_In_
	Rule_Suf
)

var Rul3s = [...]string{
	"Unknown",
	"Query",
	"Operation",
	"Offset",
	"Limit",
	"CountAll",
	"Intersection",
	"Union",
	"Attr",
	"Name",
	"IntValue",
	"number",
	"s",
	"generic_name",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"PegText",

	"Pre_",
	"_In_",
	"_Suf",
}

type TokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule Rule, begin, end, next, depth int)
	Expand(index int) TokenTree
	Tokens() <-chan token32
	Error() []token32
	trim(length int)
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	Rule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) PreOrder() (<-chan State16, [][]token16) {
	s, ordered := make(chan State16, 6), t.Order()
	go func() {
		var states [8]State16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token16{Rule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	Rule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) PreOrder() (<-chan State32, [][]token32) {
	s, ordered := make(chan State32, 6), t.Order()
	go func() {
		var states [8]State32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token32{Rule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.GetToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type StructuredParser struct {
	opers.Query

	Buffer string
	buffer []rune
	rules  [26]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	TokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *StructuredParser
}

func (e *parseError) Error() string {
	tokens, error := e.p.TokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			Rul3s[token.Rule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *StructuredParser) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *StructuredParser) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *StructuredParser) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.TokenTree.Tokens() {
		switch token.Rule {
		case RulePegText:
			begin, end = int(token.begin), int(token.end)
		case RuleAction0:
			p.Off(buffer[begin:end])
		case RuleAction1:
			p.Pa()
		case RuleAction2:
			p.Lim(buffer[begin:end])
		case RuleAction3:
			p.Pa()
		case RuleAction4:
			p.Countall(buffer[begin:end])
		case RuleAction5:
			p.Pa()
		case RuleAction6:
			p.Inter()
		case RuleAction7:
			p.Pa()
		case RuleAction8:
			p.Union()
		case RuleAction9:
			p.Pa()
		case RuleAction10:
			p.Attr(buffer[begin:end])

		}
	}
}

func (p *StructuredParser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != END_SYMBOL {
		p.buffer = append(p.buffer, END_SYMBOL)
	}

	var tree TokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.TokenTree = tree
		if matches {
			p.TokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule Rule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != END_SYMBOL {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	rules = [...]func() bool{
		nil,
		/* 0 Query <- <(Operation !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !rules[RuleOperation]() {
					goto l0
				}
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
				depth--
				add(RuleQuery, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Operation <- <(Offset / Limit / CountAll / Intersection / Union / Attr)> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				{
					position5, tokenIndex5, depth5 := position, tokenIndex, depth
					if !rules[RuleOffset]() {
						goto l6
					}
					goto l5
				l6:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
					if !rules[RuleLimit]() {
						goto l7
					}
					goto l5
				l7:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
					if !rules[RuleCountAll]() {
						goto l8
					}
					goto l5
				l8:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
					if !rules[RuleIntersection]() {
						goto l9
					}
					goto l5
				l9:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
					if !rules[RuleUnion]() {
						goto l10
					}
					goto l5
				l10:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
					if !rules[RuleAttr]() {
						goto l3
					}
				}
			l5:
				depth--
				add(RuleOperation, position4)
			}
			return true
		l3:
			position, tokenIndex, depth = position3, tokenIndex3, depth3
			return false
		},
		/* 2 Offset <- <('(' 'o' 'f' 'f' 's' 'e' 't' s IntValue Action0 s Operation ')' Action1)> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
				if buffer[position] != rune('(') {
					goto l11
				}
				position++
				if buffer[position] != rune('o') {
					goto l11
				}
				position++
				if buffer[position] != rune('f') {
					goto l11
				}
				position++
				if buffer[position] != rune('f') {
					goto l11
				}
				position++
				if buffer[position] != rune('s') {
					goto l11
				}
				position++
				if buffer[position] != rune('e') {
					goto l11
				}
				position++
				if buffer[position] != rune('t') {
					goto l11
				}
				position++
				if !rules[Rules]() {
					goto l11
				}
				if !rules[RuleIntValue]() {
					goto l11
				}
				if !rules[RuleAction0]() {
					goto l11
				}
				if !rules[Rules]() {
					goto l11
				}
				if !rules[RuleOperation]() {
					goto l11
				}
				if buffer[position] != rune(')') {
					goto l11
				}
				position++
				if !rules[RuleAction1]() {
					goto l11
				}
				depth--
				add(RuleOffset, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 3 Limit <- <('(' 'l' 'i' 'm' 'i' 't' s IntValue Action2 s Operation ')' Action3)> */
		func() bool {
			position13, tokenIndex13, depth13 := position, tokenIndex, depth
			{
				position14 := position
				depth++
				if buffer[position] != rune('(') {
					goto l13
				}
				position++
				if buffer[position] != rune('l') {
					goto l13
				}
				position++
				if buffer[position] != rune('i') {
					goto l13
				}
				position++
				if buffer[position] != rune('m') {
					goto l13
				}
				position++
				if buffer[position] != rune('i') {
					goto l13
				}
				position++
				if buffer[position] != rune('t') {
					goto l13
				}
				position++
				if !rules[Rules]() {
					goto l13
				}
				if !rules[RuleIntValue]() {
					goto l13
				}
				if !rules[RuleAction2]() {
					goto l13
				}
				if !rules[Rules]() {
					goto l13
				}
				if !rules[RuleOperation]() {
					goto l13
				}
				if buffer[position] != rune(')') {
					goto l13
				}
				position++
				if !rules[RuleAction3]() {
					goto l13
				}
				depth--
				add(RuleLimit, position14)
			}
			return true
		l13:
			position, tokenIndex, depth = position13, tokenIndex13, depth13
			return false
		},
		/* 4 CountAll <- <('(' 'c' 'o' 'u' 'n' 't' '_' 'a' 'l' 'l' s Name Action4 s Operation ')' Action5)> */
		func() bool {
			position15, tokenIndex15, depth15 := position, tokenIndex, depth
			{
				position16 := position
				depth++
				if buffer[position] != rune('(') {
					goto l15
				}
				position++
				if buffer[position] != rune('c') {
					goto l15
				}
				position++
				if buffer[position] != rune('o') {
					goto l15
				}
				position++
				if buffer[position] != rune('u') {
					goto l15
				}
				position++
				if buffer[position] != rune('n') {
					goto l15
				}
				position++
				if buffer[position] != rune('t') {
					goto l15
				}
				position++
				if buffer[position] != rune('_') {
					goto l15
				}
				position++
				if buffer[position] != rune('a') {
					goto l15
				}
				position++
				if buffer[position] != rune('l') {
					goto l15
				}
				position++
				if buffer[position] != rune('l') {
					goto l15
				}
				position++
				if !rules[Rules]() {
					goto l15
				}
				if !rules[RuleName]() {
					goto l15
				}
				if !rules[RuleAction4]() {
					goto l15
				}
				if !rules[Rules]() {
					goto l15
				}
				if !rules[RuleOperation]() {
					goto l15
				}
				if buffer[position] != rune(')') {
					goto l15
				}
				position++
				if !rules[RuleAction5]() {
					goto l15
				}
				depth--
				add(RuleCountAll, position16)
			}
			return true
		l15:
			position, tokenIndex, depth = position15, tokenIndex15, depth15
			return false
		},
		/* 5 Intersection <- <('(' 'i' 'n' 't' 'e' 'r' 's' 'e' 'c' 't' 'i' 'o' 'n' Action6 (s Operation)+ ')' Action7)> */
		func() bool {
			position17, tokenIndex17, depth17 := position, tokenIndex, depth
			{
				position18 := position
				depth++
				if buffer[position] != rune('(') {
					goto l17
				}
				position++
				if buffer[position] != rune('i') {
					goto l17
				}
				position++
				if buffer[position] != rune('n') {
					goto l17
				}
				position++
				if buffer[position] != rune('t') {
					goto l17
				}
				position++
				if buffer[position] != rune('e') {
					goto l17
				}
				position++
				if buffer[position] != rune('r') {
					goto l17
				}
				position++
				if buffer[position] != rune('s') {
					goto l17
				}
				position++
				if buffer[position] != rune('e') {
					goto l17
				}
				position++
				if buffer[position] != rune('c') {
					goto l17
				}
				position++
				if buffer[position] != rune('t') {
					goto l17
				}
				position++
				if buffer[position] != rune('i') {
					goto l17
				}
				position++
				if buffer[position] != rune('o') {
					goto l17
				}
				position++
				if buffer[position] != rune('n') {
					goto l17
				}
				position++
				if !rules[RuleAction6]() {
					goto l17
				}
				if !rules[Rules]() {
					goto l17
				}
				if !rules[RuleOperation]() {
					goto l17
				}
			l19:
				{
					position20, tokenIndex20, depth20 := position, tokenIndex, depth
					if !rules[Rules]() {
						goto l20
					}
					if !rules[RuleOperation]() {
						goto l20
					}
					goto l19
				l20:
					position, tokenIndex, depth = position20, tokenIndex20, depth20
				}
				if buffer[position] != rune(')') {
					goto l17
				}
				position++
				if !rules[RuleAction7]() {
					goto l17
				}
				depth--
				add(RuleIntersection, position18)
			}
			return true
		l17:
			position, tokenIndex, depth = position17, tokenIndex17, depth17
			return false
		},
		/* 6 Union <- <('(' 'u' 'n' 'i' 'o' 'n' Action8 (s Operation)+ ')' Action9)> */
		func() bool {
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
				if buffer[position] != rune('(') {
					goto l21
				}
				position++
				if buffer[position] != rune('u') {
					goto l21
				}
				position++
				if buffer[position] != rune('n') {
					goto l21
				}
				position++
				if buffer[position] != rune('i') {
					goto l21
				}
				position++
				if buffer[position] != rune('o') {
					goto l21
				}
				position++
				if buffer[position] != rune('n') {
					goto l21
				}
				position++
				if !rules[RuleAction8]() {
					goto l21
				}
				if !rules[Rules]() {
					goto l21
				}
				if !rules[RuleOperation]() {
					goto l21
				}
			l23:
				{
					position24, tokenIndex24, depth24 := position, tokenIndex, depth
					if !rules[Rules]() {
						goto l24
					}
					if !rules[RuleOperation]() {
						goto l24
					}
					goto l23
				l24:
					position, tokenIndex, depth = position24, tokenIndex24, depth24
				}
				if buffer[position] != rune(')') {
					goto l21
				}
				position++
				if !rules[RuleAction9]() {
					goto l21
				}
				depth--
				add(RuleUnion, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 7 Attr <- <('(' 'a' 't' 't' 'r' s Name ')' Action10)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
				if buffer[position] != rune('(') {
					goto l25
				}
				position++
				if buffer[position] != rune('a') {
					goto l25
				}
				position++
				if buffer[position] != rune('t') {
					goto l25
				}
				position++
				if buffer[position] != rune('t') {
					goto l25
				}
				position++
				if buffer[position] != rune('r') {
					goto l25
				}
				position++
				if !rules[Rules]() {
					goto l25
				}
				if !rules[RuleName]() {
					goto l25
				}
				if buffer[position] != rune(')') {
					goto l25
				}
				position++
				if !rules[RuleAction10]() {
					goto l25
				}
				depth--
				add(RuleAttr, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 8 Name <- <('"' <generic_name> '"')> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				if buffer[position] != rune('"') {
					goto l27
				}
				position++
				{
					position29 := position
					depth++
					if !rules[Rulegeneric_name]() {
						goto l27
					}
					depth--
					add(RulePegText, position29)
				}
				if buffer[position] != rune('"') {
					goto l27
				}
				position++
				depth--
				add(RuleName, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 9 IntValue <- <('[' s <number> s ']')> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if buffer[position] != rune('[') {
					goto l30
				}
				position++
				if !rules[Rules]() {
					goto l30
				}
				{
					position32 := position
					depth++
					if !rules[Rulenumber]() {
						goto l30
					}
					depth--
					add(RulePegText, position32)
				}
				if !rules[Rules]() {
					goto l30
				}
				if buffer[position] != rune(']') {
					goto l30
				}
				position++
				depth--
				add(RuleIntValue, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 10 number <- <[0-9]+> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l33
				}
				position++
			l35:
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l36
					}
					position++
					goto l35
				l36:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
				}
				depth--
				add(Rulenumber, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 11 s <- <' '+> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l37
				}
				position++
			l39:
				{
					position40, tokenIndex40, depth40 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l40
					}
					position++
					goto l39
				l40:
					position, tokenIndex, depth = position40, tokenIndex40, depth40
				}
				depth--
				add(Rules, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 12 generic_name <- <([a-z] / [0-9] / '_' / ':')+> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l46
					}
					position++
					goto l45
				l46:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l47
					}
					position++
					goto l45
				l47:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if buffer[position] != rune('_') {
						goto l48
					}
					position++
					goto l45
				l48:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if buffer[position] != rune(':') {
						goto l41
					}
					position++
				}
			l45:
			l43:
				{
					position44, tokenIndex44, depth44 := position, tokenIndex, depth
					{
						position49, tokenIndex49, depth49 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l50
						}
						position++
						goto l49
					l50:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l51
						}
						position++
						goto l49
					l51:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if buffer[position] != rune('_') {
							goto l52
						}
						position++
						goto l49
					l52:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if buffer[position] != rune(':') {
							goto l44
						}
						position++
					}
				l49:
					goto l43
				l44:
					position, tokenIndex, depth = position44, tokenIndex44, depth44
				}
				depth--
				add(Rulegeneric_name, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 14 Action0 <- <{ p.Off(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction0, position)
			}
			return true
		},
		/* 15 Action1 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction1, position)
			}
			return true
		},
		/* 16 Action2 <- <{ p.Lim(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction2, position)
			}
			return true
		},
		/* 17 Action3 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction3, position)
			}
			return true
		},
		/* 18 Action4 <- <{ p.Countall(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction4, position)
			}
			return true
		},
		/* 19 Action5 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction5, position)
			}
			return true
		},
		/* 20 Action6 <- <{ p.Inter() }> */
		func() bool {
			{
				add(RuleAction6, position)
			}
			return true
		},
		/* 21 Action7 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction7, position)
			}
			return true
		},
		/* 22 Action8 <- <{ p.Union() }> */
		func() bool {
			{
				add(RuleAction8, position)
			}
			return true
		},
		/* 23 Action9 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction9, position)
			}
			return true
		},
		/* 24 Action10 <- <{ p.Attr(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction10, position)
			}
			return true
		},
		nil,
	}
	p.rules = rules
}
