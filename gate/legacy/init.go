package legacy

import (
	"strconv"
	"os"
	"fmt"

	"github.com/agility323/liberty/lbtutil"
)

type LegacyDependency struct {
	ConnectServerService string
	ConnectServerMethod string
	ConnectServerEntity string
	LegacyRouteTypeMap map[string]int32
	ServiceAddrGetter func(string) string
	ServiceSender func(string, []byte)
	ServiceRequestHandler func([]byte) bool
	PrivateRsaKey string
	AtService func() bool
}

var dep LegacyDependency

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "legacy")

func InitLegacyDependency(ld LegacyDependency) error {
	dep = ld
	if err := InitRsaKey(dep.PrivateRsaKey); err != nil {
		return fmt.Errorf("InitRsaKey fail %s\n\t%v", dep.PrivateRsaKey, err)
	}
	return nil
}
