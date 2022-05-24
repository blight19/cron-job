package main

import (
	"dbsmonitor/trainning/client/register"
	"dbsmonitor/trainning/client/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	server := grpc.NewServer()
	helloService := new(service.HelloService)
	service.RegisterOrderManagementServer(server, helloService)
	listen, err := net.Listen("tcp", "127.0.0.1:8001")
	if err != nil {
		log.Fatalln(err)
	}
	etcdRegister, err := register.NewEtcdRegister([]string{"127.0.0.1:2379"})
	if err != nil {
		log.Println(err)
		return
	}
	defer etcdRegister.Close()
	serviceName := "order-service-1"
	addr := "127.0.0.1:8001"
	err = etcdRegister.RegisterServer("/etct/"+serviceName, addr, 5)
	if err != nil {
		log.Printf("register error %v \n", err)
		return
	}
	server.Serve(listen)

}
