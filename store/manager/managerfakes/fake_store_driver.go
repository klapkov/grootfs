// Code generated by counterfeiter. DO NOT EDIT.
package managerfakes

import (
	"sync"

	"code.cloudfoundry.org/grootfs/store/manager"
	"code.cloudfoundry.org/lager"
)

type FakeStoreDriver struct {
	ConfigureStoreStub        func(lager.Logger, string, string, int, int) error
	configureStoreMutex       sync.RWMutex
	configureStoreArgsForCall []struct {
		arg1 lager.Logger
		arg2 string
		arg3 string
		arg4 int
		arg5 int
	}
	configureStoreReturns struct {
		result1 error
	}
	configureStoreReturnsOnCall map[int]struct {
		result1 error
	}
	DeInitFilesystemStub        func(lager.Logger, string) error
	deInitFilesystemMutex       sync.RWMutex
	deInitFilesystemArgsForCall []struct {
		arg1 lager.Logger
		arg2 string
	}
	deInitFilesystemReturns struct {
		result1 error
	}
	deInitFilesystemReturnsOnCall map[int]struct {
		result1 error
	}
	InitFilesystemStub        func(lager.Logger, string, string) error
	initFilesystemMutex       sync.RWMutex
	initFilesystemArgsForCall []struct {
		arg1 lager.Logger
		arg2 string
		arg3 string
	}
	initFilesystemReturns struct {
		result1 error
	}
	initFilesystemReturnsOnCall map[int]struct {
		result1 error
	}
	MountFilesystemStub        func(lager.Logger, string, string) error
	mountFilesystemMutex       sync.RWMutex
	mountFilesystemArgsForCall []struct {
		arg1 lager.Logger
		arg2 string
		arg3 string
	}
	mountFilesystemReturns struct {
		result1 error
	}
	mountFilesystemReturnsOnCall map[int]struct {
		result1 error
	}
	ValidateFileSystemStub        func(lager.Logger, string) error
	validateFileSystemMutex       sync.RWMutex
	validateFileSystemArgsForCall []struct {
		arg1 lager.Logger
		arg2 string
	}
	validateFileSystemReturns struct {
		result1 error
	}
	validateFileSystemReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeStoreDriver) ConfigureStore(arg1 lager.Logger, arg2 string, arg3 string, arg4 int, arg5 int) error {
	fake.configureStoreMutex.Lock()
	ret, specificReturn := fake.configureStoreReturnsOnCall[len(fake.configureStoreArgsForCall)]
	fake.configureStoreArgsForCall = append(fake.configureStoreArgsForCall, struct {
		arg1 lager.Logger
		arg2 string
		arg3 string
		arg4 int
		arg5 int
	}{arg1, arg2, arg3, arg4, arg5})
	fake.recordInvocation("ConfigureStore", []interface{}{arg1, arg2, arg3, arg4, arg5})
	fake.configureStoreMutex.Unlock()
	if fake.ConfigureStoreStub != nil {
		return fake.ConfigureStoreStub(arg1, arg2, arg3, arg4, arg5)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.configureStoreReturns
	return fakeReturns.result1
}

func (fake *FakeStoreDriver) ConfigureStoreCallCount() int {
	fake.configureStoreMutex.RLock()
	defer fake.configureStoreMutex.RUnlock()
	return len(fake.configureStoreArgsForCall)
}

func (fake *FakeStoreDriver) ConfigureStoreCalls(stub func(lager.Logger, string, string, int, int) error) {
	fake.configureStoreMutex.Lock()
	defer fake.configureStoreMutex.Unlock()
	fake.ConfigureStoreStub = stub
}

func (fake *FakeStoreDriver) ConfigureStoreArgsForCall(i int) (lager.Logger, string, string, int, int) {
	fake.configureStoreMutex.RLock()
	defer fake.configureStoreMutex.RUnlock()
	argsForCall := fake.configureStoreArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *FakeStoreDriver) ConfigureStoreReturns(result1 error) {
	fake.configureStoreMutex.Lock()
	defer fake.configureStoreMutex.Unlock()
	fake.ConfigureStoreStub = nil
	fake.configureStoreReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) ConfigureStoreReturnsOnCall(i int, result1 error) {
	fake.configureStoreMutex.Lock()
	defer fake.configureStoreMutex.Unlock()
	fake.ConfigureStoreStub = nil
	if fake.configureStoreReturnsOnCall == nil {
		fake.configureStoreReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.configureStoreReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) DeInitFilesystem(arg1 lager.Logger, arg2 string) error {
	fake.deInitFilesystemMutex.Lock()
	ret, specificReturn := fake.deInitFilesystemReturnsOnCall[len(fake.deInitFilesystemArgsForCall)]
	fake.deInitFilesystemArgsForCall = append(fake.deInitFilesystemArgsForCall, struct {
		arg1 lager.Logger
		arg2 string
	}{arg1, arg2})
	fake.recordInvocation("DeInitFilesystem", []interface{}{arg1, arg2})
	fake.deInitFilesystemMutex.Unlock()
	if fake.DeInitFilesystemStub != nil {
		return fake.DeInitFilesystemStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.deInitFilesystemReturns
	return fakeReturns.result1
}

func (fake *FakeStoreDriver) DeInitFilesystemCallCount() int {
	fake.deInitFilesystemMutex.RLock()
	defer fake.deInitFilesystemMutex.RUnlock()
	return len(fake.deInitFilesystemArgsForCall)
}

func (fake *FakeStoreDriver) DeInitFilesystemCalls(stub func(lager.Logger, string) error) {
	fake.deInitFilesystemMutex.Lock()
	defer fake.deInitFilesystemMutex.Unlock()
	fake.DeInitFilesystemStub = stub
}

func (fake *FakeStoreDriver) DeInitFilesystemArgsForCall(i int) (lager.Logger, string) {
	fake.deInitFilesystemMutex.RLock()
	defer fake.deInitFilesystemMutex.RUnlock()
	argsForCall := fake.deInitFilesystemArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeStoreDriver) DeInitFilesystemReturns(result1 error) {
	fake.deInitFilesystemMutex.Lock()
	defer fake.deInitFilesystemMutex.Unlock()
	fake.DeInitFilesystemStub = nil
	fake.deInitFilesystemReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) DeInitFilesystemReturnsOnCall(i int, result1 error) {
	fake.deInitFilesystemMutex.Lock()
	defer fake.deInitFilesystemMutex.Unlock()
	fake.DeInitFilesystemStub = nil
	if fake.deInitFilesystemReturnsOnCall == nil {
		fake.deInitFilesystemReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deInitFilesystemReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) InitFilesystem(arg1 lager.Logger, arg2 string, arg3 string) error {
	fake.initFilesystemMutex.Lock()
	ret, specificReturn := fake.initFilesystemReturnsOnCall[len(fake.initFilesystemArgsForCall)]
	fake.initFilesystemArgsForCall = append(fake.initFilesystemArgsForCall, struct {
		arg1 lager.Logger
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	fake.recordInvocation("InitFilesystem", []interface{}{arg1, arg2, arg3})
	fake.initFilesystemMutex.Unlock()
	if fake.InitFilesystemStub != nil {
		return fake.InitFilesystemStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.initFilesystemReturns
	return fakeReturns.result1
}

func (fake *FakeStoreDriver) InitFilesystemCallCount() int {
	fake.initFilesystemMutex.RLock()
	defer fake.initFilesystemMutex.RUnlock()
	return len(fake.initFilesystemArgsForCall)
}

func (fake *FakeStoreDriver) InitFilesystemCalls(stub func(lager.Logger, string, string) error) {
	fake.initFilesystemMutex.Lock()
	defer fake.initFilesystemMutex.Unlock()
	fake.InitFilesystemStub = stub
}

func (fake *FakeStoreDriver) InitFilesystemArgsForCall(i int) (lager.Logger, string, string) {
	fake.initFilesystemMutex.RLock()
	defer fake.initFilesystemMutex.RUnlock()
	argsForCall := fake.initFilesystemArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeStoreDriver) InitFilesystemReturns(result1 error) {
	fake.initFilesystemMutex.Lock()
	defer fake.initFilesystemMutex.Unlock()
	fake.InitFilesystemStub = nil
	fake.initFilesystemReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) InitFilesystemReturnsOnCall(i int, result1 error) {
	fake.initFilesystemMutex.Lock()
	defer fake.initFilesystemMutex.Unlock()
	fake.InitFilesystemStub = nil
	if fake.initFilesystemReturnsOnCall == nil {
		fake.initFilesystemReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.initFilesystemReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) MountFilesystem(arg1 lager.Logger, arg2 string, arg3 string) error {
	fake.mountFilesystemMutex.Lock()
	ret, specificReturn := fake.mountFilesystemReturnsOnCall[len(fake.mountFilesystemArgsForCall)]
	fake.mountFilesystemArgsForCall = append(fake.mountFilesystemArgsForCall, struct {
		arg1 lager.Logger
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	fake.recordInvocation("MountFilesystem", []interface{}{arg1, arg2, arg3})
	fake.mountFilesystemMutex.Unlock()
	if fake.MountFilesystemStub != nil {
		return fake.MountFilesystemStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.mountFilesystemReturns
	return fakeReturns.result1
}

func (fake *FakeStoreDriver) MountFilesystemCallCount() int {
	fake.mountFilesystemMutex.RLock()
	defer fake.mountFilesystemMutex.RUnlock()
	return len(fake.mountFilesystemArgsForCall)
}

func (fake *FakeStoreDriver) MountFilesystemCalls(stub func(lager.Logger, string, string) error) {
	fake.mountFilesystemMutex.Lock()
	defer fake.mountFilesystemMutex.Unlock()
	fake.MountFilesystemStub = stub
}

func (fake *FakeStoreDriver) MountFilesystemArgsForCall(i int) (lager.Logger, string, string) {
	fake.mountFilesystemMutex.RLock()
	defer fake.mountFilesystemMutex.RUnlock()
	argsForCall := fake.mountFilesystemArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeStoreDriver) MountFilesystemReturns(result1 error) {
	fake.mountFilesystemMutex.Lock()
	defer fake.mountFilesystemMutex.Unlock()
	fake.MountFilesystemStub = nil
	fake.mountFilesystemReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) MountFilesystemReturnsOnCall(i int, result1 error) {
	fake.mountFilesystemMutex.Lock()
	defer fake.mountFilesystemMutex.Unlock()
	fake.MountFilesystemStub = nil
	if fake.mountFilesystemReturnsOnCall == nil {
		fake.mountFilesystemReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.mountFilesystemReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) ValidateFileSystem(arg1 lager.Logger, arg2 string) error {
	fake.validateFileSystemMutex.Lock()
	ret, specificReturn := fake.validateFileSystemReturnsOnCall[len(fake.validateFileSystemArgsForCall)]
	fake.validateFileSystemArgsForCall = append(fake.validateFileSystemArgsForCall, struct {
		arg1 lager.Logger
		arg2 string
	}{arg1, arg2})
	fake.recordInvocation("ValidateFileSystem", []interface{}{arg1, arg2})
	fake.validateFileSystemMutex.Unlock()
	if fake.ValidateFileSystemStub != nil {
		return fake.ValidateFileSystemStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.validateFileSystemReturns
	return fakeReturns.result1
}

func (fake *FakeStoreDriver) ValidateFileSystemCallCount() int {
	fake.validateFileSystemMutex.RLock()
	defer fake.validateFileSystemMutex.RUnlock()
	return len(fake.validateFileSystemArgsForCall)
}

func (fake *FakeStoreDriver) ValidateFileSystemCalls(stub func(lager.Logger, string) error) {
	fake.validateFileSystemMutex.Lock()
	defer fake.validateFileSystemMutex.Unlock()
	fake.ValidateFileSystemStub = stub
}

func (fake *FakeStoreDriver) ValidateFileSystemArgsForCall(i int) (lager.Logger, string) {
	fake.validateFileSystemMutex.RLock()
	defer fake.validateFileSystemMutex.RUnlock()
	argsForCall := fake.validateFileSystemArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeStoreDriver) ValidateFileSystemReturns(result1 error) {
	fake.validateFileSystemMutex.Lock()
	defer fake.validateFileSystemMutex.Unlock()
	fake.ValidateFileSystemStub = nil
	fake.validateFileSystemReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) ValidateFileSystemReturnsOnCall(i int, result1 error) {
	fake.validateFileSystemMutex.Lock()
	defer fake.validateFileSystemMutex.Unlock()
	fake.ValidateFileSystemStub = nil
	if fake.validateFileSystemReturnsOnCall == nil {
		fake.validateFileSystemReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.validateFileSystemReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeStoreDriver) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.configureStoreMutex.RLock()
	defer fake.configureStoreMutex.RUnlock()
	fake.deInitFilesystemMutex.RLock()
	defer fake.deInitFilesystemMutex.RUnlock()
	fake.initFilesystemMutex.RLock()
	defer fake.initFilesystemMutex.RUnlock()
	fake.mountFilesystemMutex.RLock()
	defer fake.mountFilesystemMutex.RUnlock()
	fake.validateFileSystemMutex.RLock()
	defer fake.validateFileSystemMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeStoreDriver) recordInvocation(key string, args []interface{}) {
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

var _ manager.StoreDriver = new(FakeStoreDriver)
