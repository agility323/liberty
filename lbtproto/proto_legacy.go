/*
compatible with legacy protos
*/
package lbtproto

var (
	legacyMethodToIndex = map[string]map[string]uint16{
		"ClientGateType": map[string]uint16{
			"Method_requestEncryptToken": 1,
			"Method_confirmEncryptKey": 2,
			"Method_connectServer":  3,
			"Method_entityMessage": 4,
			"Method_channelMessage": 5,
		},
		"ClientType": map[string]uint16{
			"Method_responseEncryptToken": 10,
			"Method_confirmEncryptKeyAck": 11,
			"Method_connectResponse": 1,
			"Method_entityMessage": 2,
			"Method_channelMessage": 3,
			"Method_createChannelEntity": 4,
		},
	}
)
