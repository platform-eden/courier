package courier

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
)

// localIP returns the ip address this node is currently using
func localIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// startmessagingServer starts the message server on a given port
func startMessagingServer(ctx context.Context, wg *sync.WaitGroup, server *grpc.Server, lis net.Listener, port string) {
	errchan := make(chan error, 1)
	done := make(chan struct{})
	defer close(errchan)
	defer wg.Done()

	go func() {
		err := server.Serve(lis)
		if err != nil {
			errchan <- err
		}
		defer close(done)
	}()

	select {
	case <-ctx.Done():
		server.GracefulStop()
		<-done
	case err := <-errchan:
		log.Print(err.Error())
		os.Exit(1)
	}
}
