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

var (
	_ types.Appliable = appliableWrapper{}
	_ types.Appliable = literalWrapper{}
)

type wrapper struct {
	types.Renderer
}

func (w wrapper) Eval(types.Environment) types.Object {
	return w
}

type appliableWrapper struct {
	// not composing wrapper to avoid a type changement on eval
	types.Renderer
}

func (w appliableWrapper) Eval(types.Environment) types.Object {
	return w
}

func (w appliableWrapper) Apply(env types.Environment, args types.Iterable) types.Object {
	if casted, ok := w.Renderer.(*jen.Statement); ok {
		argsCode := compileToCodeSlice(env, args)
		// still appliable ("f(a)(b)" is possible)
		return appliableWrapper{Renderer: casted.Clone().Call(argsCode...)}
	}
	return w
}

type literalWrapper struct {
	// not composing wrapper to avoid a type changement on eval
	types.Renderer
}

func (w literalWrapper) Eval(types.Environment) types.Object {
	return w
}

func (w literalWrapper) Apply(env types.Environment, args types.Iterable) types.Object {
	if casted, ok := w.Renderer.(*jen.Statement); ok {
		itArgs := args.Iter()
		defer itArgs.Close()

		var argsCode []jen.Code
		if arg0, ok := itArgs.Next(); ok {
			casted, _ := arg0.(*types.List)
			header, _ := casted.LoadInt(0).(types.Identifier)
			// detect Field:value (could be a classic function/operator call)
			if header == names.ListId {
				fieldId, _ := casted.LoadInt(1).(types.Identifier)
				dict := jen.Dict{jen.Id(string(fieldId)): extractCode(casted.LoadInt(2).Eval(env))}
				types.ForEach(itArgs, func(elem types.Object) bool {
					fieldDesc, _ := elem.(*types.List)
					fieldId, _ := fieldDesc.LoadInt(1).(types.Identifier)
					dict[jen.Id(string(fieldId))] = extractCode(fieldDesc.LoadInt(2).Eval(env))
					return true
				})
				argsCode = []jen.Code{dict}
			} else {
				argsCode = []jen.Code{extractCode(arg0.Eval(env))}
				argsCodeTemp := compileToCodeSlice(env, itArgs)
				argsCode = append(argsCode, argsCodeTemp...)
			}
		}
		// no more appliable ("type{A:a}{B:b}" is not valid)
		return wrapper{Renderer: casted.Clone().Values(argsCode...)}
	}
	return w
}

// all unknown Identifier eval return a wrapped jen.Id
// (the wrapper is a function call form appliable)
type compileEnvironment struct {
	types.Environment
}

func (c compileEnvironment) LoadStr(key string) (types.Object, bool) {
	res, ok := c.Environment.LoadStr(key)
	if !ok {
		res = appliableWrapper{Renderer: jen.Id(key)}
	}
	return res, true
}

func (c compileEnvironment) Load(key types.Object) types.Object {
	return types.Load(c, key)
}
