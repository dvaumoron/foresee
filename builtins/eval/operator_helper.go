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

type evalIterator struct {
	types.NoneType
	inner types.Iterator
	env   types.Environment
}

func (e evalIterator) Iter() types.Iterator {
	return e
}

func (e evalIterator) Next() (types.Object, bool) {
	value, ok := e.inner.Next()
	return value.Eval(e.env), ok
}

func (e evalIterator) Close() {
	e.inner.Close()
}

func makeEvalIterator(it types.Iterator, env types.Environment) evalIterator {
	return evalIterator{inner: it, env: env}
}

func inplaceOperatorForm(env types.Environment, itArgs types.Iterator, opStr string) types.Object {
	arg, _ := itArgs.Next()
	opCall := types.NewList(types.Identifier(opStr), arg).AddAll(itArgs)
	return types.NewList(types.Identifier(names.Assign), arg, opCall).Eval(env)
}

func processUnaryOrBinaryMoreFunc(env types.Environment, itArgs types.Iterator, unaryFunc types.NativeFunc, binaryMoreFunc types.NativeFunc) types.Object {
	args := types.NewList().AddAll(itArgs)

	itArgs2 := args.Iter()
	defer itArgs2.Close()

	if args.Size() == 1 {
		return unaryFunc(env, itArgs2)
	}

	return binaryMoreFunc(env, itArgs2)
}
