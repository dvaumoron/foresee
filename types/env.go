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

import (
	"iter"
)

// Accept only Identifier key when used as Storable.
type BaseEnvironment struct {
	NoneType
	objects map[string]Object
}

func (b BaseEnvironment) LoadStr(key string) (Object, bool) {
	res, ok := b.objects[key]
	if !ok {
		return None, false
	}
	return res, true
}

func Load(env StringLoadable, key Object) Object {
	if id, ok := key.(Identifier); ok {
		res, _ := env.LoadStr(string(id))
		return res
	}
	return None
}

func (b BaseEnvironment) Load(key Object) Object {
	return Load(b, key)
}

func (b BaseEnvironment) Store(key Object, value Object) {
	if id, ok := key.(Identifier); ok {
		b.StoreStr(string(id), value)
	}
}

func (b BaseEnvironment) StoreStr(key string, value Object) {
	b.objects[key] = value
}

func (b BaseEnvironment) Delete(key Object) {
	if id, ok := key.(Identifier); ok {
		b.DeleteStr(string(id))
	}
}

func (b BaseEnvironment) DeleteStr(key string) {
	delete(b.objects, key)
}

func (b BaseEnvironment) CopyTo(other Environment) {
	for key, value := range b.objects {
		other.StoreStr(key, value)
	}
}

func (b BaseEnvironment) Size() int {
	return len(b.objects)
}

func (b BaseEnvironment) pushIter(yield func(Object) bool) {
	for key, value := range b.objects {
		if !yield(NewList(String(key), value)) {
			break
		}
	}
}

func (b BaseEnvironment) Iter() Iterator {
	next, stop := iter.Pull(b.pushIter)

	return &pullIteratorWrapper{next: next, close: stop}
}

func MakeBaseEnvironment() BaseEnvironment {
	return BaseEnvironment{objects: map[string]Object{}}
}

type LocalEnvironment struct {
	BaseEnvironment
	parent Environment
}

func (l LocalEnvironment) LoadStr(key string) (Object, bool) {
	res, ok := l.BaseEnvironment.LoadStr(key)
	if ok {
		return res, true
	}
	return l.parent.LoadStr(key)
}

func (l LocalEnvironment) Load(key Object) Object {
	return Load(l, key)
}

func MakeLocalEnvironment(env Environment) LocalEnvironment {
	return LocalEnvironment{BaseEnvironment: MakeBaseEnvironment(), parent: env}
}
