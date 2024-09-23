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
	"iter"
	"strings"

	"github.com/dvaumoron/foresee/builtins/names"
	"github.com/dvaumoron/foresee/types"
)

func concatStrings(args types.Iterable) types.Object {
	var builder strings.Builder
	for arg := range args.Iter() {
		temp, isString := arg.(types.String)
		if !isString {
			panic(errStringType)
		}
		builder.WriteString(string(temp))
	}

	return types.String(builder.String())
}

func inplaceOperatorForm(env types.Environment, itArgs iter.Seq[types.Object], opStr string) types.Object {
	next, stop := types.Pull(itArgs)
	defer stop()

	arg, _ := next()
	opCall := types.NewList(types.Identifier(opStr), arg).AddAll(types.Push(next))
	return types.NewList(types.Identifier(names.Assign), arg, opCall).Eval(env)
}

func inplaceUnaryOperatorForm(env types.Environment, itArgs iter.Seq[types.Object], opStr string) types.Object {
	var arg0 types.Object = types.None
	for arg := range itArgs {
		arg0 = arg
		break
	}
	opCall := types.NewList(types.Identifier(opStr), arg0).Add(types.Integer(1))
	return types.NewList(types.Identifier(names.Assign), arg0, opCall).Eval(env)
}

func processUnaryOrBinaryMoreFunc(env types.Environment, itArgs iter.Seq[types.Object], unaryFunc types.NativeFunc, binaryMoreFunc types.NativeFunc) types.Object {
	args := types.NewList().AddAll(itArgs)

	itArgs = args.Iter()
	if args.Size() == 1 {
		return unaryFunc(env, itArgs)
	}
	return binaryMoreFunc(env, itArgs)
}
