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
	PostServiceManagerJob func (op string, jd interface{}) bool
	LegacyRouteTypeMap map[string]int32
	ServiceAddrGetter func(string) string
	PrivateRsaKey string
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
