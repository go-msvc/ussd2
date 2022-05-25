package ussd

import (
	"context"

	"github.com/jansemmelink/utils2/errors"
)

//functions cannot be defined after start as it refers to a code function
//which is known on startup, so there is no DynFunc()
func Func(id string, fnc func(context.Context) ([]Item, error)) ItemSvc {
	if started {
		panic(errors.Errorf("attempt to define static item Func(%s) after started", id))
	}
	if id == "" || fnc == nil {
		panic(errors.Errorf("Func(%s,%p())", id, fnc))
	}
	f := ussdFunc{
		id:  id,
		fnc: fnc,
	}
	itemByID[id] = f
	return f
}

type ussdFunc struct {
	id  string
	fnc func(context.Context) ([]Item, error)
}

func (f ussdFunc) ID() string { return f.id }

func (f ussdFunc) Exec(ctx context.Context) ([]Item, error) {
	next, err := f.fnc(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "item(%s) func failed", f.id)
	}
	return next, nil
}
