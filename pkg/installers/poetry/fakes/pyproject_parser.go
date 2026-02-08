// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package fakes

import "sync"

type PoetryPyProjectParser struct {
	ParsePythonVersionCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			String string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string) (string, error)
	}

	IsPoetryProjectCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			String string
		}
		Returns struct {
			Bool  bool
			Error error
		}
		Stub func(string) (bool, error)
	}
}

func (f *PoetryPyProjectParser) ParsePythonVersion(param1 string) (string, error) {
	f.ParsePythonVersionCall.mutex.Lock()
	defer f.ParsePythonVersionCall.mutex.Unlock()
	f.ParsePythonVersionCall.CallCount++
	f.ParsePythonVersionCall.Receives.String = param1
	if f.ParsePythonVersionCall.Stub != nil {
		return f.ParsePythonVersionCall.Stub(param1)
	}
	return f.ParsePythonVersionCall.Returns.String, f.ParsePythonVersionCall.Returns.Error
}

func (f *PoetryPyProjectParser) IsPoetryProject(param1 string) (bool, error) {
	f.IsPoetryProjectCall.mutex.Lock()
	defer f.IsPoetryProjectCall.mutex.Unlock()
	f.IsPoetryProjectCall.CallCount++
	f.IsPoetryProjectCall.Receives.String = param1
	if f.IsPoetryProjectCall.Stub != nil {
		return f.IsPoetryProjectCall.Stub(param1)
	}
	return f.IsPoetryProjectCall.Returns.Bool, f.IsPoetryProjectCall.Returns.Error
}
