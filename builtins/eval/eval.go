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
	"github.com/dvaumoron/foresee/parser"
	"github.com/dvaumoron/foresee/types"
)

var Builtins = initBuitins()

func initBuitins() types.BaseEnvironment {
	noOpAppliable := types.MakeNativeAppliable(noOp)

	base := types.MakeBaseEnvironment()
	base.StoreStr(names.Package, noOpAppliable)

	// TODO

	// give parser package a protected copy to use in user custom rules
	parser.BuiltinsCopy = types.MakeLocalEnvironment(base)
	return base
}

func noOp(env types.Environment, itArgs types.Iterator) types.Object {
	return types.None
}
