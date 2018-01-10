package api

import "fmt"

// application specific errors
var (
	Overflow = fmt.Errorf("overflow")
)

type TwoReq struct {
	X int
	Y int
}

type OneRsp struct {
	Z int
}
