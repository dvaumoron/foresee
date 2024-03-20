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

func appendForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO wrap native append behavior

	return types.None
}

func assertForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO add a matching test ?

	return evalFirstOp(env, itArgs)
}

func blockForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO manage break and continue ?
	types.ForEach(itArgs, func(o types.Object) bool {
		o.Eval((env))

		return true
	})

	return types.None
}

func breakForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func capForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO wrap native cap behavior

	return types.None
}

func caseForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func closeForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO wrap native close behavior

	return types.None
}

func constForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func continueForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func defaultForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func deferForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func deleteForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO wrap native delete behavior

	return types.None
}

func fallthroughForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func fileForm(env types.Environment, itArgs types.Iterator) types.Object {
	// init default value
	env.StoreStr(hiddenTypesName, types.MakeBaseEnvironment())

	// TODO

	return types.None
}

func forForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func funcForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func getForm(env types.Environment, itArgs types.Iterator) types.Object {
	res, _ := itArgs.Next()
	types.ForEach(itArgs, func(elem types.Object) bool {
		loadable, ok := res.(types.StringLoadable)
		if !ok {
			panic(errSelectableType)
		}

		id, ok := elem.(types.Identifier)
		if !ok {
			panic(errIdentifierType)
		}

		res, ok = loadable.LoadStr(string(id))
		if !ok {
			panic(errUnknownField)
		}

		return true
	})

	return res
}

func goForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func gotoForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func ifForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func importForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func labelForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func lambdaForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func lenForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO wrap native len behavior

	return types.None
}

func listFunc(env types.Environment, itArgs types.Iterator) types.Object {
	return types.NewList().AddAll(makeEvalIterator(itArgs, env))
}

func literalForm(env types.Environment, itArgs types.Iterator) types.Object {
	arg0, ok := itArgs.Next()
	if !ok {
		panic(errUnarySize)
	}

	typeName := extractTypeName(arg0)

	return types.MakeNativeAppliable(func(env types.Environment, itArgs types.Iterator) types.Object {
		args := types.NewList().AddAll(itArgs)
		if args.Size() != 1 {
			return initStruct(env, args, typeName)
		}

		arg1 := args.LoadInt(0)
		if casted, ok := arg1.(*types.List); ok {
			if id, _ := casted.LoadInt(0).(types.Identifier); id == names.ListId {
				return initStruct(env, args, typeName)
			}
		}

		casted, ok := arg1.Eval(env).(dynamicObject)
		if !ok {
			panic(errConversion)
		}

		// type conversion
		return copyStruct(env, casted, typeName)
	})

}

func macroForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func makeForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO wrap native make behavior

	return types.None
}

func mapTypeForm(env types.Environment, itArgs types.Iterator) types.Object {
	return initMapAppliable
}

func newForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO wrap native new behavior

	return types.None
}

func rangeForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func returnForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func selectForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func sliceOrArrayTypeForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func switchForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func typeForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func varForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}
