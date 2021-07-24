package main

import (
	"context"
	"geerpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
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
	var foo Foo
	if err := geerpc.Register(&foo); err != nil {
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
	_ = http.Serve(l, geerpc.DefaultServer)
	//geerpc.Accept(l)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

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

	client, _ := geerpc.DialHTTP("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			if err := client.Call(ctx, "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
		wg.Wait()
	}

	// reflect test
	// var foo Foo
	// log.Printf("test=%s", reflect.TypeOf(&foo))
	// var wg sync.WaitGroup
	// typ := reflect.TypeOf(&wg)
	// for i := 0; i < typ.NumMethod(); i++ {
	// 	method := typ.Method(i)
	// 	argv := make([]string, 0, method.Type.NumIn())
	// 	returns := make([]string, 0, method.Type.NumOut())
	// 	// j 从 1 开始，第 0 个入参是 wg 自己
	// 	log.Printf("func first param=%s", method.Type.In(0).String())
	// 	for j := 1; j < method.Type.NumIn(); j++ {
	// 		argv = append(argv, method.Type.In(j).Name())
	// 	}
	// 	for j := 0; j < method.Type.NumOut(); j++ {
	// 		returns = append(returns, method.Type.Out(j).Name())
	// 	}
	// 	log.Printf("func (w *%s) %s(%s) %s",
	// 		typ.Elem().Name(),
	// 		method.Name,
	// 		strings.Join(argv, ","),
	// 		strings.Join(returns, ","))
	// }
}
