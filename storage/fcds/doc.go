// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

// Package fcds provides storage layers for storing chunk data only.
//
// FCDS stands for Fixed Chunk Data Storage.
//
// Swarm Chunk data limited size property allows a very specific chunk storage
// solution that can be more performant than more generalized key/value
// databases. FCDS stores chunk data in files (shards) at fixed length offsets.
// Relations between chunk address, file number and offset in that file are
// managed by a separate MetaStore implementation.
//
// Package fcds contains the main implementation based on simple file operations
// for persisting chunk data and relaying on specific chunk meta information
// storage.
//
// The reference chunk meta information storage is implemented in fcds/mem
// package. It can be used in tests.
//
// LevelDB based chunk meta information storage is implemented in fcds/leveldb
// package. This implementation should be used as default in production.
//
// Additional FCDS Store implementation is in fcds/mock. It uses mock store and
// can be used for centralized chunk storage options that mock storage package
// provides.
//
// Package fcds/test contains test functions which can be used to validate
// behaviour of different FCDS or its MetaStore implementations.
package fcds
