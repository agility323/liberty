package main

const (
	RouteTypeRandomOne int32 = 1 << iota
	RouteTypeHash
	RouteTypeSpecific
	RouteTypeAll
)

const defaultRouteType = RouteTypeRandomOne

type routeInfo struct {
	suggested int32
	allowed int32
}

var serviceRouteInfo = map[string]map[string]routeInfo {
	"avatar_service": map[string]routeInfo {
		"link_avatar": routeInfo{RouteTypeHash, RouteTypeHash | RouteTypeSpecific},
	},
}

func getServiceRouteType(service, method string, rt int32, rp []byte) int32 {
	// default
	if rt == 0 { return defaultRouteType }
	// limit
	if m, ok := serviceRouteInfo[service]; ok {
		if info, ok := m[method]; ok {
			if rt & info.allowed == 0 { rt = info.suggested }
		}
	}
	// check
	if len(rp) == 0 && (rt & (RouteTypeHash | RouteTypeSpecific)) > 0 { return defaultRouteType }
	return rt
}
