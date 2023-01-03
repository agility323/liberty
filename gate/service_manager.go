package main

import (
	"sync"

	"github.com/agility323/liberty/lbtnet"
	"github.com/agility323/liberty/lbtproto"
	"github.com/agility323/liberty/lbtutil"
	"github.com/agility323/liberty/lbtreg"

	"github.com/howeyc/crc16"
)

type serviceEntry struct {
	connected bool
	addr string
	typ string
	cli *lbtnet.TcpClient
}

type ServiceManager struct {
	lock sync.RWMutex
	serviceMap map[string]*serviceEntry
	serviceTypeToAddrSet map[string]*lbtutil.OrderedSet
}

var serviceManager *ServiceManager

func init() {
	serviceManager = &ServiceManager{
		serviceMap: make(map[string]*serviceEntry),
		serviceTypeToAddrSet: make(map[string]*lbtutil.OrderedSet),
	}
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
		if !entry.connected {
			logger.Warn("still connecting to service %s", addr)
			continue
		}
	}
}

func (m *ServiceManager) assureServiceEntry(typ, addr string) (*serviceEntry, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if entry, ok := m.serviceMap[addr]; ok {
		return entry, false
	}
	entry := &serviceEntry{
		connected: false,
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
		entry.connected = true
		if _, ok = m.serviceTypeToAddrSet[entry.typ]; !ok {
			m.serviceTypeToAddrSet[entry.typ] = lbtutil.NewOrderedSet()
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
		if entry.connected{ return }
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
	if !entry.connected {
		logger.Warn("send to service fail 2 %s", addr)
		return
	}
	if err := entry.cli.SendData(buf); err != nil {
		logger.Warn("send to service fail 3 %s %v", addr, err)
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
			logger.Warn("service request fail 3 at %s %v", entry.addr, err)
		}
	}
	return true
}

func (m *ServiceManager) getServiceEntriesByRoute(typ string, rt int32, rp []byte) []*serviceEntry {
	m.lock.RLock()
	defer m.lock.RUnlock()

	// specific
	if rt == RouteTypeSpecific {
		addr := string(rp)
		if entry, ok := m.serviceMap[addr]; ok && entry.connected {
			return []*serviceEntry{entry}
		}
		return nil
	}
	// type based
	addrSet, ok := m.serviceTypeToAddrSet[typ]
	if !ok {
		return nil
	}
	if rt & RouteTypeRandomOne > 0 {
		v := addrSet.RandomGetOne()
		if v == nil {
			return nil
		}
		addr := v.(string)
		if entry, ok := m.serviceMap[addr]; ok && entry.connected {
			return []*serviceEntry{entry}
		}
		return nil
	} else if rt & RouteTypeHash > 0 {
		h := int(crc16.Checksum(rp, crc16.IBMTable))
		vs := addrSet.GetAll()	// TODO service sort with id number
		if len(vs) == 0 {
			return nil
		}
		v := vs[h % len(vs)]
		addr := v.(string)
		if entry, ok := m.serviceMap[addr]; ok && entry.connected {	// maybe should try next entry
			return []*serviceEntry{entry}
		}
		return nil
	} else if rt == RouteTypeAll {
		vs := addrSet.GetAll()
		entries := make([]*serviceEntry, 0)
		for _, v := range vs {
			addr := v.(string)
			if entry, ok := m.serviceMap[addr]; ok && entry.connected {
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
