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

func addressOrBitwiseAndForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		return wrappedErrorComment
	}

	valueCodesTemp := compileToCodeSlice(env, itArgs)
	if len(valueCodesTemp) == 0 {
		adressable := Renderer(extractType(arg0))
		if adressable == nil {
			adressable = compileToCode(env, arg0)
		}
		// adressing, usable to build a literal
		return literalWrapper{Renderer: jen.Op(string(names.AmpersandId)).Add(adressable)}
	}

	andCode := compileToCode(env, arg0)
	for _, code := range valueCodesTemp {
		andCode.Op(string(names.AmpersandId)).Add(code)
	}
	return wrapper{Renderer: andCode}
}

func assignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.Assign)
}

func declareAssignForm(env types.Environment, itArgs types.Iterator) types.Object {
	return processAssign(env, itArgs, names.DeclareAssign)
}

func processAssign(env types.Environment, itArgs types.Iterator, op string) types.Object {
	arg0, _ := itArgs.Next()
	values := compileToCodeSlice(env, itArgs)
	switch casted := arg0.(type) {
	case types.Identifier:
		return wrapper{Renderer: jen.Id(string(casted)).Op(op).Add(values[0])}
	case *types.List:
		if id := extractAssignTargetFromList(env, casted); id != nil {
			return wrapper{Renderer: id.Op(op).Add(values[0])}
		}

		var ids []jen.Code
		types.ForEach(casted, func(elem types.Object) bool {
			ids = append(ids, extractAssignTarget(env, elem))
			return true
		})
		return wrapper{Renderer: jen.List(ids...).Op(op).List(values...)}
	}
	return wrappedErrorComment
}
