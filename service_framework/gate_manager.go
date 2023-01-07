package service_framework

import (
	"math/rand"
	"sync"

	"github.com/agility323/liberty/lbtnet"
)

type GateManager struct {
	lock sync.RWMutex
	gateMap map[string]*lbtnet.TcpConnection
	primaryGateAddr string
}

var gateManager *GateManager

func init() {
	gateManager = &GateManager{
		gateMap: make(map[string]*lbtnet.TcpConnection),
		primaryGateAddr: "",
	}
}

func (m *GateManager) gateConnect(c *lbtnet.TcpConnection) {
	addr := c.RemoteAddr()
	m.lock.Lock()
	defer m.lock.Unlock()
	m.gateMap[addr] = c
	logger.Info("gate connect %s", addr)
}

func (m *GateManager) gateDisconnect(c *lbtnet.TcpConnection) {
	//TODO.OnConnectionClose()
	addr := c.RemoteAddr()
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.gateMap, addr)
	if m.primaryGateAddr == addr { m.primaryGateAddr = "" }
	logger.Info("gate disconnect %s", addr)
}

func (m *GateManager) getPrimaryGate() *lbtnet.TcpConnection {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.gateMap) == 0 {
		return nil
	}
	c, ok := m.gateMap[m.primaryGateAddr]
	if ok && c != nil { return c }
	n := rand.Intn(len(m.gateMap))
	for addr, c := range m.gateMap {
		if n--; n >= 0 { continue }
		m.primaryGateAddr = addr
		return c
	}
	return nil
}

func (m *GateManager) getRandomGate() *lbtnet.TcpConnection {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.gateMap) == 0 {
		return nil
	}
	n := rand.Intn(len(m.gateMap))
	for _, c := range m.gateMap {
		if n--; n >= 0 { continue }
		return c
	}
	return nil
}

func (m *GateManager) getGateByAddr(addr string) *lbtnet.TcpConnection {
	m.lock.RLock()
	defer m.lock.RUnlock()

	c, _ := m.gateMap[addr]
	return c
}

func (m *GateManager) getAllGates() []*lbtnet.TcpConnection {
	m.lock.RLock()
	defer m.lock.RUnlock()

	gates := []*lbtnet.TcpConnection{}
	for _, c := range m.gateMap {
		gates = append(gates, c)
	}
	return gates
}
