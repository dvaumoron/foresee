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
	errBooleanType    = errors.New("wait boolean value")
	errIdentifierType = errors.New("wait identifier type")
	errIntegerType    = errors.New("wait integer value")
	errListType       = errors.New("wait list type")
	errNumericType    = errors.New("wait numeric value")
	errObjectType     = errors.New("type without methods")
	errStringType     = errors.New("wait string value")
)

// Storable accepting all key type.
type dynamicMap struct {
	types.NoneType
	objects map[string]types.Object
}

func (d dynamicMap) Load(key types.Object) types.Object {
	return d.objects[extractRenderString(key)]
}

func (d dynamicMap) Store(key types.Object, value types.Object) {
	d.objects[extractRenderString(key)] = value
}

func makeDynamicMap() dynamicMap {
	return dynamicMap{objects: map[string]types.Object{}}
}

type dynamicObject struct {
	types.BaseEnvironment
	objectType customType
}

func curryMethod(d dynamicObject, method types.Appliable) types.NativeAppliable {
	return types.MakeNativeAppliable(func(env types.Environment, itArgs types.Iterator) types.Object {
		augmentedItArgs := types.NewList(d).AddAll(itArgs).Iter()
		defer augmentedItArgs.Close()

		return method.Apply(env, augmentedItArgs)
	})
}

func (d dynamicObject) LoadStr(key string) (types.Object, bool) {
	if res, ok := d.BaseEnvironment.LoadStr(key); ok {
		return res, true
	}

	if method, ok := d.objectType.methods[key]; ok {
		return curryMethod(d, method), true
	}

	return types.None, false
}

func (d dynamicObject) Load(key types.Object) types.Object {
	return types.Load(d, key)
}

func makeDynamicObject(env types.Environment, typeName string) dynamicObject {
	customTypes, _ := env.LoadStr(hiddenTypesName)
	castedTypes, _ := customTypes.(types.BaseEnvironment)
	objectType, _ := castedTypes.LoadStr(typeName)
	castedType, _ := objectType.(customType)

	return dynamicObject{BaseEnvironment: types.MakeBaseEnvironment(), objectType: castedType}
}

type customType struct {
	types.NoneType
	methods map[string]types.Appliable
}
