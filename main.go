package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/pterm/pterm"
)

type record struct {
	variable string
	count    int
}

type recordSlice []record

func (r recordSlice) Len() int {
	return len(r)
}

func (r recordSlice) Less(i, j int) bool {
	return r[i].count > r[j].count
}

func (r recordSlice) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func printRanking(records []record) error {
	var tableData pterm.TableData
	tableData = append(tableData, []string{"variable", "count"})
	for _, v := range records {
		tableData = append(tableData, []string{v.variable, strconv.Itoa(v.count)})
	}
	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	return nil
}

func findGoFiles(path string) ([]string, error) {
	goFles, err := doublestar.Glob(path + "/**/*.go")
	if err != nil {
		return nil, err
	}
	return goFles, nil
}

func isSubjectOfDeclaration(id *ast.Ident) bool {
	if id.Obj != nil && id.Obj.Decl != nil {
		switch id.Obj.Kind {
		case ast.Var, ast.Typ, ast.Fun:
			return true
		}
	}
	return false
}

func findGlobalVariables(variables []string, file string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return variables, err
	}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.VAR {
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					for _, name := range vs.Names {
						variables = append(variables, name.Name)
					}
				}
			}
		}
	}
	return variables, nil
}

func getCount(file, variable string) (int, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return 0, err
	}

	counter := 0

	ast.Inspect(f, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		// Skip if this identifier is the subject of a declaration
		if isSubjectOfDeclaration(id) {
			return true
		}

		if id.Name == variable {
			counter++
		}

		return true
	})

	return counter, nil
}

func main() {
	var (
		dir = flag.String("dir", "", "directory")
		err error
	)
	flag.Parse()

	if *dir == "" {
		panic("Input -dir")
	}
	files, err := findGoFiles(*dir)
	if err != nil {
		panic(err)
	}

	variables := []string{}
	for _, file := range files {
		variables, err = findGlobalVariables(variables, file)
		if err != nil {
			panic(err)
		}
	}
	var records recordSlice
	for _, v := range variables {
		var count int
		for _, file := range files {
			c, err := getCount(file, v)
			if err != nil {
				panic(err)
			}
			count += c
		}
		records = append(records, record{
			variable: v,
			count:    count,
		})
	}
	sort.Sort(records)
	printRanking(records)
}
