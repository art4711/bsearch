package parser

import (
	/*"bytes"*/
	"fmt"
	"math"
	"sort"
	"strconv"
)

const END_SYMBOL rune = 4

/* The rule types inferred from the grammar are below. */
type Rule uint8

const (
	RuleUnknown Rule = iota
	RuleQuery
	RuleResFiltQuery
	RuleOffLimQuery
	RuleLimQuery
	RuleQ3
	RuleOffset
	RuleLimit
	RuleParams
	RuleCountAllAttrs
	RuleCountAll
	RuleAttrs
	RuleAttr
	RuleAttribute
	RuleAttrUnion
	RuleAttributeORList
	Rulenumber
	Rules
	Rulecounter_name
	Ruleattr_name
	Ruleattr_value
	Rulegeneric_name
	RuleAction0
	RuleAction1
	RulePegText
	RuleAction2
	RuleAction3
	RuleAction4
	RuleAction5
	RuleAction6
	RuleAction7
	RuleAction8
	RuleAction9
	RuleAction10
	RuleAction11

	RulePre_
	Rule_In_
	Rule_Suf
)

var Rul3s = [...]string{
	"Unknown",
	"Query",
	"ResFiltQuery",
	"OffLimQuery",
	"LimQuery",
	"Q3",
	"Offset",
	"Limit",
	"Params",
	"CountAllAttrs",
	"CountAll",
	"Attrs",
	"Attr",
	"Attribute",
	"AttrUnion",
	"AttributeORList",
	"number",
	"s",
	"counter_name",
	"attr_name",
	"attr_value",
	"generic_name",
	"Action0",
	"Action1",
	"PegText",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",

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

type ClassicParser struct {
	Query

	Buffer string
	buffer []rune
	rules  [35]func() bool
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
	p *ClassicParser
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

func (p *ClassicParser) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *ClassicParser) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *ClassicParser) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.TokenTree.Tokens() {
		switch token.Rule {
		case RulePegText:
			begin, end = int(token.begin), int(token.end)
		case RuleAction0:
			p.Pa()
		case RuleAction1:
			p.Pa()
		case RuleAction2:
			p.Off(buffer[begin:end])
		case RuleAction3:
			p.Lim(buffer[begin:end])
		case RuleAction4:
			p.Pa()
		case RuleAction5:
			p.Countall(buffer[begin:end])
		case RuleAction6:
			p.Inter()
		case RuleAction7:
			p.Inter()
			fmt.Printf("one attribute\n")
		case RuleAction8:
			fmt.Printf("no attributes\n")
		case RuleAction9:
			p.Attr(buffer[begin:end])
		case RuleAction10:
			p.Union()
		case RuleAction11:
			p.Pa()

		}
	}
}

func (p *ClassicParser) Init() {
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
		/* 0 Query <- <(ResFiltQuery !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !rules[RuleResFiltQuery]() {
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
		/* 1 ResFiltQuery <- <OffLimQuery> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				if !rules[RuleOffLimQuery]() {
					goto l3
				}
				depth--
				add(RuleResFiltQuery, position4)
			}
			return true
		l3:
			position, tokenIndex, depth = position3, tokenIndex3, depth3
			return false
		},
		/* 2 OffLimQuery <- <((Offset LimQuery Action0) / LimQuery)> */
		func() bool {
			position5, tokenIndex5, depth5 := position, tokenIndex, depth
			{
				position6 := position
				depth++
				{
					position7, tokenIndex7, depth7 := position, tokenIndex, depth
					if !rules[RuleOffset]() {
						goto l8
					}
					if !rules[RuleLimQuery]() {
						goto l8
					}
					if !rules[RuleAction0]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
					if !rules[RuleLimQuery]() {
						goto l5
					}
				}
			l7:
				depth--
				add(RuleOffLimQuery, position6)
			}
			return true
		l5:
			position, tokenIndex, depth = position5, tokenIndex5, depth5
			return false
		},
		/* 3 LimQuery <- <((Limit Q3 Action1) / Q3)> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
				{
					position11, tokenIndex11, depth11 := position, tokenIndex, depth
					if !rules[RuleLimit]() {
						goto l12
					}
					if !rules[RuleQ3]() {
						goto l12
					}
					if !rules[RuleAction1]() {
						goto l12
					}
					goto l11
				l12:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
					if !rules[RuleQ3]() {
						goto l9
					}
				}
			l11:
				depth--
				add(RuleLimQuery, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 4 Q3 <- <Params?> */
		func() bool {
			{
				position14 := position
				depth++
				{
					position15, tokenIndex15, depth15 := position, tokenIndex, depth
					if !rules[RuleParams]() {
						goto l15
					}
					goto l16
				l15:
					position, tokenIndex, depth = position15, tokenIndex15, depth15
				}
			l16:
				depth--
				add(RuleQ3, position14)
			}
			return true
		},
		/* 5 Offset <- <(<number> s Action2)> */
		func() bool {
			position17, tokenIndex17, depth17 := position, tokenIndex, depth
			{
				position18 := position
				depth++
				{
					position19 := position
					depth++
					if !rules[Rulenumber]() {
						goto l17
					}
					depth--
					add(RulePegText, position19)
				}
				if !rules[Rules]() {
					goto l17
				}
				if !rules[RuleAction2]() {
					goto l17
				}
				depth--
				add(RuleOffset, position18)
			}
			return true
		l17:
			position, tokenIndex, depth = position17, tokenIndex17, depth17
			return false
		},
		/* 6 Limit <- <('l' 'i' 'm' ':' <number> s Action3)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				if buffer[position] != rune('l') {
					goto l20
				}
				position++
				if buffer[position] != rune('i') {
					goto l20
				}
				position++
				if buffer[position] != rune('m') {
					goto l20
				}
				position++
				if buffer[position] != rune(':') {
					goto l20
				}
				position++
				{
					position22 := position
					depth++
					if !rules[Rulenumber]() {
						goto l20
					}
					depth--
					add(RulePegText, position22)
				}
				if !rules[Rules]() {
					goto l20
				}
				if !rules[RuleAction3]() {
					goto l20
				}
				depth--
				add(RuleLimit, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 7 Params <- <(CountAllAttrs / Attrs)> */
		func() bool {
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				{
					position25, tokenIndex25, depth25 := position, tokenIndex, depth
					if !rules[RuleCountAllAttrs]() {
						goto l26
					}
					goto l25
				l26:
					position, tokenIndex, depth = position25, tokenIndex25, depth25
					if !rules[RuleAttrs]() {
						goto l23
					}
				}
			l25:
				depth--
				add(RuleParams, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 8 CountAllAttrs <- <((CountAll Attrs Action4) / Attrs)> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				{
					position29, tokenIndex29, depth29 := position, tokenIndex, depth
					if !rules[RuleCountAll]() {
						goto l30
					}
					if !rules[RuleAttrs]() {
						goto l30
					}
					if !rules[RuleAction4]() {
						goto l30
					}
					goto l29
				l30:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
					if !rules[RuleAttrs]() {
						goto l27
					}
				}
			l29:
				depth--
				add(RuleCountAllAttrs, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 9 CountAll <- <('c' 'o' 'u' 'n' 't' '_' 'a' 'l' 'l' '(' <counter_name> ')' s Action5)> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				if buffer[position] != rune('c') {
					goto l31
				}
				position++
				if buffer[position] != rune('o') {
					goto l31
				}
				position++
				if buffer[position] != rune('u') {
					goto l31
				}
				position++
				if buffer[position] != rune('n') {
					goto l31
				}
				position++
				if buffer[position] != rune('t') {
					goto l31
				}
				position++
				if buffer[position] != rune('_') {
					goto l31
				}
				position++
				if buffer[position] != rune('a') {
					goto l31
				}
				position++
				if buffer[position] != rune('l') {
					goto l31
				}
				position++
				if buffer[position] != rune('l') {
					goto l31
				}
				position++
				if buffer[position] != rune('(') {
					goto l31
				}
				position++
				{
					position33 := position
					depth++
					if !rules[Rulecounter_name]() {
						goto l31
					}
					depth--
					add(RulePegText, position33)
				}
				if buffer[position] != rune(')') {
					goto l31
				}
				position++
				if !rules[Rules]() {
					goto l31
				}
				if !rules[RuleAction5]() {
					goto l31
				}
				depth--
				add(RuleCountAll, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 10 Attrs <- <((Action6 (Attr s)+ Attr s?) / (Action7 Attr s?) / (s? Action8))> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					if !rules[RuleAction6]() {
						goto l37
					}
					if !rules[RuleAttr]() {
						goto l37
					}
					if !rules[Rules]() {
						goto l37
					}
				l38:
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						if !rules[RuleAttr]() {
							goto l39
						}
						if !rules[Rules]() {
							goto l39
						}
						goto l38
					l39:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
					}
					if !rules[RuleAttr]() {
						goto l37
					}
					{
						position40, tokenIndex40, depth40 := position, tokenIndex, depth
						if !rules[Rules]() {
							goto l40
						}
						goto l41
					l40:
						position, tokenIndex, depth = position40, tokenIndex40, depth40
					}
				l41:
					goto l36
				l37:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
					if !rules[RuleAction7]() {
						goto l42
					}
					if !rules[RuleAttr]() {
						goto l42
					}
					{
						position43, tokenIndex43, depth43 := position, tokenIndex, depth
						if !rules[Rules]() {
							goto l43
						}
						goto l44
					l43:
						position, tokenIndex, depth = position43, tokenIndex43, depth43
					}
				l44:
					goto l36
				l42:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
					{
						position45, tokenIndex45, depth45 := position, tokenIndex, depth
						if !rules[Rules]() {
							goto l45
						}
						goto l46
					l45:
						position, tokenIndex, depth = position45, tokenIndex45, depth45
					}
				l46:
					if !rules[RuleAction8]() {
						goto l34
					}
				}
			l36:
				depth--
				add(RuleAttrs, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 11 Attr <- <(AttrUnion / Attribute)> */
		func() bool {
			position47, tokenIndex47, depth47 := position, tokenIndex, depth
			{
				position48 := position
				depth++
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					if !rules[RuleAttrUnion]() {
						goto l50
					}
					goto l49
				l50:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
					if !rules[RuleAttribute]() {
						goto l47
					}
				}
			l49:
				depth--
				add(RuleAttr, position48)
			}
			return true
		l47:
			position, tokenIndex, depth = position47, tokenIndex47, depth47
			return false
		},
		/* 12 Attribute <- <(<(attr_name ':' attr_value)> Action9)> */
		func() bool {
			position51, tokenIndex51, depth51 := position, tokenIndex, depth
			{
				position52 := position
				depth++
				{
					position53 := position
					depth++
					if !rules[Ruleattr_name]() {
						goto l51
					}
					if buffer[position] != rune(':') {
						goto l51
					}
					position++
					if !rules[Ruleattr_value]() {
						goto l51
					}
					depth--
					add(RulePegText, position53)
				}
				if !rules[RuleAction9]() {
					goto l51
				}
				depth--
				add(RuleAttribute, position52)
			}
			return true
		l51:
			position, tokenIndex, depth = position51, tokenIndex51, depth51
			return false
		},
		/* 13 AttrUnion <- <(Action10 AttributeORList Action11)> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				if !rules[RuleAction10]() {
					goto l54
				}
				if !rules[RuleAttributeORList]() {
					goto l54
				}
				if !rules[RuleAction11]() {
					goto l54
				}
				depth--
				add(RuleAttrUnion, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 14 AttributeORList <- <(Attribute (s ('O' 'R') s Attribute)+)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if !rules[RuleAttribute]() {
					goto l56
				}
				if !rules[Rules]() {
					goto l56
				}
				if buffer[position] != rune('O') {
					goto l56
				}
				position++
				if buffer[position] != rune('R') {
					goto l56
				}
				position++
				if !rules[Rules]() {
					goto l56
				}
				if !rules[RuleAttribute]() {
					goto l56
				}
			l58:
				{
					position59, tokenIndex59, depth59 := position, tokenIndex, depth
					if !rules[Rules]() {
						goto l59
					}
					if buffer[position] != rune('O') {
						goto l59
					}
					position++
					if buffer[position] != rune('R') {
						goto l59
					}
					position++
					if !rules[Rules]() {
						goto l59
					}
					if !rules[RuleAttribute]() {
						goto l59
					}
					goto l58
				l59:
					position, tokenIndex, depth = position59, tokenIndex59, depth59
				}
				depth--
				add(RuleAttributeORList, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 15 number <- <[0-9]+> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l60
				}
				position++
			l62:
				{
					position63, tokenIndex63, depth63 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l63
					}
					position++
					goto l62
				l63:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
				}
				depth--
				add(Rulenumber, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 16 s <- <' '+> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l64
				}
				position++
			l66:
				{
					position67, tokenIndex67, depth67 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l67
					}
					position++
					goto l66
				l67:
					position, tokenIndex, depth = position67, tokenIndex67, depth67
				}
				depth--
				add(Rules, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 17 counter_name <- <generic_name> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				if !rules[Rulegeneric_name]() {
					goto l68
				}
				depth--
				add(Rulecounter_name, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 18 attr_name <- <generic_name> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				if !rules[Rulegeneric_name]() {
					goto l70
				}
				depth--
				add(Ruleattr_name, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 19 attr_value <- <generic_name> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				if !rules[Rulegeneric_name]() {
					goto l72
				}
				depth--
				add(Ruleattr_value, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 20 generic_name <- <([a-z] / [0-9] / '_')+> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				{
					position78, tokenIndex78, depth78 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l79
					}
					position++
					goto l78
				l79:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l80
					}
					position++
					goto l78
				l80:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if buffer[position] != rune('_') {
						goto l74
					}
					position++
				}
			l78:
			l76:
				{
					position77, tokenIndex77, depth77 := position, tokenIndex, depth
					{
						position81, tokenIndex81, depth81 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l82
						}
						position++
						goto l81
					l82:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l83
						}
						position++
						goto l81
					l83:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
						if buffer[position] != rune('_') {
							goto l77
						}
						position++
					}
				l81:
					goto l76
				l77:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
				}
				depth--
				add(Rulegeneric_name, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 22 Action0 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction0, position)
			}
			return true
		},
		/* 23 Action1 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction1, position)
			}
			return true
		},
		nil,
		/* 25 Action2 <- <{ p.Off(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction2, position)
			}
			return true
		},
		/* 26 Action3 <- <{ p.Lim(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction3, position)
			}
			return true
		},
		/* 27 Action4 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction4, position)
			}
			return true
		},
		/* 28 Action5 <- <{ p.Countall(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction5, position)
			}
			return true
		},
		/* 29 Action6 <- <{ p.Inter() }> */
		func() bool {
			{
				add(RuleAction6, position)
			}
			return true
		},
		/* 30 Action7 <- <{ p.Inter(); fmt.Printf("one attribute\n") }> */
		func() bool {
			{
				add(RuleAction7, position)
			}
			return true
		},
		/* 31 Action8 <- <{ fmt.Printf("no attributes\n") }> */
		func() bool {
			{
				add(RuleAction8, position)
			}
			return true
		},
		/* 32 Action9 <- <{ p.Attr(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction9, position)
			}
			return true
		},
		/* 33 Action10 <- <{ p.Union() }> */
		func() bool {
			{
				add(RuleAction10, position)
			}
			return true
		},
		/* 34 Action11 <- <{ p.Pa() }> */
		func() bool {
			{
				add(RuleAction11, position)
			}
			return true
		},
	}
	p.rules = rules
}
