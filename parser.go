package main

import (
	"fmt"
)

type AST struct {
	Configuration map[string]string
	Tables        []*TabelAST
}

type TabelAST struct {
	Name       string
	Colmuns    []*ColmunAST
	References []*ReferenceAST
}

type ColmunAST struct {
	Name       string
	Data_type  string
	Attributes *AttributesAST
}

type ReferenceAST struct {
	TargetTable string
	TargetCol   string
	SourceCol   string
	OnDelete    string
	OnUpdate    string
}

type AttributesAST map[string]*AttributeAST

type AttributeAST struct {
	Name   string
	Values []*AttributeArgAST
}

type AttributeArgAST struct {
	Value string
	Type  string
}

var tokenizer *Tokenizer
var parseAttrFuncMap map[string]func(*Token, []*AttributeArgAST, *ColmunAST) error
var configurable []string
var ast *AST
var currentTableAst *TabelAST

func Parse(localTokenizer *Tokenizer) (*AST, error) {
	//init
	tokenizer = localTokenizer
	parseAttrFuncMap = map[string]func(*Token, []*AttributeArgAST, *ColmunAST) error{
		"id":             parseIdAttr,
		"default":        parseDefaultAttr,
		"auto_increment": parseAutoIncrementAttr,
		"nullable":       parseNullableAttr,
		"reference":      parseReferenceAttr,
		"onDelete":       parseOnDeleteAttr,
		"onUpdate":       parseOnUpdateAttr,
	}

	configurable = []string{"provider", "url"}

	tok := tokenizer.NextToken()
	ast = &AST{map[string]string{"provider": "sqlite"}, []*TabelAST{}}

	for tok.TokenType != T_EOF {
		if tok.TokenType == T_TABLE {
			tableDefAst, err := parseTable()
			if err != nil {
				return nil, err
			}
			ast.Tables = append(ast.Tables, tableDefAst)
		} else if tok.TokenType == T_SET {
			err := parseSet()
			if err != nil {
				return nil, err
			}
		} else if tok.TokenType != T_EOL {
			return nil, createError(fmt.Sprintf("Unexpected token '%s'", tok.Literal), tok.Line, tok.Col)
		}
		tok = tokenizer.NextToken()
	}

	return ast, nil
}

func parseSet() error {
	tok := tokenizer.NextToken()
	if tok.TokenType == T_IDEN {

		if !isConfigurable(tok.Literal) {
			return createError(fmt.Sprintf("Unknown '%s'", tok.Literal), tok.Line, tok.Col)
		}

		configurable := tok.Literal

		tok = tokenizer.NextToken()

		if tok.TokenType == T_EOL {
			return createError(
				fmt.Sprintf("Expected value after '%s'", configurable),
				tok.Line,
				tok.Col,
			)
		} else if tok.TokenType != T_IDEN && tok.TokenType != T_STRING {
			return createError(
				fmt.Sprintf("Unexpected token '%s'", tok.Literal),
				tok.Line,
				tok.Col,
			)
		}

		ast.Configuration[configurable] = tok.Literal
		tokenizer.NextToken()
	} else {
		return createError(fmt.Sprintf("Expected identifier after 'set'"), tok.Line, tok.Col)
	}

	return nil
}

func isConfigurable(literal string) bool {
	for _, o := range configurable {
		if o == literal {
			return true
		}
	}
	return false
}

func parseTable() (*TabelAST, error) {
	// table <TableName>
	tok := tokenizer.NextToken()
	if tok.TokenType != T_IDEN {
		if tok.TokenType == T_EOL || tok.TokenType == T_EOF {
			return nil, createError(
				"Missing '<Table Name>' after 'table'",
				tok.Line,
				tok.Col,
			)
		} else {
			return nil, createError(
				"'<Table Name>' must start with a letter or underscore",
				tok.Line,
				tok.Col,
			)
		}
	}

	tableAst := &TabelAST{tok.Literal, []*ColmunAST{}, []*ReferenceAST{}}
	currentTableAst = tableAst

	exists, _ := getTableByName(tok.Literal)
	if exists {
		return nil, createError(
			fmt.Sprintf("Table with name '%s' already declared", tok.Literal), tok.Line, tok.Col)
	}

	tok = tokenizer.NextToken()
	if tok.TokenType != T_EOL {
		return nil, createError("Expected end of line", tok.Line, tok.Col)
	}

	err := parseCols()

	return tableAst, err
}

func getTableByName(name string) (bool, *TabelAST) {
	for _, table := range ast.Tables {
		if name == table.Name {
			return true, table
		}
	}
	return false, nil
}

func parseCols() error {
	// <colmunName> <fieldType> [<@attribute>]
	tok := tokenizer.NextToken()

	if tok.TokenType != T_IDEN {
		return createError(fmt.Sprintf("Unexpected token '%s'", tok.Literal), tok.Line, tok.Col)
	}

	for tok.TokenType != T_END && tok.TokenType != T_EOF {

		if tok.TokenType == T_TABLE {
			return createError("Missing 'end' keyword", tok.Line, tok.Col)
		}

		colAst := &ColmunAST{tok.Literal, "", &AttributesAST{}}

		if tok.TokenType != T_IDEN {
			return createError(fmt.Sprintf("Unexpected token '%s'", tok.Literal), tok.Line, tok.Col)
		}

		colAst.Name = tok.Literal

		if checkIfColExists(tok.Literal, currentTableAst) {
			return createError(
				fmt.Sprintf("Colmun with name '%s' already declared", tok.Literal),
				tok.Line,
				tok.Col,
			)
		}

		err := parseColType(colAst)
		if err != nil {
			return nil
		}

		err = ParseColAttributes(colAst)
		if err != nil {
			return err
		}

		currentTableAst.Colmuns = append(currentTableAst.Colmuns, colAst)
		tok = tokenizer.NextToken()
	}

	return nil
}

func checkIfColExists(colName string, table *TabelAST) bool {
	for _, col := range table.Colmuns {
		if colName == col.Name {
			return true
		}
	}
	return false
}

func parseColType(colAst *ColmunAST) error {
	tok := tokenizer.PeekToken()
	if tok.TokenType != T_IDEN {
		if tok.TokenType == T_RAW {
			colAst.Data_type = "raw"
			(*colAst.Attributes)["raw"] = &AttributeAST{
				"raw",
				[]*AttributeArgAST{{tok.Literal, "string"}},
			}
			tokenizer.NextToken()
		} else {
			return createError("Missing data type after colmun name", tok.Line, tok.Col)
		}
	} else {
		tokenizer.NextToken()
		colAst.Data_type = tok.Literal
	}
	return nil
}

func ParseColAttributes(colAst *ColmunAST) error {
	tok := tokenizer.NextToken()

	for tok.TokenType != T_EOL {

		if tok.TokenType != T_ATTR {
			return createError(fmt.Sprintf("Unexpected token '%s'", tok.Literal), tok.Line, tok.Col)
		}

		mtok, args, err := parseColAttr(tok)

		if err != nil {
			return err
		}

		err = parseAttr(tok, args, colAst)
		if err != nil {
			return err
		}

		tok = mtok
	}
	return nil
}

func parseColAttr(ptok *Token) (*Token, []*AttributeArgAST, error) {
	tok := tokenizer.NextToken()
	values := []*AttributeArgAST{}
	if tok.TokenType == T_LEFT_PAREN {
		tok = tokenizer.NextToken()
		if tok.TokenType == T_STRING || tok.TokenType == T_RAW {

			attrValues, err := getAttrArgs(tok)
			if err != nil {
				return nil, nil, err
			}

			values = attrValues
			tok = tokenizer.NextToken()
		} else if tok.TokenType == T_RIGHT_PAREN {
			tok = tokenizer.NextToken()
			return tok, values, nil
		} else {
			return nil, nil, createError(fmt.Sprintf("Unexpected token '%s'", tok.Literal), tok.Line, tok.Col)
		}
	}

	return tok, values, nil
}

func getAttrArgs(tok *Token) ([]*AttributeArgAST, error) {
	attrArgs := []*AttributeArgAST{}

	for {
		attrArg := &AttributeArgAST{}
		if tok.TokenType == T_STRING {
			if tokenizer.PeekToken().TokenType == T_EOF {
				return nil, createError("Unterminated string", tok.Line, tok.Col)
			}
			attrArg.Type = "string"
		} else {
			attrArg.Type = "raw"
		}

		attrArg.Value = tok.Literal
		attrArgs = append(attrArgs, attrArg)

		tok = tokenizer.NextToken()

		if tok.TokenType == T_RIGHT_PAREN {
			break
		} else if tok.TokenType != T_COMMA {
			return nil, createError(fmt.Sprintf("Unexpected token '%s'", tok.Literal), tok.Line, tok.Col)
		}

		tok = tokenizer.NextToken()
	}
	return attrArgs, nil
}

func parseAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	f, exists := parseAttrFuncMap[tok.Literal]
	if !exists {
		return createError(fmt.Sprintf("Unknown attribute @%s", tok.Literal), tok.Line, tok.Col)
	}

	return f(tok, args, colAst)
}

func parseIdAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	if len(args) != 0 {
		return createError("@id takes no parameters", tok.Line, tok.Col)
	}
	(*colAst.Attributes)[tok.Literal] = &AttributeAST{tok.Literal, args}
	return nil
}

func parseDefaultAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	if len(args) != 1 {
		return createError("@default takes one parameters", tok.Line, tok.Col)
	}
	(*colAst.Attributes)[tok.Literal] = &AttributeAST{tok.Literal, args}
	return nil
}

func parseAutoIncrementAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	if len(args) != 0 {
		return createError("@auto_increment takes no parameters", tok.Line, tok.Col)
	}
	(*colAst.Attributes)[tok.Literal] = &AttributeAST{tok.Literal, args}
	return nil
}

func parseUniqueAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	if len(args) != 0 {
		return createError("@unique takes no parameters", tok.Line, tok.Col)
	}
	(*colAst.Attributes)[tok.Literal] = &AttributeAST{tok.Literal, args}
	return nil
}

func parseNullableAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	if len(args) != 0 {
		return createError("@nullable takes no parameters", tok.Line, tok.Col)
	}
	(*colAst.Attributes)[tok.Literal] = &AttributeAST{tok.Literal, args}
	return nil
}

func parseReferenceAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	if len(args) != 2 {
		return createError("@reference takes two parameters", tok.Line, tok.Col)
	}

	if args[0].Type != "string" || args[1].Type != "string" {
		return createError("@reference Expected string values", tok.Line, tok.Col)
	}

	exists, table := getTableByName(args[0].Value)
	if !exists {
		return createError(
			fmt.Sprintf("no such table '%s'", args[0].Value),
			tok.Line,
			tok.Col,
		)
	}

	if !checkIfColExists(args[1].Value, table) {
		return createError(
			fmt.Sprintf("no such col '%s' on table '%s'", args[1].Value, table.Name),
			tok.Line,
			tok.Col,
		)
	}

	currentTableAst.References = append(
		currentTableAst.References,
		&ReferenceAST{args[0].Value, args[1].Value, colAst.Name, "", ""},
	)
	return nil
}

func parseOnDeleteAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	if len(args) != 1 {
		return createError("@onDelete takes one parameters", tok.Line, tok.Col)
	}

	if args[0].Type != "string" {
		return createError("@onDelete Expected string value", tok.Line, tok.Col)
	}

	ref, exists := getRefByColName(colAst.Name)

	if !exists {
		return createError(
			"To use the @onDelete directive, you must first declare a reference using @reference",
			tok.Line,
			tok.Col,
		)
	}

	ref.OnDelete = args[0].Value

	return nil
}

func parseOnUpdateAttr(tok *Token, args []*AttributeArgAST, colAst *ColmunAST) error {
	if len(args) != 1 {
		return createError("@onUpdate takes one parameters", tok.Line, tok.Col)
	}

	if args[0].Type != "string" {
		return createError("@onUpdate Expected string value", tok.Line, tok.Col)
	}

	ref, exists := getRefByColName(colAst.Name)

	if !exists {
		return createError(
			"To use the @onUpdate directive, you must first declare a reference using @reference",
			tok.Line,
			tok.Col,
		)
	}

	ref.OnUpdate = args[0].Value

	return nil
}

func getRefByColName(colName string) (*ReferenceAST, bool) {
	for _, ref := range currentTableAst.References {
		if colName == ref.SourceCol {
			return ref, true
		}
	}
	return nil, false
}

func createError(msg string, line int, col int) error {
	return fmt.Errorf("Syntax Error:%d:%d: %s", line, col, msg)
}
