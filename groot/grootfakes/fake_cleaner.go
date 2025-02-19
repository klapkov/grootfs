// Code generated by counterfeiter. DO NOT EDIT.
package grootfakes

import (
	"sync"

	"code.cloudfoundry.org/grootfs/groot"
	"code.cloudfoundry.org/lager"
)

type FakeCleaner struct {
	CleanStub        func(lager.Logger, int64) (bool, error)
	cleanMutex       sync.RWMutex
	cleanArgsForCall []struct {
		arg1 lager.Logger
		arg2 int64
	}
	cleanReturns struct {
		result1 bool
		result2 error
	}
	cleanReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCleaner) Clean(arg1 lager.Logger, arg2 int64) (bool, error) {
	fake.cleanMutex.Lock()
	ret, specificReturn := fake.cleanReturnsOnCall[len(fake.cleanArgsForCall)]
	fake.cleanArgsForCall = append(fake.cleanArgsForCall, struct {
		arg1 lager.Logger
		arg2 int64
	}{arg1, arg2})
	stub := fake.CleanStub
	fakeReturns := fake.cleanReturns
	fake.recordInvocation("Clean", []interface{}{arg1, arg2})
	fake.cleanMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeCleaner) CleanCallCount() int {
	fake.cleanMutex.RLock()
	defer fake.cleanMutex.RUnlock()
	return len(fake.cleanArgsForCall)
}

func (fake *FakeCleaner) CleanCalls(stub func(lager.Logger, int64) (bool, error)) {
	fake.cleanMutex.Lock()
	defer fake.cleanMutex.Unlock()
	fake.CleanStub = stub
}

func (fake *FakeCleaner) CleanArgsForCall(i int) (lager.Logger, int64) {
	fake.cleanMutex.RLock()
	defer fake.cleanMutex.RUnlock()
	argsForCall := fake.cleanArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeCleaner) CleanReturns(result1 bool, result2 error) {
	fake.cleanMutex.Lock()
	defer fake.cleanMutex.Unlock()
	fake.CleanStub = nil
	fake.cleanReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeCleaner) CleanReturnsOnCall(i int, result1 bool, result2 error) {
	fake.cleanMutex.Lock()
	defer fake.cleanMutex.Unlock()
	fake.CleanStub = nil
	if fake.cleanReturnsOnCall == nil {
		fake.cleanReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.cleanReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeCleaner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.cleanMutex.RLock()
	defer fake.cleanMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCleaner) recordInvocation(key string, args []interface{}) {
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

var _ groot.Cleaner = new(FakeCleaner)
