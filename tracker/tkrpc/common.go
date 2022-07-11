package tkrpc

import (
	"errors"
	"strings"
)

func(list *ListReq) Offset() int{
	res:=int((list.Page-1)*list.Size)
	if res==0{
		res=10
	}
	return res
}
func(list *ListReq) Limit() int{
	res:=list.Size
	if res==0{
		res=1
	}
	return int(res)
}
func(list *ListReq) SortStr() string{
	orderstr:=""
	for _, item := range strings.Split(list.Sort,",") {
		if len(item)==1{
			continue
		}
		if strings.HasPrefix(item,"-"){
			orderstr+=item[1:]+" desc,"
		}else{
			orderstr+=item+","
		}
	}
	orderstr=strings.TrimSuffix(orderstr,",")
	return orderstr
}

func(info *HtlcInfo) ValidteClientData() (err error) {
	if info.Path==""{
		return errors.New("miss path")
	}
	if info.H==""{
		return errors.New("miss h")
	}
	if info.CurrChannelId==""{
		return errors.New("miss CurrChannelId")
	}
	return nil
}
func(info *ChannelInfo) ValidteClientData() (err error) {
	if info.ChannelId==""{
		return errors.New("miss ChannelId")
	}
	if info.NodeId==""{
		return errors.New("miss NodeId")
	}
	return nil
}