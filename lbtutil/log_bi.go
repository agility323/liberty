package lbtutil

import (
	"encoding/json"
	"fmt"
	"time"
)

func BICoreLog(event string, trackInfo map[string]interface{}){
	logBI("[CORE]", event, trackInfo)
}
func BICustomLog(event string, trackInfo map[string]interface{}){
	logBI("[CUSTOM]", event, trackInfo)
}

func logBI(BTag, event string, trackingInfo map[string]interface{}) {
	trackingBytes, err := json.Marshal(trackingInfo)
	if err != nil {
		log.Error("BI %s %s [%+v], error:%v", BTag, event, trackingInfo, err.Error())
		return
	}
	trackingStr := fmt.Sprintf("[%s] [%s]", event, trackingBytes)

	now := time.Now()
	tstr := fmt.Sprintf("[%04d-%02d-%02d %02d:%02d:%02d.%03d]",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1e6)

	bPrefix := "[BI]"
	n := len(tstr) + 1 + len(bPrefix) + 1 + len(BTag) + 1 + len(trackingStr) + 1
	b := make([]byte, n, n)
	pos := 0
	var sep byte = ' '
	copy(b[pos:], tstr)
	pos += len(tstr)
	b[pos] = sep
	pos += 1
	copy(b[pos:], bPrefix)
	pos += len(bPrefix)
	b[pos] = sep
	pos += 1
	copy(b[pos:], BTag)
	pos += len(BTag)
	b[pos] = sep
	pos += 1
	copy(b[pos:], trackingStr)
	pos += len(trackingStr)
	b[pos] = '\n'

	logOut.Write(b)
}