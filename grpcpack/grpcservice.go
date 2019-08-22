package grpcpack

import (
	pb "LightningOnOmni/grpcpack/pb"
	"LightningOnOmni/rpc"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"strconv"
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

type test struct {
	Label string `json:"label"`
}

func (s *grpcService) GetNewAddress(c *gin.Context) {
	//json
	bytes, _ := c.GetRawData()
	log.Println(string(bytes))
	parse := gjson.Parse(string(bytes))
	log.Println(parse.Get("label").Get("test"))
	//log.Println(parse.Get("label").String())

	return
	//form
	label := c.PostForm("label")
	// Contact the server and print out its response.
	req := &pb.AddressRequest{Label: label}
	res, err := s.client.GetNewAddress(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"result": res,
	})
}

func (s *BtcRpcManager) GetNewAddress(ctx context.Context, in *pb.AddressRequest) (reply *pb.AddressReply, err error) {
	client := rpc.NewClient()
	result, err := client.GetNewAddress(in.GetLabel())
	if err != nil {
		log.Println(err)
	}
	return &pb.AddressReply{Address: result}, nil
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
	c.JSON(http.StatusOK, gin.H{
		"result": res,
	})
}

func (s *BtcRpcManager) GetBlockCount(ctx context.Context, in *pb.EmptyRequest) (reply *pb.BlockCountReply, err error) {
	client := rpc.NewClient()
	result, err := client.GetBlockCount()
	if err != nil {
		log.Println(err)
	}
	count, err := strconv.Atoi(result)
	return &pb.BlockCountReply{Count: int32(count)}, nil
}

func (s *grpcService) GetMiningInfo(c *gin.Context) {
	res, err := s.client.GetMiningInfo(c, &pb.EmptyRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	//parse := gjson.Parse(res.Data)
	//var node = make(map[string]interface{})
	//node["blocks"] = parse.Get("blocks").Num
	//node["currentblocksize"] = parse.Get("currentblocksize").Num
	//node["currentblockweight"] = parse.Get("currentblockweight").Num
	//node["currentblocktx"] = parse.Get("currentblocktx").Num
	//node["difficulty"] = parse.Get("difficulty").Float()
	//node["networkhashps"] = parse.Get("networkhashps").Float()
	//node["pooledtx"] = parse.Get("pooledtx").Int()
	//node["testnet"] = parse.Get("testnet").Bool()
	//node["chain"] = parse.Get("chain").String()
	c.JSON(http.StatusOK, gin.H{
		"result": res,
	})
}

func (s *BtcRpcManager) GetMiningInfo(ctx context.Context, in *pb.EmptyRequest) (reply *pb.MiningInfoReply, err error) {
	client := rpc.NewClient()
	result, err := client.GetMiningInfo()
	if err != nil {
		log.Println(err)
	}
	reply = &pb.MiningInfoReply{}
	err = json.Unmarshal([]byte(result), reply)
	return reply, err
}
