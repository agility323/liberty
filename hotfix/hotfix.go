package hotfix

import (
	"os"
	"strconv"
	"sync"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/hotfix/itf"

	"github.com/agiledragon/gomonkey/v2"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "sf")

type HotfixImpl struct {
	lock sync.Mutex
	patches *gomonkey.Patches
}

var Hotfix = &HotfixImpl{}

func (h *HotfixImpl) NewFuncEntry(target interface{}, name string, double interface{}) itf.HotfixEntry {
	return &FuncEntry{target: target, name: name, double: double}
}

func (h *HotfixImpl) ApplyHotfix(entries []itf.HotfixEntry) {
	h.lock.Lock()
	defer h.lock.Unlock()

	logger.Info("apply hotfix begin %d", len(entries))
	if h.patches == nil {
		h.patches = gomonkey.NewPatches()
	} else {
		h.patches.Reset()
		h.patches = gomonkey.NewPatches()
	}
	for _, entry := range entries {
		entry.Apply()
	}
	logger.Info("apply hotfix end")
}

func (h *HotfixImpl) ResetHotfix() {
	h.lock.Lock()
	defer h.lock.Unlock()

	logger.Info("reset hotfix begin")
	if h.patches != nil {
		h.patches.Reset()
	}
	logger.Info("reset hotfix end")
}
