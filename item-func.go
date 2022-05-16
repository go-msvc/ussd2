package ussd

import (
	"context"

	"bitbucket.org/vservices/utils/errors"
)

func Func(id string, fnc func(context.Context) ([]Item, error)) ItemSvc {
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
