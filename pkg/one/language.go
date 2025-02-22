package one

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
	treeSitterPy "github.com/smacker/go-tree-sitter/python"
	treeSitterTsx "github.com/smacker/go-tree-sitter/typescript/tsx"
	treeSitterTs "github.com/smacker/go-tree-sitter/typescript/typescript"
)

// ParseResult is the result of parsing a file.
type ParseResult struct {
	// Ast is the root node of the tree-sitter parse-tree
	// representing this file
	Ast *sitter.Node
	// Source is the raw source code of the file
	Source []byte
	// FilePath is the path to the file that was parsed
	FilePath string
	// Language is the tree-sitter language used to parse the file
	TsLanguage *sitter.Language
	// Language is the language of the file
	Language Language
	// ScopeTree represents the scope hierarchy of the file.
	// Can be nil if scope support for this language has not been implemented yet.
	ScopeTree *ScopeTree
}

type Language int

const (
	LangUnknown Language = iota
	LangPy
	LangJs  // vanilla JS and JSX
	LangTs  // TypeScript (not TSX)
	LangTsx // TypeScript with JSX extension
)

// tsGrammarForLang returns the tree-sitter grammar for the given language.
// May return `nil` when `lang` is `LangUnkown`.
func (lang Language) Grammar() *sitter.Language {
	switch lang {
	case LangPy:
		return treeSitterPy.GetLanguage()
	case LangJs:
		return treeSitterTsx.GetLanguage()
	case LangTs:
		return treeSitterTs.GetLanguage()
	case LangTsx:
		return treeSitterTsx.GetLanguage()
	default:
		return nil
	}
}

// NOTE(@injuly): TypeScript and TSX have to parsed with DIFFERENT
// grammars. Otherwise, because an expression like `<Foo>bar` is
// parsed as a (legacy) type-cast in TS, but a JSXElement in TSX.
// See: https://facebook.github.io/jsx/#prod-JSXElement

// LanguageFromFilePath returns the Language of the file at the given path
// returns `LangUnkown` if the language is not recognized (e.g: `.txt` files).
func LanguageFromFilePath(path string) Language {
	ext := filepath.Ext(path)
	switch ext {
	case ".py":
		return LangPy
		// TODO: .jsx and .js can both have JSX syntax -_-
	case ".js", ".jsx":
		return LangJs
	case ".ts":
		return LangTs
	case ".tsx":
		return LangTsx
	default:
		return LangUnknown
	}
}

func Parse(filePath string, source []byte, language Language, grammar *sitter.Language) (*ParseResult, error) {
	ast, err := sitter.ParseCtx(context.Background(), source, grammar)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	scopeTree := MakeScopeTree(language, ast, source)
	parseResult := &ParseResult{
		Ast:        ast,
		Source:     source,
		FilePath:   filePath,
		TsLanguage: grammar,
		Language:   language,
		ScopeTree:  scopeTree,
	}

	return parseResult, nil
}

// ParseFile parses the file at the given path using the appropriate
// tree-sitter grammar.
func ParseFile(filePath string) (*ParseResult, error) {
	lang := LanguageFromFilePath(filePath)
	grammar := lang.Grammar()
	if grammar == nil {
		return nil, fmt.Errorf("unsupported file type: %s", filePath)
	}

	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return Parse(filePath, source, lang, grammar)
}
