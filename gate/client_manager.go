package main

import (
	"context"
	"sync"
	"time"

	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtactor"
)

type clientEntry struct {
	c *lbtnet.TcpConnection
	serviceAddr string
}

const ClientSlotNum int = 128

type ClientManager struct {
	locks [ClientSlotNum]sync.RWMutex
	clientSlots [ClientSlotNum]map[string]*clientEntry
	boundClientSlots [ClientSlotNum]map[string]map[string]int	// []{saddr: {caddr: 1}}
}

var clientManager *ClientManager

func init() {
	clientManager = &ClientManager{}
	for i := range clientManager.clientSlots {
		clientManager.clientSlots[i] = make(map[string]*clientEntry)
	}
	for i := range clientManager.boundClientSlots {
		clientManager.boundClientSlots[i] = make(map[string]map[string]int)
	}
}

func (m *ClientManager) clientConnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	m.clientSlots[slot][addr] = &clientEntry{c: c, serviceAddr: ""}
}

func (m *ClientManager) clientDisconnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	// delete client entry
	entry, ok := m.clientSlots[slot][addr]
	if !ok { return }
	delete(m.clientSlots[slot], addr)
	saddr := entry.serviceAddr
	if saddr != "" {
		// delete bound client
		if cm, ok := m.boundClientSlots[slot][saddr]; ok {
			delete(cm, addr)
			if len(cm) == 0 {
				delete(m.boundClientSlots[slot], saddr)
			}
		}
		// send client_disconnect to service
		info := lbtproto.BindClientInfo{Caddr: addr, Saddr: saddr}
		serviceEntry := serviceManager.getServiceEntry(info.Saddr)
		if serviceEntry != nil && serviceEntry.state == ServiceStateConnected {
			lbtproto.SendMessage(serviceEntry.cli, lbtproto.Service.Method_client_disconnect, &info)
		}
	}
}

func (m *ClientManager) bindClient(info lbtproto.BindClientInfo) {
	addr := info.Caddr
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	entry, ok := m.clientSlots[slot][info.Caddr]
	if !ok { return }
	entry.serviceAddr = info.Saddr
	if _, ok := m.boundClientSlots[slot][info.Saddr]; !ok {
		m.boundClientSlots[slot][info.Saddr] = make(map[string]int)
	}
	m.boundClientSlots[slot][info.Saddr][info.Caddr] = 1
}

func (m *ClientManager) unbindClient(info lbtproto.BindClientInfo) {
	addr := info.Caddr
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	saddr := info.Saddr
	if entry, ok := m.clientSlots[slot][addr]; ok {
		delete(m.clientSlots[slot], addr)
		entry.c.CloseWithoutCallback()
		saddr = entry.serviceAddr
	}
	if cm, ok := m.boundClientSlots[slot][saddr]; ok {
		delete(cm, addr)
		if len(cm) == 0 {
			delete(m.boundClientSlots[slot], saddr)
		}
	}
}

func (m *ClientManager) getClientConnection(addr string) *lbtnet.TcpConnection {
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].RLock()
	defer m.locks[slot].RUnlock()

	if entry, ok := m.clientSlots[slot][addr]; ok {
		return entry.c
	}
	return nil
}

func (m *ClientManager) getClientServiceAddr(caddr string) string {
	slot := lbtutil.StringHash(caddr) % ClientSlotNum
	m.locks[slot].RLock()
	defer m.locks[slot].RUnlock()

	if entry, ok := m.clientSlots[slot][caddr]; ok {
		return entry.serviceAddr
	}
	return ""
}

func (m *ClientManager) serviceDisconnect(saddr string) {
	clientSlots := make(map[int]map[string]int)
	for i := 0; i < ClientSlotNum; i++ {
		cm := m.popServiceFromBoundClientSlot(saddr, i)
		if cm == nil { continue }
		for caddr, _ := range cm {
			slot := lbtutil.StringHash(caddr) % ClientSlotNum
			if _, ok := clientSlots[slot]; !ok {
				clientSlots[slot] = make(map[string]int)
			}
			clientSlots[slot][caddr] = 1
		}
	}
	for slot, cm := range clientSlots {
		m.removeClientsFromSlot(slot, cm)
	}
}

func (m *ClientManager) popServiceFromBoundClientSlot(saddr string, slot int) map[string]int {
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	cm, ok := m.boundClientSlots[slot][saddr]
	if !ok { return nil }
	delete(m.boundClientSlots[slot], saddr)
	return cm
}

func (m *ClientManager) removeClientsFromSlot(slot int, cm map[string]int) {
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	for caddr, _ := range cm {
		if entry, ok := m.clientSlots[slot][caddr]; ok {
			delete(m.clientSlots[slot], caddr)
			entry.c.CloseWithoutCallback()
		}
	}
}

func (m *ClientManager) broadcastMsg(data []byte) {
	for i := 0; i < ClientSlotNum; i++ {
		m.broadcastMsgBySlot(i, data)
	}
}

func (m *ClientManager) broadcastMsgBySlot(slot int, data []byte) {
	m.locks[slot].RLock()
	defer m.locks[slot].RUnlock()
	for _, entry := range m.clientSlots[slot] {
		entry.c.SendData(data)
	}
}

func (m *ClientManager) SoftStop() {
	// TODO avoid duplicates
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lbtactor.RunTaskActor(ctx, func() struct{} {
		ticker := time.NewTicker(50 * time.Second)
		stop := false
		defer func() {
			ticker.Stop()
			if stop { Stop() }
		}()

		for {
			select {
			case <-ctx.Done():
				return struct{}{}
			case <-ticker.C:
				n := 0
				for slot := 0; slot < ClientSlotNum; slot++ {
					n += m.getClientsNumBySlot(slot)
				}
				if n == 0 {
					stop = true
					return struct{}{}
				}
				logger.Info("soft stop tick client manager %d", n)
			}
		}
		return struct{}{}
	})
}

func (m *ClientManager) getClientsNumBySlot(slot int) int {
	m.locks[slot].RLock()
	defer m.locks[slot].RUnlock()
	return len(m.clientSlots[slot])
}
