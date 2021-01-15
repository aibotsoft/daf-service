package server

import (
	"context"
	"github.com/aibotsoft/daf-service/services/handler"
	pb "github.com/aibotsoft/gen/fortedpb"
	"github.com/aibotsoft/micro/config"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

type Server struct {
	cfg     *config.Config
	log     *zap.SugaredLogger
	gs      *grpc.Server
	handler *handler.Handler
	pb.UnimplementedFortedServer
}

func NewServer(cfg *config.Config, log *zap.SugaredLogger, handler *handler.Handler) *Server {
	return &Server{cfg: cfg, log: log, handler: handler, gs: grpc.NewServer()}
}
func (s *Server) Close() {
	s.log.Debug("begin gRPC server gracefulStop")
	s.gs.GracefulStop()
	s.handler.Close()
	s.log.Debug("end gRPC server gracefulStop")
}

func (*Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{}, nil
}
func (s *Server) ReleaseCheck(ctx context.Context, req *pb.ReleaseCheckRequest) (*pb.ReleaseCheckResponse, error) {
	s.handler.ReleaseCheck(ctx, req.GetSurebet())
	return &pb.ReleaseCheckResponse{}, nil
}
func (s *Server) Serve() error {
	//addr, err := s.handler.Conf.GetGrpcAddr(context.Background(), s.cfg.Service.Name)
	//if err != nil {
	//	s.log.Panic(err)
	//}
	hostPort := net.JoinHostPort("", s.cfg.Service.GrpcPort)

	lis, err := net.Listen("tcp", hostPort)
	if err != nil {
		return errors.Wrap(err, "net.Listen error")
	}
	pb.RegisterFortedServer(s.gs, s)
	s.log.Infow("gRPC server listens", "service", s.cfg.Service.Name, "hostPort", hostPort)
	return s.gs.Serve(lis)
}
func (s *Server) CheckLine(ctx context.Context, req *pb.CheckLineRequest) (*pb.CheckLineResponse, error) {
	sb := req.GetSurebet()
	err := s.handler.CheckLine(ctx, sb)
	if err != nil {
		s.log.Infow("handler.CheckLine error", "err", err)
	}
	return &pb.CheckLineResponse{Side: sb.Members[0]}, nil
}

func (s *Server) PlaceBet(ctx context.Context, req *pb.PlaceBetRequest) (*pb.PlaceBetResponse, error) {
	sb := req.GetSurebet()
	err := s.handler.PlaceBet(ctx, sb)
	if err != nil {
		s.log.Error(err)
	}
	return &pb.PlaceBetResponse{Side: sb.Members[0]}, nil
}
func (s *Server) GetResults(ctx context.Context, req *pb.GetResultsRequest) (*pb.GetResultsResponse, error) {
	results, err := s.handler.GetResults(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetResults from db error, service: %v", s.cfg.Service.Name)
	}
	return &pb.GetResultsResponse{Results: results}, nil
}
