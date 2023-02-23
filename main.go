/*
 * Copyright 2023 Nathan P. Bombana
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 */

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var search = flag.String("search", "", "Search for a table with specific Title. \n Must include the '*' in the title")
var pretty = flag.Bool("pretty", false, "Pretty print the JSON output")

var rowsRegex = regexp.MustCompile(`(((^|\S)*( |$)) *|$)`)

type TableReadToken = uint8

const (
	TableRead TableReadToken = iota
	ColumnRead
	ValueRead
)

type Column struct {
	Title  string `json:"columnTitle"`
	Length int    `json:"length"`
}

type Table struct {
	Title   string   `json:"title"`
	Columns []Column `json:"columns"`
	Values  [][]string
}

type OutputTable struct {
	Title  string              `json:"title"`
	Values []map[string]string `json:"values"`
}

var tables []Table
var currentTable Table

func main() {
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)

	var tableFound = false

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '!' {
			continue
		}

		read := advance(line)
		if read == TableRead {
			lastTable := tables[len(tables)-1]
			if *search != "" && lastTable.Title == *search {
				tableFound = true
				printTable()
				return
			}
		}
	}

	if *search != "" && !tableFound {
		panic("Table not found")
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	printTable()
}

func advance(nextLine string) TableReadToken {
	switch nextLine[0] {
	case '*':
		{
			tables = append(tables, currentTable)
			currentTable = Table{
				Title:   strings.Trim(nextLine, " "),
				Columns: []Column{},
				Values:  [][]string{},
			}
			return TableRead
		}
	case '@':
		{
			rowTitles := rowsRegex.FindAllStringSubmatch(nextLine, -1)
			currentTable.Columns = make([]Column, len(rowTitles))
			for i, title := range rowTitles {
				currentTable.Columns[i] = Column{
					Title:  strings.Trim(title[0], " "),
					Length: len(title[0]),
				}
			}
			return ColumnRead
		}
	default:
		{
			var values = make([]string, 0)

			var offset = 0
			for _, row := range currentTable.Columns {

				nextValue := strings.Trim(nextLine[offset:offset+row.Length], " ")
				values = append(values, nextValue)
				offset += row.Length
			}

			currentTable.Values = append(currentTable.Values, values)
			return ValueRead
		}
	}
}

func tableToJson(table Table) OutputTable {
	outputTable := OutputTable{
		Title:  table.Title,
		Values: make([]map[string]string, len(table.Values)),
	}
	for i, row := range table.Values {
		outputRow := make(map[string]string)

		for j, cell := range row {
			outputRow[table.Columns[j].Title] = cell
		}
		outputTable.Values[i] = outputRow
	}

	return outputTable
}

func tablesToJson() []any {
	jsonArray := make([]any, len(tables))
	for _, table := range tables {
		jsonBytes := tableToJson(table)
		jsonArray = append(jsonArray, jsonBytes)
	}
	return jsonArray
}

func printTable() {
	var jsonBytes any

	if *search != "" {
		jsonBytes = tableToJson(tables[len(tables)-1])
	} else {
		jsonBytes = tablesToJson()
	}

	if *pretty == true {
		str, _ := json.MarshalIndent(jsonBytes, "", "  ")
		fmt.Println(string(str))
	} else {
		str, _ := json.Marshal(jsonBytes)
		fmt.Println(string(str))
	}

}
