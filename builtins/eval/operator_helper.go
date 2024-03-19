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

func inplaceOperatorForm(env types.Environment, itArgs types.Iterator, opStr string) types.Object {
	arg, _ := itArgs.Next()
	opCall := types.NewList(types.Identifier(opStr), arg).AddAll(itArgs)
	return types.NewList(types.Identifier(names.Assign), arg, opCall).Eval(env)
}

func inplaceUnaryOperatorForm(env types.Environment, itArgs types.Iterator, opStr string) types.Object {
	arg, _ := itArgs.Next()
	opCall := types.NewList(types.Identifier(opStr), arg).Add(types.Integer(1))
	return types.NewList(types.Identifier(names.Assign), arg, opCall).Eval(env)
}

func processUnaryOrBinaryMoreFunc(env types.Environment, itArgs types.Iterator, unaryFunc types.NativeFunc, binaryMoreFunc types.NativeFunc) types.Object {
	args := types.NewList().AddAll(itArgs)

	itArgs = args.Iter()
	defer itArgs.Close()

	if args.Size() == 1 {
		return unaryFunc(env, itArgs)
	}

	return binaryMoreFunc(env, itArgs)
}
