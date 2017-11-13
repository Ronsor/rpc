
package main
import "log"
import "github.com/ronsor/rpc"
func main() {
	log.Println("RPC")
	a, b := rpc.NewLoopbackRPC()
	a.Export("hi", func(g int) { 
		log.Printf("%d", g)
		log.Println("Hello")
	})
	a.Start()
	b.Start()
	log.Println(b.Call("hi", 42))
}
