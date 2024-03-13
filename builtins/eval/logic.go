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
	"github.com/dvaumoron/foresee/types"
)

func boolOperatorForm(env types.Environment, itArgs types.Iterator, defaultB bool) types.Object {
	allBool := true
	res := types.Boolean(defaultB)
	var temp types.Boolean
	types.ForEach(itArgs, func(arg types.Object) bool {
		temp, allBool = arg.Eval(env).(types.Boolean)
		if temp != res {
			res = temp

			return false
		}

		return allBool
	})

	if !allBool {
		return types.None
	}

	return res
}
