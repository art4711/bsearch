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

type QueryParser struct {
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
	p *QueryParser
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

func (p *QueryParser) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *QueryParser) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *QueryParser) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.TokenTree.Tokens() {
		switch token.Rule {
		case RulePegText:
			begin, end = int(token.begin), int(token.end)
		case RuleAction0:
			p.PopAdd()
		case RuleAction1:
			p.PopAdd()
		case RuleAction2:
			p.Offset(buffer[begin:end])
		case RuleAction3:
			p.Limit(buffer[begin:end])
		case RuleAction4:
			p.PopAdd()
		case RuleAction5:
			p.CountAll(buffer[begin:end])
		case RuleAction6:
			p.StartIntersection()
		case RuleAction7:
			p.StartIntersection()
			fmt.Printf("one attribute\n")
		case RuleAction8:
			fmt.Printf("no attributes\n")
		case RuleAction9:
			p.Attr(buffer[begin:end])
		case RuleAction10:
			p.StartUnion()
		case RuleAction11:
			p.PopAdd()

		}
	}
}

func (p *QueryParser) Init() {
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
				{
					position2 := position
					depth++
					{
						position3 := position
						depth++
						{
							position4, tokenIndex4, depth4 := position, tokenIndex, depth
							{
								position6 := position
								depth++
								{
									position7 := position
									depth++
									if !rules[Rulenumber]() {
										goto l5
									}
									depth--
									add(RulePegText, position7)
								}
								if !rules[Rules]() {
									goto l5
								}
								{
									add(RuleAction2, position)
								}
								depth--
								add(RuleOffset, position6)
							}
							if !rules[RuleLimQuery]() {
								goto l5
							}
							{
								add(RuleAction0, position)
							}
							goto l4
						l5:
							position, tokenIndex, depth = position4, tokenIndex4, depth4
							if !rules[RuleLimQuery]() {
								goto l0
							}
						}
					l4:
						depth--
						add(RuleOffLimQuery, position3)
					}
					depth--
					add(RuleResFiltQuery, position2)
				}
				{
					position10, tokenIndex10, depth10 := position, tokenIndex, depth
					if !matchDot() {
						goto l10
					}
					goto l0
				l10:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
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
		nil,
		/* 2 OffLimQuery <- <((Offset LimQuery Action0) / LimQuery)> */
		nil,
		/* 3 LimQuery <- <((Limit Q3 Action1) / Q3)> */
		func() bool {
			position13, tokenIndex13, depth13 := position, tokenIndex, depth
			{
				position14 := position
				depth++
				{
					position15, tokenIndex15, depth15 := position, tokenIndex, depth
					{
						position17 := position
						depth++
						if buffer[position] != rune('l') {
							goto l16
						}
						position++
						if buffer[position] != rune('i') {
							goto l16
						}
						position++
						if buffer[position] != rune('m') {
							goto l16
						}
						position++
						if buffer[position] != rune(':') {
							goto l16
						}
						position++
						{
							position18 := position
							depth++
							if !rules[Rulenumber]() {
								goto l16
							}
							depth--
							add(RulePegText, position18)
						}
						if !rules[Rules]() {
							goto l16
						}
						{
							add(RuleAction3, position)
						}
						depth--
						add(RuleLimit, position17)
					}
					if !rules[RuleQ3]() {
						goto l16
					}
					{
						add(RuleAction1, position)
					}
					goto l15
				l16:
					position, tokenIndex, depth = position15, tokenIndex15, depth15
					if !rules[RuleQ3]() {
						goto l13
					}
				}
			l15:
				depth--
				add(RuleLimQuery, position14)
			}
			return true
		l13:
			position, tokenIndex, depth = position13, tokenIndex13, depth13
			return false
		},
		/* 4 Q3 <- <Params?> */
		func() bool {
			{
				position22 := position
				depth++
				{
					position23, tokenIndex23, depth23 := position, tokenIndex, depth
					{
						position25 := position
						depth++
						{
							position26, tokenIndex26, depth26 := position, tokenIndex, depth
							{
								position28 := position
								depth++
								{
									position29, tokenIndex29, depth29 := position, tokenIndex, depth
									{
										position31 := position
										depth++
										if buffer[position] != rune('c') {
											goto l30
										}
										position++
										if buffer[position] != rune('o') {
											goto l30
										}
										position++
										if buffer[position] != rune('u') {
											goto l30
										}
										position++
										if buffer[position] != rune('n') {
											goto l30
										}
										position++
										if buffer[position] != rune('t') {
											goto l30
										}
										position++
										if buffer[position] != rune('_') {
											goto l30
										}
										position++
										if buffer[position] != rune('a') {
											goto l30
										}
										position++
										if buffer[position] != rune('l') {
											goto l30
										}
										position++
										if buffer[position] != rune('l') {
											goto l30
										}
										position++
										if buffer[position] != rune('(') {
											goto l30
										}
										position++
										{
											position32 := position
											depth++
											{
												position33 := position
												depth++
												if !rules[Rulegeneric_name]() {
													goto l30
												}
												depth--
												add(Rulecounter_name, position33)
											}
											depth--
											add(RulePegText, position32)
										}
										if buffer[position] != rune(')') {
											goto l30
										}
										position++
										if !rules[Rules]() {
											goto l30
										}
										{
											add(RuleAction5, position)
										}
										depth--
										add(RuleCountAll, position31)
									}
									if !rules[RuleAttrs]() {
										goto l30
									}
									{
										add(RuleAction4, position)
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
							goto l26
						l27:
							position, tokenIndex, depth = position26, tokenIndex26, depth26
							if !rules[RuleAttrs]() {
								goto l23
							}
						}
					l26:
						depth--
						add(RuleParams, position25)
					}
					goto l24
				l23:
					position, tokenIndex, depth = position23, tokenIndex23, depth23
				}
			l24:
				depth--
				add(RuleQ3, position22)
			}
			return true
		},
		/* 5 Offset <- <(<number> s Action2)> */
		nil,
		/* 6 Limit <- <('l' 'i' 'm' ':' <number> s Action3)> */
		nil,
		/* 7 Params <- <(CountAllAttrs / Attrs)> */
		nil,
		/* 8 CountAllAttrs <- <((CountAll Attrs Action4) / Attrs)> */
		nil,
		/* 9 CountAll <- <('c' 'o' 'u' 'n' 't' '_' 'a' 'l' 'l' '(' <counter_name> ')' s Action5)> */
		nil,
		/* 10 Attrs <- <((Action6 (Attr s)+ Attr s?) / (Action7 Attr s?) / (s? Action8))> */
		func() bool {
			{
				position42 := position
				depth++
				{
					position43, tokenIndex43, depth43 := position, tokenIndex, depth
					{
						add(RuleAction6, position)
					}
					if !rules[RuleAttr]() {
						goto l44
					}
					if !rules[Rules]() {
						goto l44
					}
				l46:
					{
						position47, tokenIndex47, depth47 := position, tokenIndex, depth
						if !rules[RuleAttr]() {
							goto l47
						}
						if !rules[Rules]() {
							goto l47
						}
						goto l46
					l47:
						position, tokenIndex, depth = position47, tokenIndex47, depth47
					}
					if !rules[RuleAttr]() {
						goto l44
					}
					{
						position48, tokenIndex48, depth48 := position, tokenIndex, depth
						if !rules[Rules]() {
							goto l48
						}
						goto l49
					l48:
						position, tokenIndex, depth = position48, tokenIndex48, depth48
					}
				l49:
					goto l43
				l44:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					{
						add(RuleAction7, position)
					}
					if !rules[RuleAttr]() {
						goto l50
					}
					{
						position52, tokenIndex52, depth52 := position, tokenIndex, depth
						if !rules[Rules]() {
							goto l52
						}
						goto l53
					l52:
						position, tokenIndex, depth = position52, tokenIndex52, depth52
					}
				l53:
					goto l43
				l50:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					{
						position54, tokenIndex54, depth54 := position, tokenIndex, depth
						if !rules[Rules]() {
							goto l54
						}
						goto l55
					l54:
						position, tokenIndex, depth = position54, tokenIndex54, depth54
					}
				l55:
					{
						add(RuleAction8, position)
					}
				}
			l43:
				depth--
				add(RuleAttrs, position42)
			}
			return true
		},
		/* 11 Attr <- <(AttrUnion / Attribute)> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
				{
					position59, tokenIndex59, depth59 := position, tokenIndex, depth
					{
						position61 := position
						depth++
						{
							add(RuleAction10, position)
						}
						{
							position63 := position
							depth++
							if !rules[RuleAttribute]() {
								goto l60
							}
							if !rules[Rules]() {
								goto l60
							}
							if buffer[position] != rune('O') {
								goto l60
							}
							position++
							if buffer[position] != rune('R') {
								goto l60
							}
							position++
							if !rules[Rules]() {
								goto l60
							}
							if !rules[RuleAttribute]() {
								goto l60
							}
						l64:
							{
								position65, tokenIndex65, depth65 := position, tokenIndex, depth
								if !rules[Rules]() {
									goto l65
								}
								if buffer[position] != rune('O') {
									goto l65
								}
								position++
								if buffer[position] != rune('R') {
									goto l65
								}
								position++
								if !rules[Rules]() {
									goto l65
								}
								if !rules[RuleAttribute]() {
									goto l65
								}
								goto l64
							l65:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
							}
							depth--
							add(RuleAttributeORList, position63)
						}
						{
							add(RuleAction11, position)
						}
						depth--
						add(RuleAttrUnion, position61)
					}
					goto l59
				l60:
					position, tokenIndex, depth = position59, tokenIndex59, depth59
					if !rules[RuleAttribute]() {
						goto l57
					}
				}
			l59:
				depth--
				add(RuleAttr, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 12 Attribute <- <(<(attr_name ':' attr_value)> Action9)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				{
					position69 := position
					depth++
					{
						position70 := position
						depth++
						if !rules[Rulegeneric_name]() {
							goto l67
						}
						depth--
						add(Ruleattr_name, position70)
					}
					if buffer[position] != rune(':') {
						goto l67
					}
					position++
					{
						position71 := position
						depth++
						if !rules[Rulegeneric_name]() {
							goto l67
						}
						depth--
						add(Ruleattr_value, position71)
					}
					depth--
					add(RulePegText, position69)
				}
				{
					add(RuleAction9, position)
				}
				depth--
				add(RuleAttribute, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 13 AttrUnion <- <(Action10 AttributeORList Action11)> */
		nil,
		/* 14 AttributeORList <- <(Attribute (s ('O' 'R') s Attribute)+)> */
		nil,
		/* 15 number <- <[0-9]+> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l75
				}
				position++
			l77:
				{
					position78, tokenIndex78, depth78 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l78
					}
					position++
					goto l77
				l78:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
				}
				depth--
				add(Rulenumber, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 16 s <- <' '+> */
		func() bool {
			position79, tokenIndex79, depth79 := position, tokenIndex, depth
			{
				position80 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l79
				}
				position++
			l81:
				{
					position82, tokenIndex82, depth82 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l82
					}
					position++
					goto l81
				l82:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
				}
				depth--
				add(Rules, position80)
			}
			return true
		l79:
			position, tokenIndex, depth = position79, tokenIndex79, depth79
			return false
		},
		/* 17 counter_name <- <generic_name> */
		nil,
		/* 18 attr_name <- <generic_name> */
		nil,
		/* 19 attr_value <- <generic_name> */
		nil,
		/* 20 generic_name <- <((&('_') '_') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l86
						}
						position++
						break
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l86
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l86
						}
						position++
						break
					}
				}

			l88:
				{
					position89, tokenIndex89, depth89 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l89
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l89
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l89
							}
							position++
							break
						}
					}

					goto l88
				l89:
					position, tokenIndex, depth = position89, tokenIndex89, depth89
				}
				depth--
				add(Rulegeneric_name, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 22 Action0 <- <{ p.PopAdd() }> */
		nil,
		/* 23 Action1 <- <{ p.PopAdd() }> */
		nil,
		nil,
		/* 25 Action2 <- <{ p.Offset(buffer[begin:end]) }> */
		nil,
		/* 26 Action3 <- <{ p.Limit(buffer[begin:end]) }> */
		nil,
		/* 27 Action4 <- <{ p.PopAdd() }> */
		nil,
		/* 28 Action5 <- <{ p.CountAll(buffer[begin:end]) }> */
		nil,
		/* 29 Action6 <- <{ p.StartIntersection() }> */
		nil,
		/* 30 Action7 <- <{ p.StartIntersection(); fmt.Printf("one attribute\n") }> */
		nil,
		/* 31 Action8 <- <{ fmt.Printf("no attributes\n") }> */
		nil,
		/* 32 Action9 <- <{ p.Attr(buffer[begin:end]) }> */
		nil,
		/* 33 Action10 <- <{ p.StartUnion() }> */
		nil,
		/* 34 Action11 <- <{ p.PopAdd() }> */
		nil,
	}
	p.rules = rules
}
