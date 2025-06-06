// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package fakes

import "sync"

type SitePackageProcess struct {
	ExecuteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			TargetLayerPath string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string) (string, error)
	}
}

func (f *SitePackageProcess) Execute(param1 string) (string, error) {
	f.ExecuteCall.mutex.Lock()
	defer f.ExecuteCall.mutex.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.TargetLayerPath = param1
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1)
	}
	return f.ExecuteCall.Returns.String, f.ExecuteCall.Returns.Error
}
