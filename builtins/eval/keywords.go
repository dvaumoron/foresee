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

import "github.com/dvaumoron/foresee/types"

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

func caseForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

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

func fallthroughForm(env types.Environment, itArgs types.Iterator) types.Object {
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

func genTypeForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func getForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
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

func listFunc(env types.Environment, itArgs types.Iterator) types.Object {
	return types.NewList().AddAll(makeEvalIterator(itArgs, env))
}

func literalForm(env types.Environment, itArgs types.Iterator) types.Object {
	return initStructAppliable
}

func macroForm(env types.Environment, itArgs types.Iterator) types.Object {
	// TODO

	return types.None
}

func mapTypeForm(env types.Environment, itArgs types.Iterator) types.Object {
	return initMapAppliable
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
