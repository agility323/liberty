package lbtproto

// using gogofaster

import (
	"fmt"
	"reflect"
	"os"
	"strconv"

	grpc "google.golang.org/grpc"

	"github.com/agility323/liberty/lbtutil"
)

var logger = lbtutil.NewLogger(strconv.Itoa(os.Getpid()), "lbtproto")

// client - gate
type ClientGateType struct {
	Method_connectServer,
	Method_entityMessage,
	Method_channelMessage,

	Method_service_request,
	Method_entity_msg uint16
}

// gate - service
type ServiceType struct {
	Method_register_reply,
	Method_client_disconnect,
	Method_service_request,
	Method_entity_msg,
	Method_service_shutdown uint16
}

// service - gate
type ServiceGateType struct {
	Method_register_service,
	Method_bind_client,
	Method_service_reply,
	Method_create_entity,
	Method_entity_msg uint16
}

// gate - client
type ClientType struct {
	Method_connectResponse,
	Method_createChannelEntity,
	Method_entityMessage,
	Method_channelMessage,

	Method_service_reply,
	Method_create_entity,
	Method_entity_msg uint16
}

var (
	ClientGate ClientGateType = ClientGateType{}
	Service ServiceType = ServiceType{}
	ServiceGate ServiceGateType = ServiceGateType{}
	Client ClientType = ClientType{}
)

type serviceDescPair struct {
	service interface {}
	desc *grpc.ServiceDesc
}
var serviceDesc []serviceDescPair = []serviceDescPair{
	serviceDescPair{&ClientGate, &_IClientGate_serviceDesc,},
	serviceDescPair{&Service, &_IService_serviceDesc},
	serviceDescPair{&ServiceGate, &_IServiceGate_serviceDesc},
	serviceDescPair{&Client, &_IClient_serviceDesc},
}

const commonIndexBase uint16 = 1000
var commonIndexMap = make(map[string]reflect.Value)

func init() {
	initServiceMethodIndex()
	checkServiceMethodIndex()
}

func initServiceMethodIndex() {
	// handle legacy index
	legacyIndexToMethod := make(map[string]map[uint16]string)
	for tname, methodToIndex := range legacyMethodToIndex {
		legacyIndexToMethod[tname] = make(map[uint16]string)
		for method, index := range methodToIndex {
			legacyIndexToMethod[tname][index] = method
		}
	}
	// handle common index
	cibase := commonIndexBase
	// generate index
	for _, pair := range serviceDesc {
		pservice := pair.service
		pdesc := pair.desc
		val := reflect.ValueOf(pservice).Elem()
		typ := val.Type()
		tname := typ.Name()
		nf := val.NumField()
		if nf >= int(cibase) {
			panic(fmt.Sprintf("number of fields exceeeds cibase in %s", pdesc.ServiceName))
		}
		fnameToIndex := make(map[string]uint16)
		indexToFname := make([]string, nf, nf)
		for i := 0; i < nf; i++ {
			name := typ.Field(i).Name
			fnameToIndex[name] = uint16(i)
			indexToFname[i] = name
		}
		for i := 0; i < nf; i++ {	// guarantee order with i
			method := indexToFname[i]
			fval := val.Field(i)
			// 1. legacy method is set with fixed index (legacyMethodToIndex)
			methodToIndex := legacyMethodToIndex[tname]
			if index, ok := methodToIndex[method]; ok {
				fval.SetUint(uint64(index))
				logger.Info("init service method index: %s.%s %d", pdesc.ServiceName, method, fval.Uint())
				continue
			}
			// 2. methods with same name among services, use same index
			if cival, ok := commonIndexMap[method]; ok {
				index := uint16(cival.Uint())
				old := index
				if index < commonIndexBase {
					index = cibase
					cibase += 1
					cival.SetUint(uint64(index))
					logger.Info("service method index change: %s from %d to %d", method, old, index)
				}
				fval.SetUint(uint64(index))
				logger.Info("init service method index: %s.%s %d", pdesc.ServiceName, method, fval.Uint())
				continue
			}
			// 3. i conlict with legacy field index, use conflicted index
			indexToMethod := legacyIndexToMethod[tname]
			if conflictedMethod, ok := indexToMethod[uint16(i)]; ok {
				index := fnameToIndex[conflictedMethod]
				fval.SetUint(uint64(index))
				commonIndexMap[method] = fval
				logger.Info("init service method index: %s.%s %d", pdesc.ServiceName, method, fval.Uint())
				continue
			}
			// 4. normally set with field order,
			fval.SetUint(uint64(i))
			commonIndexMap[method] = fval
			logger.Info("init service method index: %s.%s %d", pdesc.ServiceName, method, fval.Uint())
		}
	}
}

func checkServiceMethodIndex() {
	for _, pair := range serviceDesc {
		pservice := pair.service
		pdesc := pair.desc
		val := reflect.ValueOf(pservice).Elem()
		method2Index := legacyMethodToIndex[val.Type().Name()]
		index2Method := make(map[uint16]string)
		for method, index := range method2Index { index2Method[index] = method }
		for i, entry := range pdesc.Methods {
			method := "Method_" + entry.MethodName
			idxToCheck := uint16(val.FieldByName(method).Uint())
			var wrong uint16
			// 1. legacy method
			if index, ok := method2Index[method]; ok && idxToCheck == index {
				continue
			} else { wrong = index }
			// skip conflicted index
			if _, ok := index2Method[uint16(i)]; ok {
				continue
			}
			// 2. indices with same name
			index := uint16(commonIndexMap[method].Uint())
			isCommonIndex := (index >= commonIndexBase)
			if isCommonIndex && idxToCheck == index {
				continue
			} else { wrong = index }
			// 3. normal index
			if !isCommonIndex && idxToCheck == uint16(i) {
				continue
			} else { wrong = uint16(i)}
			panic(fmt.Sprintf("wrong method index: %s.%s %d!=%d",
				pdesc.ServiceName, method, idxToCheck, wrong))
		}
		logger.Info("check service method index pass: %s", pdesc.ServiceName)
	}
}
