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

import "sync"

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

func (b BaseEnvironment) Iter() Iterator {
	it := &chanIterator{channel: make(chan Object), done: make(chan NoneType)}
	go it.sendMapValue(b.objects)
	return it
}

type chanIterator struct {
	NoneType
	channel chan Object
	done    chan NoneType
	once    sync.Once
}

func (it *chanIterator) Iter() Iterator {
	return it
}

func (it *chanIterator) Next() (Object, bool) {
	value, ok := <-it.channel
	if !ok {
		return None, false
	}
	return value, true
}

func (it *chanIterator) Close() {
	it.once.Do(func() {
		close(it.done)
	})
}

func (it *chanIterator) sendMapValue(objects map[string]Object) {
ForLoop:
	for key, value := range objects {
		select {
		case it.channel <- NewList(String(key), value):
		case <-it.done:
			break ForLoop
		}
	}
	close(it.channel)
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
