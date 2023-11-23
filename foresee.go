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
		fmt.Printf("Error while reading go.mod : %s", err)
		return false
	}
	defer goModFile.Close()

	moduleName, ok := "", false
	scanner := bufio.NewScanner(goModFile)
	for scanner.Scan() {
		if moduleName, ok = strings.CutPrefix("module ", scanner.Text()); !ok {
			fmt.Printf("Error while parsing go.mod : should start with module declaration line")
			return false
		}
	}
	if err = scanner.Err(); err != nil {
		fmt.Printf("Error while parsing go.mod : %s", err)
		return false
	}
	if ok = moduleName != ""; ok {
		eval.Builtins.StoreStr(names.HiddenModule, types.String(moduleName))
		compile.Builtins.StoreStr(names.HiddenModule, types.String(moduleName))
	}
	return ok
}

func processFile(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error while reading %s : %s", filePath, err)
		return
	}

	parsed, err := parser.Parse(string(data))
	if err != nil {
		fmt.Printf("Error while parsing %s : %s", filePath, err)
		return
	}

	expanded, err := eval.ExpandMacro(parsed)
	if err != nil {
		fmt.Printf("Error while expanding %s : %s", filePath, err)
		return
	}

	// TODO manage inference across multiple file
	infered, err := infer.InferTypes(expanded)
	if err != nil {
		fmt.Printf("Error while infering %s : %s", filePath, err)
		return
	}

	var outputdata bytes.Buffer
	outputPath := computeOutputPath(filePath)
	if err = compile.Compile(infered).Render(&outputdata); err != nil {
		fmt.Printf("Error while rendering %s : %s", outputPath, err)
		return
	}

	if err = os.WriteFile(outputPath, outputdata.Bytes(), 0644); err != nil {
		fmt.Printf("Error while writing %s : %s", outputPath, err)
	}
}

func computeOutputPath(filePath string) string {
	if dotIndex := strings.LastIndexByte(filePath, '.'); dotIndex != -1 {
		filePath = filePath[:dotIndex]
	}
	return filePath + ".go"
}
