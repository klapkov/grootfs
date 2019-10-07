// Code generated by counterfeiter. DO NOT EDIT.
package overlayxfsfakes

import (
	"sync"

	"code.cloudfoundry.org/grootfs/store/filesystems/overlayxfs"
)

type FakeUnmounter struct {
	IsRootlessStub        func() bool
	isRootlessMutex       sync.RWMutex
	isRootlessArgsForCall []struct {
	}
	isRootlessReturns struct {
		result1 bool
	}
	isRootlessReturnsOnCall map[int]struct {
		result1 bool
	}
	UnmountStub        func(string) error
	unmountMutex       sync.RWMutex
	unmountArgsForCall []struct {
		arg1 string
	}
	unmountReturns struct {
		result1 error
	}
	unmountReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeUnmounter) IsRootless() bool {
	fake.isRootlessMutex.Lock()
	ret, specificReturn := fake.isRootlessReturnsOnCall[len(fake.isRootlessArgsForCall)]
	fake.isRootlessArgsForCall = append(fake.isRootlessArgsForCall, struct {
	}{})
	fake.recordInvocation("IsRootless", []interface{}{})
	fake.isRootlessMutex.Unlock()
	if fake.IsRootlessStub != nil {
		return fake.IsRootlessStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.isRootlessReturns
	return fakeReturns.result1
}

func (fake *FakeUnmounter) IsRootlessCallCount() int {
	fake.isRootlessMutex.RLock()
	defer fake.isRootlessMutex.RUnlock()
	return len(fake.isRootlessArgsForCall)
}

func (fake *FakeUnmounter) IsRootlessCalls(stub func() bool) {
	fake.isRootlessMutex.Lock()
	defer fake.isRootlessMutex.Unlock()
	fake.IsRootlessStub = stub
}

func (fake *FakeUnmounter) IsRootlessReturns(result1 bool) {
	fake.isRootlessMutex.Lock()
	defer fake.isRootlessMutex.Unlock()
	fake.IsRootlessStub = nil
	fake.isRootlessReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeUnmounter) IsRootlessReturnsOnCall(i int, result1 bool) {
	fake.isRootlessMutex.Lock()
	defer fake.isRootlessMutex.Unlock()
	fake.IsRootlessStub = nil
	if fake.isRootlessReturnsOnCall == nil {
		fake.isRootlessReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.isRootlessReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeUnmounter) Unmount(arg1 string) error {
	fake.unmountMutex.Lock()
	ret, specificReturn := fake.unmountReturnsOnCall[len(fake.unmountArgsForCall)]
	fake.unmountArgsForCall = append(fake.unmountArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("Unmount", []interface{}{arg1})
	fake.unmountMutex.Unlock()
	if fake.UnmountStub != nil {
		return fake.UnmountStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.unmountReturns
	return fakeReturns.result1
}

func (fake *FakeUnmounter) UnmountCallCount() int {
	fake.unmountMutex.RLock()
	defer fake.unmountMutex.RUnlock()
	return len(fake.unmountArgsForCall)
}

func (fake *FakeUnmounter) UnmountCalls(stub func(string) error) {
	fake.unmountMutex.Lock()
	defer fake.unmountMutex.Unlock()
	fake.UnmountStub = stub
}

func (fake *FakeUnmounter) UnmountArgsForCall(i int) string {
	fake.unmountMutex.RLock()
	defer fake.unmountMutex.RUnlock()
	argsForCall := fake.unmountArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeUnmounter) UnmountReturns(result1 error) {
	fake.unmountMutex.Lock()
	defer fake.unmountMutex.Unlock()
	fake.UnmountStub = nil
	fake.unmountReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeUnmounter) UnmountReturnsOnCall(i int, result1 error) {
	fake.unmountMutex.Lock()
	defer fake.unmountMutex.Unlock()
	fake.UnmountStub = nil
	if fake.unmountReturnsOnCall == nil {
		fake.unmountReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.unmountReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeUnmounter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.isRootlessMutex.RLock()
	defer fake.isRootlessMutex.RUnlock()
	fake.unmountMutex.RLock()
	defer fake.unmountMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeUnmounter) recordInvocation(key string, args []interface{}) {
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

var _ overlayxfs.Unmounter = new(FakeUnmounter)
