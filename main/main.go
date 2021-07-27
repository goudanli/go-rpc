package main

import (
	"context"
	"geerpc"
	"geerpc/xclient"
	"log"
	"net"
	"net/http"
	"sync"
)

type Foo int

type Args struct {
	Num1, Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addr chan string) {
	srver := geerpc.NewServer()
	var foo Foo
	if err := srver.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	//geerpc.HandleHTTP()
	addr <- l.Addr().String()
	_ = http.Serve(l, srver)
	//geerpc.Accept(l)
}

func call(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"http@" + addr1, "http@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			//foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			var reply int
			var err error
			args := &Args{Num1: i, Num2: i * i}
			err = xc.Call(context.Background(), "Foo.Sum", args, &reply)
			if err != nil {
				log.Printf("call Foo.Sum error: %v", err)
			} else {
				log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
			}
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch1 := make(chan string)
	ch2 := make(chan string)
	go startServer(ch1)
	go startServer(ch2)
	addr1 := <-ch1
	addr2 := <-ch2
	call(addr1, addr2)
	//in fact, following code is like a simple geerpc client
	// conn, _ := net.Dial("tcp", <-addr)
	// defer func() { _ = conn.Close() }()

	// time.Sleep(time.Second)
	// // send options
	// _ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
	// cc := codec.NewGobCodec(conn)

	// // send request & receive response
	// for i := 0; i < 5; i++ {
	// 	h := &codec.Header{
	// 		ServiceMethod: "Foo.Sum",
	// 		Seq:           uint64(i),
	// 	}
	// 	_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
	// 	_ = cc.ReadHeader(h)
	// 	var reply string
	// 	_ = cc.ReadBody(&reply)
	// 	log.Println("reply:", reply)
	// }

	// client, _ := geerpc.DialHTTP("tcp", addr1)
	// defer func() { _ = client.Close() }()

	// time.Sleep(time.Second)
	// var wg sync.WaitGroup
	// for i := 0; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		args := &Args{Num1: i, Num2: i * i}
	// 		var reply int
	// 		ctx, _ := context.WithTimeout(context.Background(), time.Second)
	// 		if err := client.Call(ctx, "Foo.Sum", args, &reply); err != nil {
	// 			log.Fatal("call Foo.Sum error:", err)
	// 		}
	// 		log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
	// 	}(i)
	// 	wg.Wait()
	// }
}
