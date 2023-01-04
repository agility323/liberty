package service_framework

import (
	"sync"

	"github.com/agility323/liberty/lbtutil"
)

var ccbmgr *ClientCallbackManager

func init() {
	ccbmgr = &ClientCallbackManager{}
	for slot := range ccbmgr.ccbSlots {
		ccbmgr.ccbSlots[slot] = make(map[string]ClientCallback)
	}
}

type ClientCallback interface {
	OnClientDisconnect()
}

const ClientCallbackSlotNum = 64

type ClientCallbackManager struct {
	locks [ClientCallbackSlotNum]sync.RWMutex
	ccbSlots [ClientCallbackSlotNum]map[string]ClientCallback
}

func (m *ClientCallbackManager) onStart() {
	tickmgr.AddTickJob(m.OnTick)
}

func (m *ClientCallbackManager) OnTick() {
	n := 0
	for slot := range m.ccbSlots {
		m.locks[slot].RLock()
		n += len(m.ccbSlots[slot])
		m.locks[slot].RUnlock()
	}
	logger.Info("ccb manager tick %d", n)
}

func registerClientCallback(caddr string, cb ClientCallback) {
	slot := lbtutil.StringHash(caddr) % ClientCallbackSlotNum
	ccbmgr.locks[slot].Lock()
	defer ccbmgr.locks[slot].Unlock()

	ccbmgr.ccbSlots[slot][caddr] = cb
}

func unregisterClientCallback(caddr string) {
	slot := lbtutil.StringHash(caddr) % ClientCallbackSlotNum
	ccbmgr.locks[slot].Lock()
	defer ccbmgr.locks[slot].Unlock()

	delete(ccbmgr.ccbSlots[slot], caddr)
}

func getClientCallback(caddr string) ClientCallback {
	slot := lbtutil.StringHash(caddr) % ClientCallbackSlotNum
	ccbmgr.locks[slot].RLock()
	defer ccbmgr.locks[slot].RUnlock()

	return ccbmgr.ccbSlots[slot][caddr]
}

func popClientCallback(caddr string) ClientCallback {
	slot := lbtutil.StringHash(caddr) % ClientCallbackSlotNum
	ccbmgr.locks[slot].Lock()
	defer ccbmgr.locks[slot].Unlock()

	cb, ok := ccbmgr.ccbSlots[slot][caddr]
	if ok { delete(ccbmgr.ccbSlots[slot], caddr) }
	return cb
}
