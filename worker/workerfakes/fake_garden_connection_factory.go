// This file was generated by counterfeiter
package workerfakes

import (
	"sync"

	"code.cloudfoundry.org/garden/client/connection"
	"github.com/concourse/atc/worker"
)

type FakeGardenConnectionFactory struct {
	BuildConnectionStub        func() connection.Connection
	buildConnectionMutex       sync.RWMutex
	buildConnectionArgsForCall []struct{}
	buildConnectionReturns     struct {
		result1 connection.Connection
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGardenConnectionFactory) BuildConnection() connection.Connection {
	fake.buildConnectionMutex.Lock()
	fake.buildConnectionArgsForCall = append(fake.buildConnectionArgsForCall, struct{}{})
	fake.recordInvocation("BuildConnection", []interface{}{})
	fake.buildConnectionMutex.Unlock()
	if fake.BuildConnectionStub != nil {
		return fake.BuildConnectionStub()
	} else {
		return fake.buildConnectionReturns.result1
	}
}

func (fake *FakeGardenConnectionFactory) BuildConnectionCallCount() int {
	fake.buildConnectionMutex.RLock()
	defer fake.buildConnectionMutex.RUnlock()
	return len(fake.buildConnectionArgsForCall)
}

func (fake *FakeGardenConnectionFactory) BuildConnectionReturns(result1 connection.Connection) {
	fake.BuildConnectionStub = nil
	fake.buildConnectionReturns = struct {
		result1 connection.Connection
	}{result1}
}

func (fake *FakeGardenConnectionFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.buildConnectionMutex.RLock()
	defer fake.buildConnectionMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeGardenConnectionFactory) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ worker.GardenConnectionFactory = new(FakeGardenConnectionFactory)
