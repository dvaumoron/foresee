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

// handle "a" as a,  "(* a)" as a and ([] a b c) as a[b][c]
func buildAssignFunc(env types.Environment, object types.Object) func(types.Object) {
	switch casted := object.(type) {
	case types.Identifier:
		return func(value types.Object) {
			env.StoreStr(string(casted), value)
		}
	case *types.List:
		return buildAssignFuncFromList(env, casted)
	}
	return nil
}

func buildAssignFuncFromList(env types.Environment, list *types.List) func(types.Object) {
	size := list.Size()
	switch op, _ := list.LoadInt(0).(types.Identifier); op {
	case names.Load:
		if size < 3 {
			return nil
		}

		prevEnv, ok := env.Load(list.LoadInt(1)).(types.Storable)
		if !ok {
			return nil
		}

		i := 3
		index := list.LoadInt(2)
		currentEnv, ok := prevEnv.Load(index).(types.Storable)
		for i < size {
			if !ok {
				break
			}

			prevEnv = currentEnv
			index = list.LoadInt(i)
			currentEnv, ok = prevEnv.Load(index).(types.Storable)
			i++
		}

		if i != size { // breaked
			return nil
		}

		return func(value types.Object) {
			prevEnv.Store(index, value)
		}
	case names.StarId:
		if size > 1 {
			return buildAssignFunc(env, list.LoadInt(1))
		}
	}

	return nil
}
