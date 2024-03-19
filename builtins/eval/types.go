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
	"errors"

	"github.com/dvaumoron/foresee/types"
)

var (
	errIdentifierType = errors.New("wait identifier type")
	errListType       = errors.New("wait list type")
)

// Storable accepting all key type.
type dynamic struct {
	types.NoneType
	objects map[string]types.Object
}

func (d dynamic) Load(key types.Object) types.Object {
	return d.objects[extractString(key)]
}

func (d dynamic) Store(key types.Object, value types.Object) {
	d.objects[extractString(key)] = value
}

func makeDynamic() dynamic {
	return dynamic{objects: map[string]types.Object{}}
}
