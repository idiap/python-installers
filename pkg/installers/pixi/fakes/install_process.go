// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package fakes

import "sync"

type InstallProcess struct {
	TranslateArchitectureCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			DependencyArch string
		}
		Returns struct {
			Arch string
		}
		Stub func(string) string
	}

	ExecuteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			DestLayerPath  string
			SrcLayerPath   string
			DependencyArch string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, string) error
	}
}

func (f *InstallProcess) TranslateArchitecture(param1 string) string {
	f.TranslateArchitectureCall.mutex.Lock()
	defer f.TranslateArchitectureCall.mutex.Unlock()
	f.TranslateArchitectureCall.CallCount++
	f.TranslateArchitectureCall.Receives.DependencyArch = param1
	if f.TranslateArchitectureCall.Stub != nil {
		return f.TranslateArchitectureCall.Stub(param1)
	}
	return f.TranslateArchitectureCall.Returns.Arch
}

func (f *InstallProcess) Execute(param1 string, param2 string, param3 string) error {
	f.ExecuteCall.mutex.Lock()
	defer f.ExecuteCall.mutex.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.DestLayerPath = param1
	f.ExecuteCall.Receives.SrcLayerPath = param2
	f.ExecuteCall.Receives.DependencyArch = param3
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2, param3)
	}
	return f.ExecuteCall.Returns.Error
}
