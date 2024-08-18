/*
 *
 * Copyright 2023 foresee authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, v. 2.0.
 *
 */

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dvaumoron/foresee/builtins/compile"
	"github.com/dvaumoron/foresee/builtins/eval"
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/infer"
	"github.com/dvaumoron/foresee/parser"
	"github.com/dvaumoron/foresee/types"
)

const (
	fileExt    = ".fc"
	fileExtLen = len(fileExt)
)

//go:generate gennames -output "builtins/compile/hints.go" -package "compile" -name "standardLibraryHints" -standard -novendor -path "./..."

func main() {
	if !loadGoMod() {
		return
	}

	if len(os.Args) > 1 {
		for _, filePath := range os.Args[1:] {
			processFile(filePath)
		}
		return
	}

	fmt.Println("No files listed, walking current directory")
	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && path[len(path)-fileExtLen:] == fileExt {
			processFile(path)
		}
		return err
	})
}

func loadGoMod() bool {
	goModFile, err := os.Open("go.mod")
	if err != nil {
		fmt.Println("Error while reading go.mod :", err)
		return false
	}
	defer goModFile.Close()

	moduleName, ok := "", false
	scanner := bufio.NewScanner(goModFile)
	if scanner.Scan() {
		if moduleName, ok = strings.CutPrefix(scanner.Text(), "module "); !ok {
			fmt.Println("Error while parsing go.mod : should start with module declaration line")
			return false
		}
	}
	if err = scanner.Err(); err != nil {
		fmt.Println("Error while parsing go.mod :", err)
		return false
	}
	if ok = moduleName != ""; ok {
		eval.Builtins.StoreStr(names.HiddenModule, types.String(moduleName))
		compile.Builtins.StoreStr(names.HiddenModule, types.String(moduleName))
	}
	return ok
}

func parseFile(filePath string) (*types.List, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parser.Parse(file)
}

func processFile(filePath string) {
	parsed, err := parseFile(filePath)
	if err != nil {
		fmt.Println("Error while opening and parsing", filePath, ":", err)
		return
	}

	expanded, err := eval.ExpandMacro(parsed)
	if err != nil {
		fmt.Println("Error while expanding", filePath, ":", err)
		return
	}

	// TODO manage inference across multiple file
	infered, err := infer.InferTypes(expanded)
	if err != nil {
		fmt.Println("Error while infering", filePath, ":", err)
		return
	}

	var outputdata bytes.Buffer
	outputPath := computeOutputPath(filePath)
	if err = compile.Compile(infered).Render(&outputdata); err != nil {
		fmt.Println("Error while rendering", outputPath, ":", err)
		return
	}

	if err = os.WriteFile(outputPath, outputdata.Bytes(), 0644); err != nil {
		fmt.Println("Error while writing", outputPath, ":", err)
	}
}

func computeOutputPath(filePath string) string {
	if dotIndex := strings.LastIndexByte(filePath, '.'); dotIndex != -1 {
		filePath = filePath[:dotIndex]
	}
	return filePath + ".go"
}
