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
	"errors"

	"github.com/dvaumoron/foresee/types"
)

var (
	initMapAppliable    = types.MakeNativeAppliable(initMapForm)
	initStructAppliable = types.MakeNativeAppliable(initStructForm)

	errIdentifierType = errors.New("wait identifier type")
	errListType       = errors.New("wait list type")
	errPairSize       = errors.New("wait at least 2 elements")
)

func initBaseFromPairs(env types.Environment, itArgs types.Iterator, pairAdder func(types.BaseEnvironment, *types.List, types.Environment)) types.Object {
	res := types.MakeBaseEnvironment()
	types.ForEach(itArgs, func(elem types.Object) bool {
		pair, ok := elem.(*types.List)
		if !ok {
			panic(errListType)
		}
		if pair.Size() < 3 {
			panic(errPairSize)
		}

		pairAdder(res, pair, env)

		return true
	})

	return res
}

func initMapForm(env types.Environment, itArgs types.Iterator) types.Object {
	return initBaseFromPairs(env, itArgs, mapPairAdder)
}

func initStructForm(env types.Environment, itArgs types.Iterator) types.Object {
	return initBaseFromPairs(env, itArgs, structPairAdder)
}

func mapPairAdder(res types.BaseEnvironment, pair *types.List, env types.Environment) {
	res.Store(pair.LoadInt(1).Eval(env), pair.LoadInt(2).Eval(env))
}

func structPairAdder(res types.BaseEnvironment, pair *types.List, env types.Environment) {
	id, ok := pair.LoadInt(1).(types.Identifier)
	if !ok {
		panic(errIdentifierType)
	}

	res.StoreStr(string(id), pair.LoadInt(2).Eval(env))
}
