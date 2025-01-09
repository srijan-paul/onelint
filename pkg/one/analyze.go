package one

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type Issue struct {
	// The message to display to the user
	Message string
	// The range of the issue in the source code
	Range sitter.Range
	// (optional) The AST node that caused the issue 
	Node *sitter.Node
}

type Analyzer struct {
	Language    Language
	ParseResult *ParseResult
	// rules is a list of all rules that should be applied to the AST
	// for this language.
	rules []Rule
	// entryRules maps node types to the rules that should be applied
	// when entering that node.
	entryRulesForNode map[string][]Rule
	// exitRules maps node types to the rules that should be applied
	// when leaving that node.
	exitRulesForNode map[string][]Rule
	issuesRaised     []*Issue
}

func FromFile(filePath string, baseRules []Rule) (*Analyzer, error) {
	res, err := ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	return NewAnalyzer(res, baseRules), nil
}

func NewAnalyzer(file *ParseResult, rules []Rule) *Analyzer {
	ana := &Analyzer{
		ParseResult:       file,
		Language:          file.Language,
		entryRulesForNode: map[string][]Rule{},
		exitRulesForNode:  map[string][]Rule{},
	}

	for _, rule := range rules {
		ana.AddRule(rule)
	}

	return ana
}

func (ana *Analyzer) Analyze() []*Issue {
	WalkTree(ana.ParseResult.Ast, ana)
	return ana.issuesRaised
}

func (ana *Analyzer) AddRule(rule Rule) {
	ana.rules = append(ana.rules, rule)
	typ := rule.NodeType()

	if rule.OnEnter() != nil {
		ana.entryRulesForNode[typ] = append(ana.entryRulesForNode[typ], rule)
	}

	if rule.OnLeave() != nil {
		ana.exitRulesForNode[typ] = append(ana.exitRulesForNode[typ], rule)
	}
}

func (ana *Analyzer) OnEnterNode(node *sitter.Node) bool {
	nodeType := node.Type()
	rules := ana.entryRulesForNode[nodeType]
	for _, rule := range rules {
		visitFn := rule.OnEnter()
		if visitFn != nil {
			(*visitFn)(rule, ana, node)
		}
	}
	return true
}

func (ana *Analyzer) OnLeaveNode(node *sitter.Node) {
	nodeType := node.Type()
	rules := ana.exitRulesForNode[nodeType]
	for _, rule := range rules {
		visitFn := rule.OnLeave()
		if visitFn != nil {
			(*visitFn)(rule, ana, node)
		}
	}
}

func (ana *Analyzer) Report(issue *Issue) {
	ana.issuesRaised = append(ana.issuesRaised, issue)
}