package messaging

// import (
// 	"context"
// 	"fmt"
// 	"math/rand"
// 	"net"
// 	"os"
// 	"sync"
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/platform-edn/courier/messaging/proto"
// 	"google.golang.org/grpc/test/bufconn"
// )

// var testMessageServer TestMessageServer

// type TestMessageServer interface {
// 	proto.MessageServerServer
// 	messageServer
// 	start(context.Context, *sync.WaitGroup) error
// 	BufDialer(context.Context, string) (net.Conn, error)
// 	SetToFail()
// 	SetToPass()
// 	Clear()
// 	MessagesLength() int
// 	ResponsesLength() int
// }

// func RemovePointers(nodes []*Node) []Node {
// 	updated := []Node{}

// 	for _, n := range nodes {
// 		updated = append(updated, *n)
// 	}

// 	return updated
// }

// // TestNodeOptions that can be passed into CreateTestNodes
// type TestNodeOptions struct {
// 	SubscribedSubjects  []string
// 	BroadcastedSubjects []string
// }

// // CreateTestNodes creates a quantity of randomized nodes based on the options passed in
// func CreateTestNodes(count int, options *TestNodeOptions) []*Node {
// 	nodes := []*Node{}
// 	var broadSubjects []string
// 	var subSubjects []string

// 	if len(options.SubscribedSubjects) == 0 {
// 		subSubjects = []string{"sub", "sub1", "sub2"}
// 	} else {
// 		subSubjects = options.SubscribedSubjects
// 	}
// 	if len(options.BroadcastedSubjects) == 0 {
// 		broadSubjects = []string{"broad", "broad1"}
// 	} else {
// 		broadSubjects = options.BroadcastedSubjects
// 	}

// 	for i := 0; i < count; i++ {
// 		ip := fmt.Sprintf("%v.%v.%v.%v", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
// 		port := fmt.Sprint(rand.Intn(9999-1000) + 1000)
// 		subcount := (rand.Intn(len(subSubjects)) + 1)
// 		broadcount := rand.Intn(len(broadSubjects) + 1)
// 		var subs []string
// 		var broads []string

// 		for i := 0; i < subcount; i++ {
// 			subs = append(subs, subSubjects[i])
// 		}

// 		for i := 0; i < broadcount; i++ {
// 			broads = append(broads, broadSubjects[i])
// 		}

// 		n := NewNode(uuid.NewString(), ip, port, subs, broads)
// 		nodes = append(nodes, n)
// 	}

// 	return nodes
// }

// var mainLis = bufconn.Listen(1024 * 1024)

// func TestMain(m *testing.M) {
// 	testMessageServer = NewMockServer(mainLis, "3000", false)
// 	wg := sync.WaitGroup{}
// 	ctx, cancel := context.WithCancel(context.Background())

// 	wg.Add(1)
// 	testMessageServer.start(ctx, &wg)

// 	exitVal := m.Run()

// 	cancel()
// 	os.Exit(exitVal)
// }
