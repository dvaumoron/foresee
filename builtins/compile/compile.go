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

package compile

import (
	"github.com/dave/jennifer/jen"
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

const (
	// user can not directly use this kind of id (# start a comment)
	hiddenImportsName = "#imports"
	hiddenPackageName = "#package"

	mainId types.Identifier = "main"
)

var (
	wrappedErrorComment = wrapper{Renderer: jen.Comment("line with error, can't generate correct go code")}

	Builtins = initBuitins()
)

func Compile(l *types.List) types.Object {
	return l.Eval(compileEnvironment{Environment: types.MakeLocalEnvironment(Builtins)})
}

func initBuitins() types.BaseEnvironment {
	base := types.MakeBaseEnvironment()
	base.StoreStr(names.AddAssign, types.MakeNativeAppliable(addAssignForm))
	base.StoreStr(string(names.AmpersandId), types.MakeNativeAppliable(addressOrBitwiseAndForm))
	base.StoreStr(names.And, types.MakeNativeAppliable(andForm))
	base.StoreStr(names.AndAssign, types.MakeNativeAppliable(bitwiseAndAssignForm))
	base.StoreStr(names.AndNot, types.MakeNativeAppliable(bitwiseAndNotForm))
	base.StoreStr(names.AndNotAssign, types.MakeNativeAppliable(bitwiseAndNotAssignForm))
	base.StoreStr(names.Arrow, types.MakeNativeAppliable(receivingOrSendingForm))
	base.StoreStr(names.Assign, types.MakeNativeAppliable(assignForm))
	base.StoreStr(names.Block, types.MakeNativeAppliable(blockForm))
	base.StoreStr(names.Caret, types.MakeNativeAppliable(bitwiseXOrForm))
	base.StoreStr(names.Const, types.MakeNativeAppliable(constForm))
	base.StoreStr(names.DeclareAssign, types.MakeNativeAppliable(declareAssignForm))
	base.StoreStr(names.Decrement, types.MakeNativeAppliable(decrementForm))
	base.StoreStr(names.DivAssign, types.MakeNativeAppliable(divideAssignForm))
	base.StoreStr(names.Dot, types.MakeNativeAppliable(callMethodForm))
	base.StoreStr(string(names.EllipsisId), types.MakeNativeAppliable(extendSliceForm))
	base.StoreStr(names.Equal, types.MakeNativeAppliable(equalForm))
	base.StoreStr(string(names.FileId), types.MakeNativeAppliable(fileForm))
	base.StoreStr(string(names.FuncId), types.MakeNativeAppliable(funcForm)) // TODO lambda keyword to manage closure case
	base.StoreStr(names.Greater, types.MakeNativeAppliable(greaterForm))
	base.StoreStr(names.GreaterEqual, types.MakeNativeAppliable(greaterEqualForm))
	base.StoreStr(names.Import, types.MakeNativeAppliable(importForm))
	base.StoreStr(names.Increment, types.MakeNativeAppliable(incrementForm))
	base.StoreStr(names.Lesser, types.MakeNativeAppliable(lesserForm))
	base.StoreStr(names.LesserEqual, types.MakeNativeAppliable(lesserEqualForm))
	base.StoreStr(names.LShift, types.MakeNativeAppliable(leftShiftForm))
	base.StoreStr(names.LShiftAssign, types.MakeNativeAppliable(leftShiftAssignForm))
	base.StoreStr(string(names.LoadId), types.MakeNativeAppliable(indexOrSliceForm))
	base.StoreStr(names.Minus, types.MakeNativeAppliable(substractionForm))
	base.StoreStr(names.ModAssign, types.MakeNativeAppliable(moduloAssignForm))
	base.StoreStr(names.MultAssign, types.MakeNativeAppliable(multiplyAssignForm))
	base.StoreStr(names.Not, types.MakeNativeAppliable(notForm))
	base.StoreStr(names.NotEqual, types.MakeNativeAppliable(notEqualForm))
	base.StoreStr(names.Or, types.MakeNativeAppliable(orForm))
	base.StoreStr(names.OrAssign, types.MakeNativeAppliable(bitwiseOrAssignForm))
	base.StoreStr(names.Package, types.MakeNativeAppliable(packageForm))
	base.StoreStr(names.Percent, types.MakeNativeAppliable(moduloForm))
	base.StoreStr(names.Pipe, types.MakeNativeAppliable(bitwiseOrForm))
	base.StoreStr(names.Plus, types.MakeNativeAppliable(additionForm))
	base.StoreStr(names.Return, types.MakeNativeAppliable(returnForm))
	base.StoreStr(names.RShift, types.MakeNativeAppliable(rightShiftForm))
	base.StoreStr(names.RShiftAssign, types.MakeNativeAppliable(rightShiftAssignForm))
	base.StoreStr(names.Slash, types.MakeNativeAppliable(divideForm))
	base.StoreStr(string(names.StarId), types.MakeNativeAppliable(dereferenceOrMultiplyForm))
	base.StoreStr(names.SubAssign, types.MakeNativeAppliable(substractAssignForm))
	base.StoreStr(names.Var, types.MakeNativeAppliable(varForm))
	base.StoreStr(names.XorAssign, types.MakeNativeAppliable(bitwiseXOrAssignForm))

	// TODO
	return base
}
