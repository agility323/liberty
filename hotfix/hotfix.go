package hotfix

import (
	"os"
	"strconv"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/hotfix/itf"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")

type HotfixImpl struct {}

var Hotfix *HotfixImpl

func (h *HotfixImpl) ApplyHotfix(entries []itf.HotfixEntry) {
	logger.Info("apply hotfix begin %d", len(entries))
	for _, entry := range entries {
		entry.Apply()
	}
	logger.Info("apply hotfix end")
}
