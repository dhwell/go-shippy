package main

import (
	"context"
	"fmt"
	pb "github.com/dhwell/go-shippy/consignment-service/proto/consignment"
	vesselProto "github.com/dhwell/go-shippy/vessel-service/proto/vessel"
	"log"

	micro "github.com/micro/go-micro"
)

type IRepository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

// Repository - 模拟一个数据库，我们会在此后使用真正的数据库替代他
type Repository struct {
	consignments []*pb.Consignment
}

func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	return consignment, nil
}
func (repo *Repository) GetAll() []*pb.Consignment {
	return repo.consignments
}

// service要实现在proto中定义的所有方法。当你不确定时
// 可以去对应的*.pb.go文件里查看需要实现的方法及其定义
type service struct {
	repo         IRepository
	vesselClient vesselProto.VesselServiceClient
}

// CreateConsignment - 在proto中，我们只给这个微服务定一个了一个方法
// 就是这个CreateConsignment方法，它接受一个context以及proto中定义的
// Consignment消息，这个Consignment是由gRPC的服务器处理后提供给你的
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {
	// 这里，我们通过货船服务的客户端对象，向货船服务发出了一个请求
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})

	log.Printf("Found vessel: %s \n", vesselResponse.Vessel.Name)
	if err != nil {
		return err
	}

	req.VesselId = vesselResponse.Vessel.Id
	consignment, err := s.repo.Create(req)
	if err != nil {
		return err
	}
	res.Created = true
	res.Consignment = consignment
	return nil
}
func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	consignments := s.repo.GetAll()
	res.Consignments = consignments
	return nil
}
func main() {
	repo := &Repository{}
	// 注意，在这里我们使用go-micro的NewService方法来创建新的微服务服务器，
	// 而不是上一篇文章中所用的标准
	srv := micro.NewService(micro.Name("go.micro.srv.consignment"))
	// 我们在这里使用预置的方法生成了一个货船服务的客户端对象
	vesselClient := vesselProto.NewVesselServiceClient("go.micro.srv.vessel", srv.Client())

	// Init方法会解析命令行flags
	srv.Init()
	pb.RegisterShippingServiceHandler(srv.Server(), &service{repo, vesselClient})
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
