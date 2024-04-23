package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func main() {
	file := flag.String("file", "", "A single file")
	dir := flag.String("dir", "", "Directory to multiple files")
	flag.Parse()

	if *file != "" && *dir != "" {
		log.Fatalln("Cannot specify both `file` and `dir`")
	}

	var filesList []string

	if *file != "" {
		filesList = append(filesList, *file)
	} else if *dir != "" {
		err := filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				filesList = append(filesList, path)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking directory: %v", err)
		}
	} else {
		log.Fatalln("Please specify either `file` or `dir`")
	}

	for _, filePath := range filesList {
		f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
		if err != nil {
			log.Fatalln("Unable to open file ", err)
		}

		instrumentedSourceCode, err := AddFmtToHandler(f)
		if err != nil {
			log.Fatalln(err)
		}
		f.WriteAt([]byte(instrumentedSourceCode), 0)
		f.Close()

	}
}

func AddFmtToHandler(file *os.File) (string, error) {
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", string(fileContent), 0)

	var imports []string

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			selectorExp, ok := node.Fun.(*ast.SelectorExpr)
			if ok {
				if selectorExp.Sel.String() == "HandleFunc" {
					basicLitArg, ok := node.Args[0].(*ast.BasicLit)
					if ok {
						handlerPath := StripQuote(basicLitArg.Value)

						newPrintExp := &ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "fmt",
									},
									Sel: &ast.Ident{
										Name: "Println",
									},
								},
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: fmt.Sprintf("\"Invoking HandlerFunc: '%s'\"", handlerPath),
									},
								},
							},
						}
						handler, ok := node.Args[1].(*ast.FuncLit)
						if ok {
							handler.Body.List = append([]ast.Stmt{newPrintExp}, handler.Body.List...)
						}
					}

				}
			}
		case *ast.ImportSpec:
			imports = append(imports, strings.ToLower(StripQuote(node.Path.Value)))
		}
		return true
	})

	if !slices.Contains(imports, "fmt") {
		fmtImport := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote("fmt"),
			},
		}

		var impDeclCount int
		var lastImportDecl *ast.GenDecl
		for _, tmpDecl := range f.Decls {
			if tmpDecl, ok := tmpDecl.(*ast.GenDecl); ok && tmpDecl.Tok == token.IMPORT {
				impDeclCount++
				lastImportDecl = tmpDecl
			}
		}

		if impDeclCount == 0 {
			decl := &ast.GenDecl{
				Tok: token.IMPORT,
			}
			decl.Specs = append(decl.Specs, fmtImport)
			f.Decls = append([]ast.Decl{decl}, f.Decls...)
		} else {
			lastImportDecl.Specs = append(lastImportDecl.Specs, fmtImport)
		}
	}

	var buffer []byte
	buf := bytes.NewBuffer(buffer)

	err = format.Node(buf, fset, f)
	if err != nil {
		fmt.Println("format: ", err)
		return "", err
	}

	out, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println("source :", err)
		return "", err
	}

	buf.Reset()
	buf.Write(out)

	return buf.String(), nil
}

func StripQuote(in string) string {
	return strings.Replace(in, "\"", "", -1)
}
