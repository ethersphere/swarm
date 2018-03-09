package main

import (
	"bytes"
	"fmt"
	"github.com/2tvenom/cbor"
)

//custom struct
type Vector struct {
	X, Y, Z int
	Range   []Range
	Label   string
}

type Range struct {
	Length int
	Align  float32
}

func main() {
	v := &Vector{
		X: 10,
		Y: 15,
		Z: 100,
		Range: []Range{
			Range{1, 10},
			Range{223432423, 30},
			Range{3, 41.5},
			Range{174, 55555.2},
		},
		Label: "HoHoHo",
	}

	//create encoder and marshal
	var buffTest bytes.Buffer
	encoder := cbor.NewEncoder(&buffTest)
	ok, error := encoder.Marshal(v)
	//check binary string
	if !ok {
		fmt.Errorf("Error decoding %s", error)
	} else {
		fmt.Printf("Variable Hex = % x\n", buffTest.Bytes())
		fmt.Printf("Variable = %v\n", buffTest.Bytes())
	}
	fmt.Printf("-----------------\n")

	//unmarshal binary string to new struct
	var vd Vector
	ok, err := encoder.Unmarshal(buffTest.Bytes(), &vd)

	if !ok {
		fmt.Printf("Error Unmarshal %s", err)
		return
	}
	//output
	fmt.Printf("%v", vd)
}
