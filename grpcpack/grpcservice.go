package grpcpack

import (
	pb "LightningOnOmni/grpcpack/pb"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type grpcService struct {
	client pb.BtcServiceClient
}

var instance *grpcService

func GetGrpcService() *grpcService {
	if instance == nil {
		instance = &grpcService{}
	}
	return instance
}

func (s *grpcService) SetClient(client pb.BtcServiceClient) {
	s.client = client
}

func (s *grpcService) GetNewAddress(c *gin.Context) {
	label := c.Param("label")
	// Contact the server and print out its response.
	req := &pb.AddressRequest{Label: label}
	res, err := s.client.GetNewAddress(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	jsonStr, _ := json.Marshal(res)
	c.JSON(http.StatusOK, gin.H{
		"result": string(jsonStr),
	})
}

func (s *grpcService) GetBlockCount(c *gin.Context) {
	// Contact the server and print out its response.
	res, err := s.client.GetBlockCount(c, &pb.EmptyRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	jsonStr, _ := json.Marshal(res)
	c.JSON(http.StatusOK, gin.H{
		"result": string(jsonStr),
	})
}
