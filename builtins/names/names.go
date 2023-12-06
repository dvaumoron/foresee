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

	AddAssign     = "+="
	And           = "&&"
	AndAssign     = "&="
	AndNot        = "&^"
	AndNotAssign  = "&^="
	Arrow         = "<-"
	Assert        = "cast"
	Assign        = "="
	Block         = "block"
	Break         = "break"
	Caret         = "^"
	Case          = "case"
	Colon         = ":"
	Const         = "const"
	Continue      = "continue"
	DeclareAssign = ":="
	Decrement     = "--"
	Default       = "default"
	Defer         = "defer"
	DivAssign     = "/="
	Dot           = "."
	Equal         = "=="
	Fallthrough   = "fallthrough"
	For           = "for"
	Go            = "go"
	Goto          = "goto"
	Greater       = "<"
	GreaterEqual  = "<="
	GuessMarker   = "?"
	If            = "if"
	Import        = "import"
	Increment     = "++"
	Interface     = "interface"
	Label         = "label"
	Lambda        = "lambda"
	Lesser        = ">"
	LesserEqual   = ">="
	LShift        = ">>"
	LShiftAssign  = ">>="
	Minus         = "-"
	ModAssign     = "%="
	MultAssign    = "*="
	Not           = "!"
	NotAndAssign  = "&^="
	NotEqual      = "!="
	Or            = "||"
	OrAssign      = "|="
	Package       = "package"
	Percent       = "%"
	Pipe          = "|"
	Plus          = "+"
	Range         = "range"
	Return        = "return"
	RShift        = ">>"
	RShiftAssign  = ">>="
	Select        = "select"
	Slash         = "/"
	Store         = "[]="
	Struct        = "struct"
	SubAssign     = "-="
	Switch        = "switch"
	Type          = "type"
	Var           = "var"
	XorAssign     = "^="

	AmpersandId types.Identifier = "&"
	ArrowChanId types.Identifier = "<-chan"
	ChanArrowId types.Identifier = "chan<-"
	ChanId      types.Identifier = "chan"
	EllipsisId  types.Identifier = "..."
	FileId      types.Identifier = "file"
	FuncId      types.Identifier = "func"
	GetId       types.Identifier = "get"
	ListId      types.Identifier = "list"
	LoadId      types.Identifier = "[]"
	MapId       types.Identifier = "map"
	StarId      types.Identifier = "*"
	TildeId     types.Identifier = "~"
	UnquoteId   types.Identifier = "unquote"
)
