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

	AndEqual    = "&="
	Arrow       = "<-"
	Assign      = "="
	Block       = "block"
	Const       = "const"
	GuessMarker = "?"
	Import      = "import"
	NotAndEqual = "&^="
	Package     = "package"
	Return      = "return"
	Var         = ":="

	AmpersandId types.Identifier = "&"
	ArrowChanId types.Identifier = "<-chan"
	ChanArrowId types.Identifier = "chan<-"
	ChanId      types.Identifier = "chan"
	EllipsisId  types.Identifier = "..."
	FileId      types.Identifier = "file"
	FuncId      types.Identifier = "func"
	ListId      types.Identifier = ":"
	LoadId      types.Identifier = "[]"
	MapId       types.Identifier = "map"
	StarId      types.Identifier = "*"
	StoreId     types.Identifier = "[]="
	UnquoteId   types.Identifier = "unquote"
)

/*
TODO:

for if else break continue select switch case type struct default fallthrough range go defer return goto

+ - * / % | ^ << >> &^ && || == != < > ! <= >= += -= *= /= %= &= &^= ++ --

Go keywords excluded:

interface var
*/
