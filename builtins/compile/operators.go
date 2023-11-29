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

func assignForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	values := compileToCodeSlice(env, itArgs)
	switch casted := arg0.(type) {
	case types.Identifier:
		return wrapper{Renderer: jen.Id(string(casted)).Op(names.Assign).Add(values[0])}
	case *types.List:
		if id := extractAssignTargetFromList(env, casted); id != nil {
			return wrapper{Renderer: id.Op(names.Assign).Add(values[0])}
		}

		var ids []jen.Code
		types.ForEach(casted, func(elem types.Object) bool {
			ids = append(ids, extractAssignTarget(env, elem))
			return true
		})
		return wrapper{Renderer: jen.List(ids...).Op(names.Assign).List(values...)}
	}
	return wrappedErrorComment
}

// TODO manage multiline (using the keyword instead of ":=" ?)
func varForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, _ := itArgs.Next()
	values := compileToCodeSlice(env, itArgs)
	if len(values) == 0 {
		if list, ok := arg0.(*types.List); ok && list.Size() > 2 {
			varId, _ := list.LoadInt(1).(types.Identifier)
			return wrapper{Renderer: jen.Var().Id(string(varId)).Add(extractType(list.LoadInt(2)))}
		}
		return wrappedErrorComment
	}

	switch casted := arg0.(type) {
	case types.Identifier:
		return wrapper{Renderer: jen.Var().Id(string(casted)).Op(names.Assign).Add(values[0])}
	case *types.List:
		// test if it's a:b instead of (a:b c:d)
		if firstId, _ := casted.LoadInt(0).(types.Identifier); firstId == names.ListId {
			if casted.Size() < 2 {
				return wrappedErrorComment
			}

			varId, _ := casted.LoadInt(1).(types.Identifier)
			varCode := jen.Var().Id(string(varId))
			if typeStmt := extractType(casted.LoadInt(2)); typeStmt != nil {
				varCode.Add(typeStmt)
			}
			return wrapper{Renderer: varCode.Op(names.Assign).List(values...)}
		}

		var varIds []jen.Code
		var typeIds []*jen.Statement
		types.ForEach(casted, func(varDesc types.Object) bool {
			switch casted2 := varDesc.(type) {
			case types.Identifier:
				varIds = append(varIds, jen.Id(string(casted2)))
				typeIds = append(typeIds, nil)
			case *types.List:
				// assume it's in a:b format
				varId, _ := casted2.LoadInt(1).(types.Identifier)
				varIds = append(varIds, jen.Id(string(varId)))
				typeIds = append(typeIds, extractType(casted2.LoadInt(2)))
			}
			return true
		})

		// add cast calls
		for index, typeId := range typeIds {
			if typeId != nil {
				values[index] = typeId.Call(values[index])
			}
		}
		return wrapper{Renderer: jen.List(varIds...).Op(names.Var).List(values...)}
	}
	return wrappedErrorComment
}
