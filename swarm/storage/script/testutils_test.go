// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package script_test

import (
	"encoding/json"
	"reflect"
	"testing"
)

func JSONEquals(t *testing.T, expected, actual string) {
	//credit for the trick: turtlemonvh https://gist.github.com/turtlemonvh/e4f7404e28387fadb8ad275a99596f67
	var e interface{}
	var a interface{}

	err := json.Unmarshal([]byte(expected), &e)
	if err != nil {
		t.Fatalf("Error mashalling expected :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(actual), &a)
	if err != nil {
		t.Fatalf("Error mashalling actual :: %s", err.Error())
	}

	if !reflect.DeepEqual(e, a) {
		t.Fatalf("Error comparing JSON. Expected %s. Got %s", expected, actual)
	}
}
