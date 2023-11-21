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

package types

import "io"

type Loadable interface {
	Load(Object) Object
}

type Storable interface {
	Loadable
	Store(Object, Object)
}

type StringLoadable interface {
	LoadStr(string) (Object, bool)
}

type Environment interface {
	Object
	Storable
	Delete(Object)
	StringLoadable
	StoreStr(string, Object)
	DeleteStr(string)
	CopyTo(Environment)
}

type Object interface {
	io.WriterTo
	Eval(Environment) Object
}

type Sizable interface {
	Size() int
}

type Iterator interface {
	Iterable
	Next() (Object, bool)
	Close()
}

type Iterable interface {
	Object
	Iter() Iterator
}

type Appliable interface {
	Object
	Apply(Environment, Iterable) Object
}
