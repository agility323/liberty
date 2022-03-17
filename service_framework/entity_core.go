package service_framework

import (
	"github.com/agility323/liberty/lbtutil"
)

type EntityCore struct {
	typ string
	id lbtutil.ObjectId
}

func (ec *EntityCore) init(typ string) {
	ec.id = lbtutil.NewObjectId()
	ec.typ = typ
}

func (ec *EntityCore) GetType() string {
	return ec.typ
}

func (ec *EntityCore) GetId() lbtutil.ObjectId {
	return ec.id
}

func (ec *EntityCore) Dump() map[string]string {
	return map[string]string {
		"id": string(ec.id),
		"typ": ec.typ,
	}
}
