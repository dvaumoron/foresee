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

package names

import "github.com/dvaumoron/foresee/types"

const (
	HiddenModule = "#module"

	Assign      = "="
	Block       = "block"
	Const       = "const"
	Func        = "func"
	GuessMarker = "?"
	Import      = "import"
	Package     = "package"
	Var         = ":="

	AmpersandId types.Identifier = "&"
	FileId      types.Identifier = "file"
	ListId      types.Identifier = "list"
	LoadId      types.Identifier = "[]"
	StarId      types.Identifier = "*"
	StoreId     types.Identifier = "[]="
	UnquoteId   types.Identifier = "unquote"
)

/*
TODO:

for if else break continue select switch case type struct chan map default fallthrough range go defer return goto

+ - * / % | ^ << >> &^ && || <- == != < > ! <= >= += -= *= /= %= &= &^= ++ --

Go keywords excluded:

interface var
*/
