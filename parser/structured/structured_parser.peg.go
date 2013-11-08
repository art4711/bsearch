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
	RuleOpType
	RuleName
	RuleValue
	RuleIntValue
	RuleStrValue
	Rulenumber
	Rules
	Rulegeneric_name
	Ruleopname
	RuleAction0
	RulePegText
	RuleAction1
	RuleAction2
	RuleAction3
	RuleAction4

	RulePre_
	Rule_In_
	Rule_Suf
)

var Rul3s = [...]string{
	"Unknown",
	"Query",
	"Operation",
	"OpType",
	"Name",
	"Value",
	"IntValue",
	"StrValue",
	"number",
	"s",
	"generic_name",
	"opname",
	"Action0",
	"PegText",
	"Action1",
	"Action2",
	"Action3",
	"Action4",

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

type Parser struct {
	opers.Query

	Buffer string
	buffer []rune
	rules  [18]func() bool
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
	p *Parser
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

func (p *Parser) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Parser) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *Parser) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.TokenTree.Tokens() {
		switch token.Rule {
		case RulePegText:
			begin, end = int(token.begin), int(token.end)
		case RuleAction0:
			p.OpEnd()
		case RuleAction1:
			p.OpStart(buffer[begin:end])
		case RuleAction2:
			p.OpName(buffer[begin:end])
		case RuleAction3:
			p.OpIntValue(buffer[begin:end])
		case RuleAction4:
			p.OpStrValue(buffer[begin:end])

		}
	}
}

func (p *Parser) Init() {
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
		/* 1 Operation <- <('(' OpType (s Name)? (s Value)* (s Operation)* ')' Action0)> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				if buffer[position] != rune('(') {
					goto l3
				}
				position++
				if !rules[RuleOpType]() {
					goto l3
				}
				{
					position5, tokenIndex5, depth5 := position, tokenIndex, depth
					if !rules[Rules]() {
						goto l5
					}
					if !rules[RuleName]() {
						goto l5
					}
					goto l6
				l5:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
				}
			l6:
			l7:
				{
					position8, tokenIndex8, depth8 := position, tokenIndex, depth
					if !rules[Rules]() {
						goto l8
					}
					if !rules[RuleValue]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex, depth = position8, tokenIndex8, depth8
				}
			l9:
				{
					position10, tokenIndex10, depth10 := position, tokenIndex, depth
					if !rules[Rules]() {
						goto l10
					}
					if !rules[RuleOperation]() {
						goto l10
					}
					goto l9
				l10:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
				}
				if buffer[position] != rune(')') {
					goto l3
				}
				position++
				if !rules[RuleAction0]() {
					goto l3
				}
				depth--
				add(RuleOperation, position4)
			}
			return true
		l3:
			position, tokenIndex, depth = position3, tokenIndex3, depth3
			return false
		},
		/* 2 OpType <- <(<opname> Action1)> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
				{
					position13 := position
					depth++
					if !rules[Ruleopname]() {
						goto l11
					}
					depth--
					add(RulePegText, position13)
				}
				if !rules[RuleAction1]() {
					goto l11
				}
				depth--
				add(RuleOpType, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 3 Name <- <('"' <generic_name> '"' Action2)> */
		func() bool {
			position14, tokenIndex14, depth14 := position, tokenIndex, depth
			{
				position15 := position
				depth++
				if buffer[position] != rune('"') {
					goto l14
				}
				position++
				{
					position16 := position
					depth++
					if !rules[Rulegeneric_name]() {
						goto l14
					}
					depth--
					add(RulePegText, position16)
				}
				if buffer[position] != rune('"') {
					goto l14
				}
				position++
				if !rules[RuleAction2]() {
					goto l14
				}
				depth--
				add(RuleName, position15)
			}
			return true
		l14:
			position, tokenIndex, depth = position14, tokenIndex14, depth14
			return false
		},
		/* 4 Value <- <(IntValue / StrValue)> */
		func() bool {
			position17, tokenIndex17, depth17 := position, tokenIndex, depth
			{
				position18 := position
				depth++
				{
					position19, tokenIndex19, depth19 := position, tokenIndex, depth
					if !rules[RuleIntValue]() {
						goto l20
					}
					goto l19
				l20:
					position, tokenIndex, depth = position19, tokenIndex19, depth19
					if !rules[RuleStrValue]() {
						goto l17
					}
				}
			l19:
				depth--
				add(RuleValue, position18)
			}
			return true
		l17:
			position, tokenIndex, depth = position17, tokenIndex17, depth17
			return false
		},
		/* 5 IntValue <- <('[' s <number> s ']' Action3)> */
		func() bool {
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
				if buffer[position] != rune('[') {
					goto l21
				}
				position++
				if !rules[Rules]() {
					goto l21
				}
				{
					position23 := position
					depth++
					if !rules[Rulenumber]() {
						goto l21
					}
					depth--
					add(RulePegText, position23)
				}
				if !rules[Rules]() {
					goto l21
				}
				if buffer[position] != rune(']') {
					goto l21
				}
				position++
				if !rules[RuleAction3]() {
					goto l21
				}
				depth--
				add(RuleIntValue, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 6 StrValue <- <('[' s <generic_name> s ']' Action4)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if buffer[position] != rune('[') {
					goto l24
				}
				position++
				if !rules[Rules]() {
					goto l24
				}
				{
					position26 := position
					depth++
					if !rules[Rulegeneric_name]() {
						goto l24
					}
					depth--
					add(RulePegText, position26)
				}
				if !rules[Rules]() {
					goto l24
				}
				if buffer[position] != rune(']') {
					goto l24
				}
				position++
				if !rules[RuleAction4]() {
					goto l24
				}
				depth--
				add(RuleStrValue, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 7 number <- <[0-9]+> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l27
				}
				position++
			l29:
				{
					position30, tokenIndex30, depth30 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l30
					}
					position++
					goto l29
				l30:
					position, tokenIndex, depth = position30, tokenIndex30, depth30
				}
				depth--
				add(Rulenumber, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 8 s <- <' '+> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l31
				}
				position++
			l33:
				{
					position34, tokenIndex34, depth34 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l34
					}
					position++
					goto l33
				l34:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
				}
				depth--
				add(Rules, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 9 generic_name <- <([a-z] / [A-Z] / [0-9] / '_' / ':')+> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				{
					position39, tokenIndex39, depth39 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l40
					}
					position++
					goto l39
				l40:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l41
					}
					position++
					goto l39
				l41:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l42
					}
					position++
					goto l39
				l42:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if buffer[position] != rune('_') {
						goto l43
					}
					position++
					goto l39
				l43:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if buffer[position] != rune(':') {
						goto l35
					}
					position++
				}
			l39:
			l37:
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					{
						position44, tokenIndex44, depth44 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l45
						}
						position++
						goto l44
					l45:
						position, tokenIndex, depth = position44, tokenIndex44, depth44
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l46
						}
						position++
						goto l44
					l46:
						position, tokenIndex, depth = position44, tokenIndex44, depth44
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l47
						}
						position++
						goto l44
					l47:
						position, tokenIndex, depth = position44, tokenIndex44, depth44
						if buffer[position] != rune('_') {
							goto l48
						}
						position++
						goto l44
					l48:
						position, tokenIndex, depth = position44, tokenIndex44, depth44
						if buffer[position] != rune(':') {
							goto l38
						}
						position++
					}
				l44:
					goto l37
				l38:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
				}
				depth--
				add(Rulegeneric_name, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 10 opname <- <([a-z] / '_')+> */
		func() bool {
			position49, tokenIndex49, depth49 := position, tokenIndex, depth
			{
				position50 := position
				depth++
				{
					position53, tokenIndex53, depth53 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l54
					}
					position++
					goto l53
				l54:
					position, tokenIndex, depth = position53, tokenIndex53, depth53
					if buffer[position] != rune('_') {
						goto l49
					}
					position++
				}
			l53:
			l51:
				{
					position52, tokenIndex52, depth52 := position, tokenIndex, depth
					{
						position55, tokenIndex55, depth55 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l56
						}
						position++
						goto l55
					l56:
						position, tokenIndex, depth = position55, tokenIndex55, depth55
						if buffer[position] != rune('_') {
							goto l52
						}
						position++
					}
				l55:
					goto l51
				l52:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
				}
				depth--
				add(Ruleopname, position50)
			}
			return true
		l49:
			position, tokenIndex, depth = position49, tokenIndex49, depth49
			return false
		},
		/* 12 Action0 <- <{ p.OpEnd() }> */
		func() bool {
			{
				add(RuleAction0, position)
			}
			return true
		},
		nil,
		/* 14 Action1 <- <{ p.OpStart(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction1, position)
			}
			return true
		},
		/* 15 Action2 <- <{ p.OpName(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction2, position)
			}
			return true
		},
		/* 16 Action3 <- <{ p.OpIntValue(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction3, position)
			}
			return true
		},
		/* 17 Action4 <- <{ p.OpStrValue(buffer[begin:end]) }> */
		func() bool {
			{
				add(RuleAction4, position)
			}
			return true
		},
	}
	p.rules = rules
}
