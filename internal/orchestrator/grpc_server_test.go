package orchestrator

import (
	"context"
	"net"
	"strings"
	"testing"

	"calculator/internal/database"
	pb "calculator/internal/grpc/calculator"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func dialer() (*grpc.ClientConn, func(), error) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	store, err := database.NewStore(":memory:")
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	if err := store.InitDB(); err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	scheduler := NewScheduler(store)
	pb.RegisterCalculatorAgentServiceServer(srv, NewCalculatorGRPCServer(store, scheduler.GetOperationTimes(), scheduler))
	go srv.Serve(lis)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { conn.Close(); srv.Stop() }
	return conn, cleanup, nil
}

func TestGetTask_NoTask(t *testing.T) {
	conn, cleanup, err := dialer()
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Skip("skip gRPC tests: cgo disabled or in-memory DB not available")
	}
	defer cleanup()

	client := pb.NewCalculatorAgentServiceClient(conn)
	resp, err := client.GetTask(context.Background(), &pb.GetTaskRequest{AgentId: "test"})
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	if _, ok := resp.TaskInfo.(*pb.GetTaskResponse_NoTask); !ok {
		t.Errorf("expected NoTaskAvailable, got %T", resp.TaskInfo)
	}
}