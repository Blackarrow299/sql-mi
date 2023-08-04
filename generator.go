package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Provider string
type Type map[Provider]string

const (
	sqlite Provider = "sqlite"
)

var provider Provider = sqlite
var types map[string]Type
var attrFuncMap map[string]func(attr AttributeAST) (string, error)

func initValues() {
	INT := Type{
		sqlite: "INTEGER",
	}

	STRING := Type{
		sqlite: "TEXT",
	}

	BOOL := Type{
		sqlite: "NUMERIC",
	}

	DATETIME := Type{
		sqlite: "NUMERIC",
	}

	FLOAT := Type{
		sqlite: "REAL",
	}

	BLOB := Type{
		sqlite: "BLOB",
	}

	types = map[string]Type{
		"int":      INT,
		"string":   STRING,
		"boolean":  BOOL,
		"bool":     BOOL,
		"datetime": DATETIME,
		"float":    FLOAT,
		"blob":     BLOB,
	}

	attrFuncMap = map[string]func(AttributeAST) (string, error){
		"id":             handleIdAttr,
		"default":        handleDefaultAttr,
		"auto_increment": handleAutoIncrementAttr,
		"nullable":       handleNullableAttr,
	}
}

func GenerateSQL(ast AST) (string, error) {

	initValues()

	if !isProviderAvailable(ast.Provider) {
		return "", errors.New("Error: Provider not supported")
	}

	if len(ast.Tables) == 0 {
		return "", errors.New("Error: No tables declared")
	}

	builder := strings.Builder{}
	for _, table := range ast.Tables {
		sqlStr, err := generateTableSQL(table)
		if err != nil {
			return "", err
		}
		builder.WriteString(sqlStr + "\n\n")
	}

	return builder.String(), nil
}

func generateTableSQL(tableAST TabelAST) (string, error) {
	if !isValidTableName(tableAST.Name) {
		return "", errors.New(fmt.Sprintf("Error: Bad name for table '%s'", tableAST.Name))
	}

	if len(tableAST.Colmuns) == 0 {
		return "", errors.New("Error: No Colmuns Specified for Table")
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableAST.Name))

	for _, colmun := range tableAST.Colmuns {
		colStr, err := handleColmun(colmun, tableAST.Name)
		if err != nil {
			return "", err
		}
		builder.WriteString("\t" + colStr + "\n")
	}

	for _, ref := range tableAST.References {
		builder.WriteString(handleRef(ref))
	}

	builder.WriteString(");")

	return builder.String(), nil
}

func handleRef(ref ReferenceAST) string {
	return fmt.Sprintf(
		"\tFOREIGN KEY %s REFERENCES %s(%s),\n",
		ref.SourceCol,
		ref.TargetTable,
		ref.TargetCol,
	)
}

func handleColmun(
	colmun ColmunAST,
	tableName string,
) (string, error) {

	if !isValidColmunName(colmun.Name) {
		return "", errors.New(
			fmt.Sprintf("Error: Bad colmun name '%s' for table '%s'", colmun.Name, tableName),
		)
	}

	colmunType, err := getType(colmun)
	if err != nil {
		return "", err
	}

	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s %s", colmun.Name, colmunType))

	hasNullableAttr := false
	for _, attr := range colmun.Attributes {
		if attr.Name == "nullable" {
			hasNullableAttr = true
		}

		str, err := handleAttr(attr)
		if err != nil {
			return "", err
		}
		builder.WriteString(" " + str)
	}

	if !hasNullableAttr {
		builder.WriteString(" NOT NULL")
	}

	builder.WriteString(",")
	return builder.String(), nil
}

func handleAttr(attr AttributeAST) (string, error) {
	if attr.Name == "raw" {
		return "", nil
	}

	f, exists := attrFuncMap[attr.Name]
	if !exists {
		return "", errors.New(
			fmt.Sprintf("Error: '%s' Does not exist in the current context.", attr.Name),
		)
	}
	sqlStr, err := f(attr)
	return sqlStr, err
}

func handleIdAttr(attr AttributeAST) (string, error) {
	if len(attr.Values) != 0 {
		return "", errors.New("Error: id takes no parameters")
	}
	return "PRIMARY KEY UNIQUE", nil
}

func handleDefaultAttr(attr AttributeAST) (string, error) {
	output := "DEFAULT "
	if len(attr.Values) != 1 {
		return "", errors.New("Error: default takes one parameter")
	}

	if attr.Values[0].Type == "SqlFunction" {
		output += fmt.Sprintf("%s", attr.Values[0].Value)
	} else {
		output += fmt.Sprintf("'%s'", attr.Values[0].Value)
	}

	return output, nil
}

func handleAutoIncrementAttr(attr AttributeAST) (string, error) {
	if len(attr.Values) > 0 {
		return "", errors.New("Error: auto_increment takes no parameters")
	}
	return "AUTO_INCREMENT", nil
}

func handleNullableAttr(attr AttributeAST) (string, error) {
	if len(attr.Values) > 0 {
		return "", errors.New("Error: nullable takes no parameters")
	}
	return "NULL", nil
}

func getType(colmun ColmunAST) (string, error) {
	var colmunDataTypeRes string

	if colmun.Data_type == "raw" {
		attr, exists := colmun.Attributes["raw"]
		if !exists {
			return "", errors.New("Error: Expected raw attribute")
		}

		if len(attr.Values) != 1 {
			return "", errors.New("Error: Raw attribute requires one parameter")
		}

		colmunDataTypeRes = attr.Values[0].Value
	} else {
		colmunDataType, exists := types[colmun.Data_type][provider]
		if !exists {
			return "", errors.New(fmt.Sprintf("Error: Invalid data type: %s", colmun.Data_type))
		}
		colmunDataTypeRes = colmunDataType
	}
	return colmunDataTypeRes, nil
}

func isProviderAvailable(provider string) bool {
	if provider != string(sqlite) {
		return false
	}
	return true
}

func isValidTableName(tableName string) bool {
	pattern := `^[a-zA-Z_][a-zA-Z0-9_$]*$`

	regex := regexp.MustCompile(pattern)

	return regex.MatchString(tableName)
}

func isValidColmunName(colName string) bool {
	return true
}
