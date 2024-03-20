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

var initMapAppliable = types.MakeNativeAppliable(initMapForm)

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

func initFromPairs[T types.Object](env types.Environment, itArgs types.Iterator, o T, pairAdder func(T, *types.List, types.Environment)) types.Object {
	types.ForEach(itArgs, func(elem types.Object) bool {
		pair, ok := elem.(*types.List)
		if !ok {
			panic(errListType)
		}
		if pair.Size() < 3 {
			panic(errPairSize)
		}

		pairAdder(o, pair, env)

		return true
	})

	return o
}

func initMapForm(env types.Environment, itArgs types.Iterator) types.Object {
	return initFromPairs[types.Storable](env, itArgs, makeDynamicMap(), mapPairAdder)
}

func initStructForm(env types.Environment, itArgs types.Iterator, typeName string) types.Object {
	return initFromPairs[types.Environment](env, itArgs, makeDynamicObject(env, typeName), structPairAdder)
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
