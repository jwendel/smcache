// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package smcache is an implementation of the Cache
// within acme autocert that will store data within Google
// Cloud's Secret Manager.
//
// It uses the Google created GRPC client to
// communicate with the Secret Manager API, which allows the autocert
// library to Get/Put/Detelete certificates within Secret Manager.
//
// For more details, see the README.md, which is published at
// https://github.com/jwendel/smcache
package smcache
