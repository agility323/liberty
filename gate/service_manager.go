package main

import (
	"sync"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtreg"
	"github.com/agility323/liberty/lbtutil"

	"github.com/howeyc/crc16"
)

const (
	ServiceStateInit = iota
	ServiceStateConnected
	ServiceStateBanned
)

type serviceEntry struct {
	state int
	addr string
	typ string
	cli *lbtnet.TcpClient
}

type ServiceManager struct {
	lock sync.RWMutex
	serviceMap map[string]*serviceEntry
	serviceTypeToAddrSet map[string]*lbtutil.SortedSetStr
}

var serviceManager *ServiceManager

func init() {
	serviceManager = &ServiceManager{
		serviceMap: make(map[string]*serviceEntry),
		serviceTypeToAddrSet: make(map[string]*lbtutil.SortedSetStr),
	}
}

func (m *ServiceManager) onStart() {
	tickmgr.AddTickJob(m.OnTick)
}

func (m *ServiceManager) OnTick() {
	m.lock.RLock()
	defer m.lock.RUnlock()
	logger.Info("service manager tick %v %v", m.serviceMap, m.serviceTypeToAddrSet)
}

func (m *ServiceManager) OnDiscoverService(services map[string][]byte) {
	for service, _ := range services {
		pair := lbtreg.SplitEtcdKey(service, 2)
		if len(pair) != 2 { continue }
		typ := pair[0]
		addr := pair[1]
		entry, isnew := m.assureServiceEntry(typ, addr)
		if isnew {
			entry.cli.SetReconnectTime(10)
			entry.cli.StartConnect(3)
			continue
		}
		if entry.state == ServiceStateInit {
			logger.Warn("still connecting to service %s", addr)
			continue
		}
		if entry.state == ServiceStateBanned { continue }
	}
}

func (m *ServiceManager) assureServiceEntry(typ, addr string) (*serviceEntry, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if entry, ok := m.serviceMap[addr]; ok {
		return entry, false
	}
	entry := &serviceEntry{
		state: ServiceStateInit,
		addr: addr,
		typ: typ,
		cli: lbtnet.NewTcpClient(addr, &ServiceConnectionHandler{}),
	}
	m.serviceMap[addr] = entry
	return entry, true
}

func (m *ServiceManager) serviceConnect(c *lbtnet.TcpConnection) {
	m.lock.Lock()
	defer m.lock.Unlock()

	addr := c.RemoteAddr()
	if entry, ok := m.serviceMap[addr]; ok {
		entry.state = ServiceStateConnected
		if _, ok = m.serviceTypeToAddrSet[entry.typ]; !ok {
			m.serviceTypeToAddrSet[entry.typ] = lbtutil.NewSortedSetStr(65536)
		}
		m.serviceTypeToAddrSet[entry.typ].Add(addr)
	} else {
		logger.Warn("service connected with no client %s", addr)
		c.Close()
	}
}

func (m *ServiceManager) serviceConnectFail(cli *lbtnet.TcpClient) {
	m.lock.Lock()
	defer m.lock.Unlock()

	addr := cli.RemoteAddr()
	if entry, ok := m.serviceMap[addr]; ok {
		if entry.state == ServiceStateConnected { return }
		delete(m.serviceMap, addr)
	}
}

func (m *ServiceManager) serviceDisconnect(c *lbtnet.TcpConnection) {
	m.lock.Lock()
	defer m.lock.Unlock()

	addr := c.RemoteAddr()
	if entry, ok := m.serviceMap[addr]; ok {
		m.serviceTypeToAddrSet[entry.typ].Remove(addr)
		delete(m.serviceMap, addr)
		clientManager.serviceDisconnect(addr)
	}
}

func (m *ServiceManager) getServiceEntry(addr string) *serviceEntry {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if entry, ok := m.serviceMap[addr]; ok { return entry }
	return nil
}

func (m *ServiceManager) sendToService(addr string, buf []byte) {
	entry := m.getServiceEntry(addr)
	if entry == nil {
		logger.Warn("send to service fail 1 %s", addr)
		return
	}
	if entry.state != ServiceStateConnected {
		logger.Warn("send to service fail 2 %s %d", addr, entry.state)
		return
	}
	if err := entry.cli.SendData(buf); err != nil {
		logger.Error("send to service fail 3 %s [%v]", addr, err)
	} else {
		logger.Debug("send to service done %s", addr)
	}
}

func (m *ServiceManager) serviceRequest(buf []byte) bool {
	msg := &lbtproto.ServiceRequest{}
	if err := lbtproto.DecodeMessage(buf, msg); err != nil {
		logger.Warn("service request fail 1")
		return false
	}
	// route
	rt := getServiceRouteType(msg.Type, msg.Method, msg.Routet, msg.Routep)
	entries := m.getServiceEntriesByRoute(msg.Type, rt, msg.Routep)
	if entries == nil {
		logger.Warn("service request fail 2 %v", msg)
		return false
	}
	for _, entry := range entries {
		if err := entry.cli.SendData(buf); err == nil {
			logger.Debug("service request sent to %s", entry.addr)
		} else {
			logger.Error("service request fail 3 at %s [%v]", entry.addr, err)
		}
	}
	return true
}

func (m *ServiceManager) getServiceEntriesByRoute(typ string, rt int32, rp []byte) []*serviceEntry {
	m.lock.RLock()
	defer m.lock.RUnlock()

	// specific
	if rt == lbtproto.RouteTypeSpecific {
		addr := string(rp)
		if entry, ok := m.serviceMap[addr]; ok && entry.state == ServiceStateConnected {
			return []*serviceEntry{entry}
		}
		return nil
	}
	// type based
	addrSet, ok := m.serviceTypeToAddrSet[typ]
	if !ok {
		return nil
	}
	if rt & lbtproto.RouteTypeRandomOne > 0 {
		addr := addrSet.RandomGet()
		if addr == "" {
			return nil
		}
		if entry, ok := m.serviceMap[addr]; ok && entry.state == ServiceStateConnected {
			return []*serviceEntry{entry}
		}
		return nil
	} else if rt & lbtproto.RouteTypeHash > 0 {
		h := uint64(crc16.Checksum(rp, crc16.IBMTable))
		addr := addrSet.HashGet(h)
		if addr == "" {
			return nil
		}
		if entry, ok := m.serviceMap[addr]; ok && entry.state == ServiceStateConnected {// maybe should try next entry
			return []*serviceEntry{entry}
		}
		return nil
	} else if rt == lbtproto.RouteTypeAll {
		addrs := addrSet.GetAll()
		entries := make([]*serviceEntry, 0)
		for _, addr := range addrs {
			if entry, ok := m.serviceMap[addr]; ok && entry.state == ServiceStateConnected {
				entries = append(entries, entry)
			}
		}
		if len(entries) > 0 {
			return entries
		}
		return nil
	}
	return nil
}

func (m *ServiceManager) banService(typ, addr string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if entry, ok := m.serviceMap[addr]; ok {
		entry.state = ServiceStateBanned
		m.serviceTypeToAddrSet[typ].Remove(addr)
	}
	logger.Info("ban service %s %s", typ, addr)
}

func (m *ServiceManager) getAllServiceClients() []*lbtnet.TcpClient {
	m.lock.Lock()
	defer m.lock.Unlock()

	n := len(m.serviceMap)
	clients := make([]*lbtnet.TcpClient, n, n)
	i := 0
	for _, entry := range m.serviceMap {
		clients[i] = entry.cli
		i++
	}
	return clients
}

func (m *ServiceManager) notifyGateStop() {
	b, err := makeGateStopData()
	if err != nil {
		logger.Warn("notify gate stop fail encode [%v]", err)
		return
	}
	clients := m.getAllServiceClients()
	for _, cli := range clients {
		if err = cli.SendData(b); err != nil {
			logger.Warn("notify gate stop fail send %s [%v]", cli.RemoteAddr(), err)
		}
	}
}

func (m *ServiceManager) onServiceStop(typ, addr string) {
	m.banService(typ, addr)
}
