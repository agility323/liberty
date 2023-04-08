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
	filterData map[string]int32
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

func (m *ClientManager) onStart() {
	tickmgr.AddTickJob(m.OnTick)
}

func (m *ClientManager) OnTick() {
	n := 0
	for slot := range m.clientSlots {
		m.locks[slot].RLock()
		n += len(m.clientSlots[slot])
		m.locks[slot].RUnlock()
	}
	logger.Info("client manager tick %d", n)
}

func (m *ClientManager) clientConnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	m.clientSlots[slot][addr] = &clientEntry{c: c, serviceAddr: "", filterData: make(map[string]int32)}
	logger.Info("client connect %s", addr)
}

func (m *ClientManager) clientDisconnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	saddr := m.PopDisconnectedClient(slot, addr)
	if saddr != "" {
		// send client_disconnect to service
		info := lbtproto.BindClientInfo{Caddr: addr, Saddr: saddr}
		serviceEntry := serviceManager.getServiceEntry(saddr)	// service manager lock
		if serviceEntry != nil && serviceEntry.state == ServiceStateConnected {
			if err := lbtproto.SendMessage(serviceEntry.cli, lbtproto.Service.Method_client_disconnect, &info); err != nil {
				logger.Error("send client disconnect fail %v [%v]", info, err)
			}
		}
	}
	logger.Info("client disconnect %s %s", addr, saddr)
}

func (m *ClientManager) PopDisconnectedClient(slot int, addr string) string {
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	// delete client entry
	entry, ok := m.clientSlots[slot][addr]
	if !ok { return "" }
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
	}
	return saddr
}

func (m *ClientManager) bindClient(info lbtproto.BindClientInfo) {
	caddr := info.Caddr
	saddr := info.Saddr
	slot := lbtutil.StringHash(caddr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	entry, ok := m.clientSlots[slot][caddr]
	if !ok { return }
	entry.serviceAddr = saddr
	if _, ok := m.boundClientSlots[slot][saddr]; !ok {
		m.boundClientSlots[slot][saddr] = make(map[string]int)
	}
	m.boundClientSlots[slot][saddr][caddr] = 1
	logger.Info("client bind %s %s", caddr, saddr)
}

func (m *ClientManager) unbindClient(info lbtproto.BindClientInfo) {
	caddr := info.Caddr
	saddr := info.Saddr
	slot := lbtutil.StringHash(caddr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()

	if entry, ok := m.clientSlots[slot][caddr]; ok {
		delete(m.clientSlots[slot], caddr)
		entry.c.CloseWithoutCallback()
		saddr = entry.serviceAddr
	}
	if cm, ok := m.boundClientSlots[slot][saddr]; ok {
		delete(cm, caddr)
		if len(cm) == 0 {
			delete(m.boundClientSlots[slot], saddr)
		}
	}
	logger.Info("client unbind %s %s", caddr, saddr)
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
		if err := entry.c.SendData(data); err != nil {
			addr := "nil"
			if entry.c != nil { addr = entry.c.RemoteAddr() }
			logger.Warn("broadcast msg fail at %d %s [%v]", slot, addr, err)
		}
	}
}

func (m *ClientManager) SoftStop(ttl, tt time.Duration) <-chan bool {
	logger.Info("client manager soft stop begin")
	done := make(chan bool, 1)
	lbtactor.RunTask("ClientManager.SoftStop", func() {
		defer func() {
			done <- true
		}()
		ctx, cancel := context.WithTimeout(context.Background(), ttl * time.Second)
		defer cancel()
		ticker := time.NewTicker(tt * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("client manager soft stop timeout")
				return
			case <-ticker.C:
				n := 0
				for slot := 0; slot < ClientSlotNum; slot++ {
					n += m.getClientsNumBySlot(slot)
				}
				if n == 0 {
					logger.Info("client manager soft stop finish")
					return
				}
				logger.Info("client manager soft stop tick %d", n)
			}
		}
	})
	return done
}

func (m *ClientManager) getClientsNumBySlot(slot int) int {
	m.locks[slot].RLock()
	defer m.locks[slot].RUnlock()
	return len(m.clientSlots[slot])
}

func (m *ClientManager) setFilterData(addr string, filterData map[string]int32) {
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].RLock()
	defer m.locks[slot].RUnlock()
	entry, ok := m.clientSlots[slot][addr]
	if !ok { return }
	entry.filterData = filterData
}

func (m *ClientManager) updateFilterData(addr string, filterData map[string]int32) {
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()
	entry, ok := m.clientSlots[slot][addr]
	if !ok { return }
	for k, v := range filterData {
		entry.filterData[k] = v
	}
}

func (m *ClientManager) deleteFilterData(addr string, filterData map[string]int32) {
	slot := lbtutil.StringHash(addr) % ClientSlotNum
	m.locks[slot].Lock()
	defer m.locks[slot].Unlock()
	entry, ok := m.clientSlots[slot][addr]
	if !ok { return }
	if len(filterData) == 0 {
		entry.filterData = make(map[string]int32)
	} else {
		for k, _ := range filterData {
			delete(entry.filterData, k)
		}
	}
}

func (m *ClientManager) filterClients(filters []*lbtproto.Filter) [][]*lbtnet.TcpConnection {
	arr := make([][]*lbtnet.TcpConnection, 0)
	for i := 0; i < ClientSlotNum; i++ {
		if clients := m.filterClientsBySlot(i, filters); len(clients) > 0 {
			arr = append(arr, clients)
		}
	}
	return arr
}

func (m *ClientManager) filterClientsBySlot(slot int, filters []*lbtproto.Filter) []*lbtnet.TcpConnection {
	m.locks[slot].RLock()
	defer m.locks[slot].RUnlock()
	clients := make([]*lbtnet.TcpConnection, 0)
	for _, entry := range m.clientSlots[slot] {
		if checkFilter(filters, entry.filterData) {
			clients = append(clients, entry.c)
		}
	}
	return clients
}

func checkFilter(filters []*lbtproto.Filter, fdata map[string]int32) bool {
	for _, filter := range filters {
		attr := filter.Attr
		op := filter.Op
		val := filter.Val
		f, ok := filterOpFunc[op]
		if !ok {
			logger.Warn("no filter func for %s", op)
			return false
		}
		v, ok := fdata[attr]
		if !ok { return false }
		if !f(v, val) { return false }
	}
	return true
}

var filterOpFunc = map[string]func(int32, int32) bool {
	">": filterGreater,
	"<": filterLess,
	"=": filterEqual,
	">=": filterGreaterOrEqual,
	"<=": filterLessOrEqual,
}

func filterGreater(v, th int32) bool { return v > th }
func filterLess(v, th int32) bool { return v < th }
func filterEqual(v, th int32) bool { return v == th }
func filterGreaterOrEqual(v, th int32) bool { return v >= th }
func filterLessOrEqual(v, th int32) bool { return v <= th }
