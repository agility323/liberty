package lbtproto

const (
	RouteTypeRandomOne int32 = 1 << iota
	RouteTypeHash
	RouteTypeSpecific
	RouteTypeAll
)

const DefaultRouteType = RouteTypeRandomOne
