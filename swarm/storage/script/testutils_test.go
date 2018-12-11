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
