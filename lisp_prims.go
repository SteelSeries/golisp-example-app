package main

import (
	"errors"
	. "github.com/steelseries/golisp"
)

func init() {
	Global.BindTo(SymbolWithName("CONSTANT"), FloatWithValue(float32(42.0)))
	MakePrimitiveFunction("go-fact", "1", GoFactImpl)
}

func GoFactImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	val, err := Eval(Car(args), env)
	if err != nil {
		return
	}
	if !FloatP(val) {
		return nil, errors.New("go-fact requires a float argument")
	}
	n := int(FloatValue(val))
	f := 1
	for i := 1; i <= n; i++ {
		f *= i
	}
	return FloatWithValue(float32(f)), nil
}
