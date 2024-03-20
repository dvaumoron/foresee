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
	"github.com/dvaumoron/foresee/types"
)

var initOrConvertMapAppliable = types.MakeNativeAppliable(initOrConverMapForm)

func copyStruct(env types.Environment, src types.Iterable, typeName string) types.Object {
	return initFromPairs[types.Environment](env, src, makeDynamicObject(env, typeName), 2, copyPairAdder)
}

func extractTypeName(o types.Object) string {
	switch casted := o.(type) {
	case types.Identifier:
		return string(casted)
	case *types.List:
		if op, _ := casted.LoadInt(0).(types.Identifier); op == names.AmpersandId || op == names.GenId || op == names.LitId {
			return extractTypeName(casted.LoadInt(1)) // no need to test for too short list : LoadInt call return None, recursive call panic the same way.
		}
	}
	panic(errIdentifierType)
}

func initFromPairs[T types.Object](env types.Environment, args types.Iterable, o T, tooSmallSize int, pairAdder func(T, *types.List, types.Environment)) types.Object {
	types.ForEach(args, func(elem types.Object) bool {
		pair, ok := elem.(*types.List)
		if !ok {
			panic(errListType)
		}
		if pair.Size() < tooSmallSize {
			panic(errPairSize)
		}

		pairAdder(o, pair, env)

		return true
	})

	return o
}

func initOrConverMapForm(env types.Environment, itArgs types.Iterator) types.Object {
	args := types.NewList().AddAll(itArgs)
	if args.Size() != 1 {
		return initFromPairs[types.Storable](env, itArgs, makeDynamicMap(), 3, mapPairAdder)
	}

	arg1 := args.LoadInt(0)
	if casted, ok := arg1.(*types.List); ok {
		if id, _ := casted.LoadInt(0).(types.Identifier); id == names.ListId {
			return initFromPairs[types.Storable](env, itArgs, makeDynamicMap(), 3, mapPairAdder)
		}
	}

	casted, ok := arg1.Eval(env).(dynamicMap)
	if !ok {
		panic(errConversion)
	}

	// type conversion case (nothing to do, eval mode does not track map subtype)
	return casted
}

func initStruct(env types.Environment, args types.Iterable, typeName string) types.Object {
	return initFromPairs[types.Environment](env, args, makeDynamicObject(env, typeName), 3, structPairAdder)
}

func copyPairAdder(res types.Environment, pair *types.List, _ types.Environment) {
	id, ok := pair.LoadInt(0).(types.String)
	if !ok {
		panic(errStringType)
	}

	res.StoreStr(string(id), pair.LoadInt(1))
}

func mapPairAdder(res types.Storable, pair *types.List, env types.Environment) {
	res.Store(pair.LoadInt(1).Eval(env), pair.LoadInt(2).Eval(env))
}

func structPairAdder(res types.Environment, pair *types.List, env types.Environment) {
	id, ok := pair.LoadInt(1).(types.Identifier)
	if !ok {
		panic(errIdentifierType)
	}

	res.StoreStr(string(id), pair.LoadInt(2).Eval(env))
}
