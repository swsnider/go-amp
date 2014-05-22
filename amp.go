package amp

import (
	"fmt"
	"io"
	"net/rpc"
	"reflect"
	"strconv"
)

type serverCodec struct {
	conn io.ReadWriteCloser

	req Box
}

func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{}
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	box, err := Decode(c.conn)
	if err != nil {
		return err
	}
	r.ServiceMethod = box["_command"]
	seq, err := strconv.Atoi(box["_ask"])
	if err != nil {
		return err
	}
	r.Seq = uint64(seq)
	return nil
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	rv := reflect.ValueOf(x).Elem()
	rt := rv.Type()
	l := rv.NumField()
	for i := 0; i < l; i++ {
		f := rv.Field(i)
		if !f.IsValid() {
			continue
		}
		if !f.CanSet() {
			continue
		}
		nameTag := rt.Field(i).Tag.Get("amp")
		if nameTag == "" {
			nameTag = rt.Field(i).Name
		}
		if _, ok := c.req[nameTag]; !ok {
			continue
		}
		switch f.Type().Name() {
		case "string":
			f.SetString(c.req[nameTag])
		case "int":
			i, err := strconv.Atoi(c.req[nameTag])
			if err != nil {
				return err
			}
			f.SetInt(int64(i))
		case "float":
			fl, err := strconv.ParseFloat(c.req[nameTag], 64)
			if err != nil {
				return err
			}
			f.SetFloat(fl)
		case "bool":
			if c.req[nameTag] == "True" {
				f.SetBool(true)
			} else if c.req[nameTag] == "False" {
				f.SetBool(false)
			} else {
				return fmt.Errorf("%q is not a valid boolean.", c.req[nameTag])
			}
		}
	}
	return nil
}

func (c *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	return nil
}

func (c *serverCodec) Close() error {
	return nil
}
