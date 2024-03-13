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

package eval

import (
	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/parser"
	"github.com/dvaumoron/foresee/types"
)

var Builtins = initBuitins()

func initBuitins() types.BaseEnvironment {
	noOpAppliable := types.MakeNativeAppliable(noOp)

	base := types.MakeBaseEnvironment()
	base.StoreStr(names.AddAssign, types.MakeNativeAppliable(sumSetForm))
	base.StoreStr(string(names.AmpersandId), types.MakeNativeAppliable(addressOrBitwiseAndForm))
	base.StoreStr(names.And, types.MakeNativeAppliable(andForm))
	base.StoreStr(names.AndAssign, types.MakeNativeAppliable(bitwiseAndAssignForm))
	base.StoreStr(names.AndNot, types.MakeNativeAppliable(bitwiseAndNotFunc))
	base.StoreStr(names.AndNotAssign, types.MakeNativeAppliable(bitwiseAndNotAssignForm))
	base.StoreStr(names.Arrow, types.MakeNativeAppliable(receivingOrSendingForm))
	base.StoreStr(names.Assert, types.MakeNativeAppliable(assertForm))
	base.StoreStr(names.Assign, types.MakeNativeAppliable(assignForm))
	base.StoreStr(names.Block, types.MakeNativeAppliable(blockForm))
	base.StoreStr(names.Break, types.MakeNativeAppliable(breakForm))
	base.StoreStr(names.Caret, types.MakeNativeAppliable(bitwiseXOrFunc))
	base.StoreStr(names.Case, types.MakeNativeAppliable(caseForm))
	base.StoreStr(names.Const, types.MakeNativeAppliable(constForm))
	base.StoreStr(names.Continue, types.MakeNativeAppliable(continueForm))
	base.StoreStr(names.DeclareAssign, types.MakeNativeAppliable(assignForm))
	base.StoreStr(names.Decrement, types.MakeNativeAppliable(decrementForm))
	base.StoreStr(names.Default, types.MakeNativeAppliable(defaultForm))
	base.StoreStr(names.Defer, types.MakeNativeAppliable(deferForm))
	base.StoreStr(names.DivAssign, types.MakeNativeAppliable(divideSetForm))
	base.StoreStr(names.Dot, types.MakeNativeAppliable(callMethodForm))
	base.StoreStr(string(names.EllipsisId), types.MakeNativeAppliable(extendSliceForm))
	base.StoreStr(names.Equal, types.MakeNativeAppliable(equalForm))
	base.StoreStr(names.Fallthrough, types.MakeNativeAppliable(fallthroughForm))
	base.StoreStr(string(names.FileId), types.MakeNativeAppliable(fileForm))
	base.StoreStr(names.For, types.MakeNativeAppliable(forForm))
	base.StoreStr(string(names.FuncId), types.MakeNativeAppliable(funcForm))
	base.StoreStr(string(names.GenId), types.MakeNativeAppliable(genTypeForm))
	base.StoreStr(string(names.GetId), types.MakeNativeAppliable(getForm))
	base.StoreStr(names.Go, types.MakeNativeAppliable(goForm))
	base.StoreStr(names.Goto, types.MakeNativeAppliable(gotoForm))
	base.StoreStr(names.Greater, types.MakeNativeAppliable(greaterForm))
	base.StoreStr(names.GreaterEqual, types.MakeNativeAppliable(greaterEqualForm))
	base.StoreStr(names.If, types.MakeNativeAppliable(ifForm))
	base.StoreStr(names.Import, types.MakeNativeAppliable(importForm))
	base.StoreStr(names.Increment, types.MakeNativeAppliable(incrementForm))
	base.StoreStr(names.Label, types.MakeNativeAppliable(labelForm))
	base.StoreStr(names.Lambda, types.MakeNativeAppliable(lambdaForm))
	base.StoreStr(names.Lesser, types.MakeNativeAppliable(lesserForm))
	base.StoreStr(names.Lesser, types.MakeNativeAppliable(lesserForm))
	base.StoreStr(names.LesserEqual, types.MakeNativeAppliable(lesserEqualForm))
	base.StoreStr(string(names.LitId), types.MakeNativeAppliable(literalForm))
	base.StoreStr(names.Load, types.MakeNativeAppliable(indexOrSliceForm))
	base.StoreStr(names.LShift, types.MakeNativeAppliable(leftShiftForm))
	base.StoreStr(names.LShiftAssign, types.MakeNativeAppliable(leftShiftAssignForm))
	base.StoreStr(names.Macro, types.MakeNativeAppliable(macroForm))
	base.StoreStr(string(names.MapId), types.MakeNativeAppliable(mapTypeForm))
	base.StoreStr(names.Minus, types.MakeNativeAppliable(minusFunc))
	base.StoreStr(names.ModAssign, types.MakeNativeAppliable(remainderSetForm))
	base.StoreStr(names.MultAssign, types.MakeNativeAppliable(productSetForm))
	base.StoreStr(string(names.NotId), types.MakeNativeAppliable(notFunc))
	base.StoreStr(names.NotEqual, types.MakeNativeAppliable(notEqualForm))
	base.StoreStr(names.Or, types.MakeNativeAppliable(orForm))
	base.StoreStr(names.OrAssign, types.MakeNativeAppliable(bitwiseOrAssignForm))
	base.StoreStr(names.Package, noOpAppliable)
	base.StoreStr(names.Percent, types.MakeNativeAppliable(remainderFunc))
	base.StoreStr(names.Pipe, types.MakeNativeAppliable(bitwiseXOrFunc))
	base.StoreStr(names.Plus, types.MakeNativeAppliable(sumFunc))
	base.StoreStr(names.Range, types.MakeNativeAppliable(rangeForm))
	base.StoreStr(names.Return, types.MakeNativeAppliable(returnForm))
	base.StoreStr(names.RShift, types.MakeNativeAppliable(rightShiftForm))
	base.StoreStr(names.RShiftAssign, types.MakeNativeAppliable(rightShiftAssignForm))
	base.StoreStr(names.Select, types.MakeNativeAppliable(selectForm))
	base.StoreStr(names.Slash, types.MakeNativeAppliable(divideFunc))
	base.StoreStr(string(names.SliceId), types.MakeNativeAppliable(sliceOrArrayTypeForm))
	base.StoreStr(string(names.StarId), types.MakeNativeAppliable(dereferenceOrMultiplyForm))
	base.StoreStr(names.Store, types.MakeNativeAppliable(storeForm))
	base.StoreStr(names.SubAssign, types.MakeNativeAppliable(minusSetForm))
	base.StoreStr(names.Switch, types.MakeNativeAppliable(switchForm))
	base.StoreStr(names.Type, types.MakeNativeAppliable(typeForm))
	base.StoreStr(names.Var, types.MakeNativeAppliable(varForm))
	base.StoreStr(names.XorAssign, types.MakeNativeAppliable(bitwiseXOrAssignForm))

	// give parser package a protected copy to use in user custom rules
	parser.BuiltinsCopy = types.MakeLocalEnvironment(base)
	return base
}

func evalFirstOp(env types.Environment, itArgs types.Iterator) types.Object {
	args0, _ := itArgs.Next()

	return args0.Eval(env)
}

func noOp(env types.Environment, itArgs types.Iterator) types.Object {
	return types.None
}
