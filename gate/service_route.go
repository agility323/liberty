package main

import (
	"github.com/agility323/liberty/lbtproto"
)

type routeInfo struct {
	suggested int32
	allowed int32
}

var serviceRouteInfo = map[string]map[string]routeInfo {
	"avatar_service": map[string]routeInfo {
		// method: {suggested, allowed}
		"link_avatar": routeInfo{lbtproto.RouteTypeHash, lbtproto.RouteTypeHash | lbtproto.RouteTypeSpecific},
	},
}

func getServiceRouteType(service, method string, rt int32, rp []byte) int32 {
	// default
	if rt == 0 { return lbtproto.DefaultRouteType }
	// limit
	if m, ok := serviceRouteInfo[service]; ok {
		if info, ok := m[method]; ok {
			if rt & info.allowed == 0 { rt = info.suggested }
		}
	}
	// check
	if len(rp) == 0 && (rt & (lbtproto.RouteTypeHash | lbtproto.RouteTypeSpecific)) > 0 {
		return lbtproto.DefaultRouteType
	}
	return rt
}
