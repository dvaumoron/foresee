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
	_ types.Appliable = callableWrapper{}
	_ types.Appliable = literalWrapper{}
)

// match jen types
type Renderer interface {
	types.Renderer
	jen.Code
	Assert(jen.Code) *jen.Statement
	Dot(string) *jen.Statement
	Index(...jen.Code) *jen.Statement
	Op(string) *jen.Statement
}

// Augment jen types with Eval in order to match types.Object
type wrapper struct {
	Renderer
}

func (w wrapper) Eval(types.Environment) types.Object {
	return w
}

// Augment jen types with Eval in order to match types.Object
// and an Apply which create a function call
type callableWrapper struct {
	// not composing wrapper to avoid a type change on eval
	Renderer
}

func (w callableWrapper) Eval(types.Environment) types.Object {
	return w
}

func (w callableWrapper) Apply(env types.Environment, args types.Iterable) types.Object {
	if casted, ok := w.Renderer.(*jen.Statement); ok {
		argsCode := compileToCodeSlice(env, args)
		// still appliable ("f(a)(b)" is possible)
		return callableWrapper{Renderer: casted.Clone().Call(argsCode...)}
	}
	return w
}

// Augment jen types with Eval in order to match types.Object
// and an Apply which create a literal
type literalWrapper struct {
	// not composing wrapper to avoid a type change on eval
	Renderer
}

func (w literalWrapper) Eval(types.Environment) types.Object {
	return w
}

func (w literalWrapper) Apply(env types.Environment, args types.Iterable) types.Object {
	if casted, ok := w.Renderer.(*jen.Statement); ok {
		next, stop := types.Pull(args.Iter())
		defer stop()

		var argsCode []jen.Code
		if arg0, ok := next(); ok {
			casted2, _ := arg0.(*types.List)
			// detect Field:value (could be a classic function/operator call)
			if header, _ := casted2.LoadInt(0).(types.Identifier); header == names.ListId {
				dict := jen.Dict{compileToCode(env, casted2.LoadInt(1)): compileToCode(env, casted2.LoadInt(2))}
				for {
					elem, ok := next()
					if !ok {
						break
					}

					fieldDesc, _ := elem.(*types.List)
					dict[compileToCode(env, fieldDesc.LoadInt(1))] = compileToCode(env, fieldDesc.LoadInt(2))
				}
				argsCode = []jen.Code{dict}
			} else {
				argsCode = []jen.Code{compileToCode(env, arg0)}
				argsCodeTemp := compileToCodeSlice(env, types.Push(next))
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
		res = callableWrapper{Renderer: jen.Id(key)}
	}
	return res, true
}

func (c compileEnvironment) Load(key types.Object) types.Object {
	return types.Load(c, key)
}
