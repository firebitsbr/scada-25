package iec101

import (
	"scada/points"
	"io"
)

type IECer interface {
	C_SC_NA_1(conn io.ReadWriter, point *points.Point) (err error)
}
