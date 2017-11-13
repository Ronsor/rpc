package rpc

import (
//	"time"
	"fmt"
	"github.com/orcaman/concurrent-map"
	"os"
	"os/exec"
	"time"
	"errors"
	"github.com/vmihailenco/msgpack"
	"io"
	"encoding/binary"
)

type event struct {
	id int64
	ret []interface{}
}
type call struct {
	ID string
	Name string
	Args []interface{}
}
type RPC struct {
	Exports map[string]interface{}
	events cmap.ConcurrentMap
	mutex bool
	R io.Reader
	W io.Writer
	running bool
}
func NewRPC() (*RPC) {
	return &RPC{
		Exports: make(map[string]interface{}),
		events: cmap.New(),
		running: false,
	}
}
func (r *RPC) rawcall(id string, name string, args ...interface{}) {
	cstr := call{ID:id,Name:name,Args:args}
	dat, err := msgpack.Marshal(cstr)
	if err != nil { panic(err) }
	flen := make([]byte, 4)
	//println(len(dat))
	binary.LittleEndian.PutUint32(flen, uint32(len(dat)))
	r.W.Write(flen)
	r.W.Write(dat)	
}
func (r *RPC) Call(name string, args ...interface{}) []interface{} {
	//println(name)
	id := fmt.Sprintf("%x", time.Now().UnixNano()^int64(len(name)+len(args)))
	//for r.mutex {time.Sleep(1*time.Millisecond)}
	//r.mutex = true
	r.events.Set(id, make(chan call, 1))
	//r.mutex = false
	r.rawcall(id, name, args...)
	//for r.mutex {time.Sleep(1*time.Millisecond)}
	retx, _ := r.events.Get(id)
	ret := <- retx.(chan call)
	//r.mutex = true
	r.events.Remove(id)
	//r.mutex = false
	return ret.Args
}
func (r *RPC) Export(name string, f interface{}) {
	r.Exports[name] = f
}
func (r *RPC) Start() {
	go func() {
		r.running = true
		pktlen := make([]byte, 4)
		for {
			_, err := io.ReadFull(r.R, pktlen)
			if err != nil { break }
			len := int(binary.LittleEndian.Uint32(pktlen))
			//println(len)
			buf := make([]byte, len)
			rlen, err := io.ReadFull(r.R, buf)
			if rlen != len || err != nil { break }
			var inc call
			err = msgpack.Unmarshal(buf, &inc)
			if err != nil { break }
			//println(inc.Name)
			if inc.Name == "*" {
				if r.events.Has(inc.ID) {
					x, _ := r.events.Get(inc.ID); x.(chan call) <- inc
				}
			} else {
				if r.Exports[inc.Name] != nil {
					go func() {
						ret, err := inCall(r.Exports[inc.Name], inc.Args...)
						if err != nil { panic(err); ret = []interface{}{err} }
						r.rawcall(inc.ID, "*", ret...)
					} ()
				} else {
					r.rawcall(inc.ID, "*", errors.New("Not exported"))
				}
			}
		}
		//println("done")
		r.running = false
	} ()
}
func NewCommandRPC(c *exec.Cmd) *RPC {
	a := NewRPC()
	a.R, _ = c.StdoutPipe()
	r, w := io.Pipe()
	a.W = w
	c.Stdin = r
	return a
}
func NewStdioRPC() *RPC {
	a := NewRPC()
	a.R = os.Stdin
	a.W = os.Stdout
	return a
}
func NewLoopbackRPC() (*RPC, *RPC) {
	a := NewRPC()
	b := NewRPC()
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()
	a.R = r1
	a.W = w2
	b.R = r2
	b.W = w1
	return a, b
}
