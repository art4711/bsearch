package structured

import (
	/*"bytes"*/
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
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
	TokenTree

	tokenIndex int
	depth      int
	position   int
	tree       TokenTree
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

func (p *Parser) Parse(rule ...int) error {
	r := 1
	if len(rule) > 0 {
		r = rule[0]
	}
	matches := p.rules[r]()
	p.TokenTree = p.tree
	if matches {
		p.TokenTree.trim(p.tokenIndex)
		return nil
	}
	return &parseError{p}
}

func (p *Parser) Reset() {
	p.position, p.tokenIndex, p.depth = 0, 0, 0
}

func (p *Parser) add(rule Rule, begin int) {
	if t := p.tree.Expand(p.tokenIndex); t != nil {
		p.tree = t
	}
	p.tree.Add(rule, begin, p.position, p.depth, p.tokenIndex)
	p.tokenIndex++
}

func (p *Parser) matchDot() bool {
	if p.buffer[p.position] != END_SYMBOL {
		p.position++
		return true
	}
	return false
}

func (p *Parser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != END_SYMBOL {
		p.buffer = append(p.buffer, END_SYMBOL)
	}

	p.tree = &tokens16{tree: make([]token16, math.MaxInt16)}
	p.tokenIndex = 0

	/*matchChar := func(c byte) bool {
		if buffer[p.position] == c {
			p.position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[p.position]; c >= lower && c <= upper {
			p.position++
			return true
		}
		return false
	}*/

	p.rules = [...]func() bool{
		nil,
		p.XRuleQuery,
		p.XRuleOperation,
		p.XRuleOpType,
		p.XRuleName,
		p.XRuleValue,
		p.XRuleIntValue,
		p.XRuleStrValue,
		p.XRulenumber,
		p.XRules,
		p.XRulegeneric_name,
		p.XRuleopname,
		p.XRuleAction0,
		nil,
		p.XRuleAction1,
		p.XRuleAction2,
		p.XRuleAction3,
		p.XRuleAction4,
	}
}

/* 0 Query <- <(Operation !.)> */
func (p *Parser) XRuleQuery() bool {
	position0, tokenIndex0, depth0 := p.position, p.tokenIndex, p.depth
	{
		position1 := p.position
		p.depth++
		if !p.XRuleOperation() {
			goto l0
		}
		{
			position2, tokenIndex2, depth2 := p.position, p.tokenIndex, p.depth
			if !p.matchDot() {
				goto l2
			}
			goto l0
		l2:
			p.position, p.tokenIndex, p.depth = position2, tokenIndex2, depth2
		}
		p.depth--
		p.add(RuleQuery, position1)
	}
	return true
l0:
	p.position, p.tokenIndex, p.depth = position0, tokenIndex0, depth0
	return false
}

/* 1 Operation <- <('(' OpType (s Name)? (s Value)* (s Operation)* ')' Action0)> */
func (p *Parser) XRuleOperation() bool {
	position3, tokenIndex3, depth3 := p.position, p.tokenIndex, p.depth
	{
		position4 := p.position
		p.depth++
		if p.buffer[p.position] != rune('(') {
			goto l3
		}
		p.position++
		if !p.XRuleOpType() {
			goto l3
		}
		{
			position5, tokenIndex5, depth5 := p.position, p.tokenIndex, p.depth
			if !p.XRules() {
				goto l5
			}
			if !p.XRuleName() {
				goto l5
			}
			goto l6
		l5:
			p.position, p.tokenIndex, p.depth = position5, tokenIndex5, depth5
		}
	l6:
	l7:
		{
			position8, tokenIndex8, depth8 := p.position, p.tokenIndex, p.depth
			if !p.XRules() {
				goto l8
			}
			if !p.XRuleValue() {
				goto l8
			}
			goto l7
		l8:
			p.position, p.tokenIndex, p.depth = position8, tokenIndex8, depth8
		}
	l9:
		{
			position10, tokenIndex10, depth10 := p.position, p.tokenIndex, p.depth
			if !p.XRules() {
				goto l10
			}
			if !p.XRuleOperation() {
				goto l10
			}
			goto l9
		l10:
			p.position, p.tokenIndex, p.depth = position10, tokenIndex10, depth10
		}
		if p.buffer[p.position] != rune(')') {
			goto l3
		}
		p.position++
		if !p.XRuleAction0() {
			goto l3
		}
		p.depth--
		p.add(RuleOperation, position4)
	}
	return true
l3:
	p.position, p.tokenIndex, p.depth = position3, tokenIndex3, depth3
	return false
}

/* 2 OpType <- <(<opname> Action1)> */
func (p *Parser) XRuleOpType() bool {
	position11, tokenIndex11, depth11 := p.position, p.tokenIndex, p.depth
	{
		position12 := p.position
		p.depth++
		{
			position13 := p.position
			p.depth++
			if !p.XRuleopname() {
				goto l11
			}
			p.depth--
			p.add(RulePegText, position13)
		}
		if !p.XRuleAction1() {
			goto l11
		}
		p.depth--
		p.add(RuleOpType, position12)
	}
	return true
l11:
	p.position, p.tokenIndex, p.depth = position11, tokenIndex11, depth11
	return false
}

/* 3 Name <- <('"' <generic_name> '"' Action2)> */
func (p *Parser) XRuleName() bool {
	position14, tokenIndex14, depth14 := p.position, p.tokenIndex, p.depth
	{
		position15 := p.position
		p.depth++
		if p.buffer[p.position] != rune('"') {
			goto l14
		}
		p.position++
		{
			position16 := p.position
			p.depth++
			if !p.XRulegeneric_name() {
				goto l14
			}
			p.depth--
			p.add(RulePegText, position16)
		}
		if p.buffer[p.position] != rune('"') {
			goto l14
		}
		p.position++
		if !p.XRuleAction2() {
			goto l14
		}
		p.depth--
		p.add(RuleName, position15)
	}
	return true
l14:
	p.position, p.tokenIndex, p.depth = position14, tokenIndex14, depth14
	return false
}

/* 4 Value <- <(IntValue / StrValue)> */
func (p *Parser) XRuleValue() bool {
	position17, tokenIndex17, depth17 := p.position, p.tokenIndex, p.depth
	{
		position18 := p.position
		p.depth++
		{
			position19, tokenIndex19, depth19 := p.position, p.tokenIndex, p.depth
			if !p.XRuleIntValue() {
				goto l20
			}
			goto l19
		l20:
			p.position, p.tokenIndex, p.depth = position19, tokenIndex19, depth19
			if !p.XRuleStrValue() {
				goto l17
			}
		}
	l19:
		p.depth--
		p.add(RuleValue, position18)
	}
	return true
l17:
	p.position, p.tokenIndex, p.depth = position17, tokenIndex17, depth17
	return false
}

/* 5 IntValue <- <('[' s <number> s ']' Action3)> */
func (p *Parser) XRuleIntValue() bool {
	position21, tokenIndex21, depth21 := p.position, p.tokenIndex, p.depth
	{
		position22 := p.position
		p.depth++
		if p.buffer[p.position] != rune('[') {
			goto l21
		}
		p.position++
		if !p.XRules() {
			goto l21
		}
		{
			position23 := p.position
			p.depth++
			if !p.XRulenumber() {
				goto l21
			}
			p.depth--
			p.add(RulePegText, position23)
		}
		if !p.XRules() {
			goto l21
		}
		if p.buffer[p.position] != rune(']') {
			goto l21
		}
		p.position++
		if !p.XRuleAction3() {
			goto l21
		}
		p.depth--
		p.add(RuleIntValue, position22)
	}
	return true
l21:
	p.position, p.tokenIndex, p.depth = position21, tokenIndex21, depth21
	return false
}

/* 6 StrValue <- <('[' s <generic_name> s ']' Action4)> */
func (p *Parser) XRuleStrValue() bool {
	position24, tokenIndex24, depth24 := p.position, p.tokenIndex, p.depth
	{
		position25 := p.position
		p.depth++
		if p.buffer[p.position] != rune('[') {
			goto l24
		}
		p.position++
		if !p.XRules() {
			goto l24
		}
		{
			position26 := p.position
			p.depth++
			if !p.XRulegeneric_name() {
				goto l24
			}
			p.depth--
			p.add(RulePegText, position26)
		}
		if !p.XRules() {
			goto l24
		}
		if p.buffer[p.position] != rune(']') {
			goto l24
		}
		p.position++
		if !p.XRuleAction4() {
			goto l24
		}
		p.depth--
		p.add(RuleStrValue, position25)
	}
	return true
l24:
	p.position, p.tokenIndex, p.depth = position24, tokenIndex24, depth24
	return false
}

/* 7 number <- <[0-9]+> */
func (p *Parser) XRulenumber() bool {
	position27, tokenIndex27, depth27 := p.position, p.tokenIndex, p.depth
	{
		position28 := p.position
		p.depth++
		if c := p.buffer[p.position]; c < rune('0') || c > rune('9') {
			goto l27
		}
		p.position++
	l29:
		{
			position30, tokenIndex30, depth30 := p.position, p.tokenIndex, p.depth
			if c := p.buffer[p.position]; c < rune('0') || c > rune('9') {
				goto l30
			}
			p.position++
			goto l29
		l30:
			p.position, p.tokenIndex, p.depth = position30, tokenIndex30, depth30
		}
		p.depth--
		p.add(Rulenumber, position28)
	}
	return true
l27:
	p.position, p.tokenIndex, p.depth = position27, tokenIndex27, depth27
	return false
}

/* 8 s <- <' '+> */
func (p *Parser) XRules() bool {
	position31, tokenIndex31, depth31 := p.position, p.tokenIndex, p.depth
	{
		position32 := p.position
		p.depth++
		if p.buffer[p.position] != rune(' ') {
			goto l31
		}
		p.position++
	l33:
		{
			position34, tokenIndex34, depth34 := p.position, p.tokenIndex, p.depth
			if p.buffer[p.position] != rune(' ') {
				goto l34
			}
			p.position++
			goto l33
		l34:
			p.position, p.tokenIndex, p.depth = position34, tokenIndex34, depth34
		}
		p.depth--
		p.add(Rules, position32)
	}
	return true
l31:
	p.position, p.tokenIndex, p.depth = position31, tokenIndex31, depth31
	return false
}

/* 9 generic_name <- <((&(':') ':') | (&('_') '_') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> */
func (p *Parser) XRulegeneric_name() bool {
	position35, tokenIndex35, depth35 := p.position, p.tokenIndex, p.depth
	{
		position36 := p.position
		p.depth++
		{
			if strings.ContainsRune(":_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", p.buffer[p.position]) {
				p.position++
			} else {
				goto l35
			}
		}

	l37:
		{
			position38, tokenIndex38, depth38 := p.position, p.tokenIndex, p.depth
			{
				if strings.ContainsRune(":_0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", p.buffer[p.position]) {
					p.position++
				} else {
					goto l38
				}
			}

			goto l37
		l38:
			p.position, p.tokenIndex, p.depth = position38, tokenIndex38, depth38
		}
		p.depth--
		p.add(Rulegeneric_name, position36)
	}
	return true
l35:
	p.position, p.tokenIndex, p.depth = position35, tokenIndex35, depth35
	return false
}

/* 10 opname <- <([a-z] / '_')+> */
func (p *Parser) XRuleopname() bool {
	position41, tokenIndex41, depth41 := p.position, p.tokenIndex, p.depth
	{
		position42 := p.position
		p.depth++
		{
			position45, tokenIndex45, depth45 := p.position, p.tokenIndex, p.depth
			if c := p.buffer[p.position]; c < rune('a') || c > rune('z') {
				goto l46
			}
			p.position++
			goto l45
		l46:
			p.position, p.tokenIndex, p.depth = position45, tokenIndex45, depth45
			if p.buffer[p.position] != rune('_') {
				goto l41
			}
			p.position++
		}
	l45:
	l43:
		{
			position44, tokenIndex44, depth44 := p.position, p.tokenIndex, p.depth
			{
				position47, tokenIndex47, depth47 := p.position, p.tokenIndex, p.depth
				if c := p.buffer[p.position]; c < rune('a') || c > rune('z') {
					goto l48
				}
				p.position++
				goto l47
			l48:
				p.position, p.tokenIndex, p.depth = position47, tokenIndex47, depth47
				if p.buffer[p.position] != rune('_') {
					goto l44
				}
				p.position++
			}
		l47:
			goto l43
		l44:
			p.position, p.tokenIndex, p.depth = position44, tokenIndex44, depth44
		}
		p.depth--
		p.add(Ruleopname, position42)
	}
	return true
l41:
	p.position, p.tokenIndex, p.depth = position41, tokenIndex41, depth41
	return false
}

/* 12 Action0 <- <{ p.OpEnd() }> */
func (p *Parser) XRuleAction0() bool {
	{
		p.add(RuleAction0, p.position)
	}
	return true
}

/* 14 Action1 <- <{ p.OpStart(buffer[begin:end]) }> */
func (p *Parser) XRuleAction1() bool {
	{
		p.add(RuleAction1, p.position)
	}
	return true
}

/* 15 Action2 <- <{ p.OpName(buffer[begin:end]) }> */
func (p *Parser) XRuleAction2() bool {
	{
		p.add(RuleAction2, p.position)
	}
	return true
}

/* 16 Action3 <- <{ p.OpIntValue(buffer[begin:end]) }> */
func (p *Parser) XRuleAction3() bool {
	{
		p.add(RuleAction3, p.position)
	}
	return true
}

/* 17 Action4 <- <{ p.OpStrValue(buffer[begin:end]) }> */
func (p *Parser) XRuleAction4() bool {
	{
		p.add(RuleAction4, p.position)
	}
	return true
}
