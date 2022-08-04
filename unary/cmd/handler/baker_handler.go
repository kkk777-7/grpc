package handler

import (
	"context"
	"math/rand"
	"pancake/maker/pkg/test"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type myHandler struct {
	test.UnimplementedPancakeBakerServiceServer
	report *report
}

type report struct {
	sync.Mutex
	data map[test.Pancake_Menu]int
}

func NewBakerHandler() *myHandler {
	return &myHandler{
		report: &report{
			data: make(map[test.Pancake_Menu]int),
		},
	}
}

func (h *myHandler) Bake(ctx context.Context, req *test.BakeRequest) (*test.BakeResponse, error) {
	if req.Menu == test.Pancake_UNKNOWN || req.Menu > test.Pancake_SPICY_CURRY {
		return nil, status.Errorf(codes.InvalidArgument, "パンケーキを選んでください!")
	}

	now := time.Now()
	h.report.Lock()
	h.report.data[req.Menu] += 1
	h.report.Unlock()

	return &test.BakeResponse{
		Pancake: &test.Pancake{
			Menu:           req.Menu,
			ChefName:       "ririko",
			TechnicalScore: rand.Float32(),
			CreateTime: &timestamp.Timestamp{
				Seconds: now.Unix(),
				Nanos:   int32(now.Nanosecond()),
			},
		},
	}, nil
}

func (h *myHandler) Report(ctx context.Context, req *test.ReportRequest) (*test.ReportResponse, error) {
	counts := make([]*test.Report_BakeCount, 0)

	h.report.Lock()
	for k, v := range h.report.data {
		counts = append(counts, &test.Report_BakeCount{
			Menu:  k,
			Count: int32(v),
		})
	}
	h.report.Unlock()

	return &test.ReportResponse{
		Report: &test.Report{
			BakeCounts: counts,
		},
	}, nil
}
