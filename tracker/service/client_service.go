package service

import (
	"github.com/omnilaboratory/obd/tool"
	"github.com/omnilaboratory/obd/tracker/config"
	"strconv"
)
// get tracker node id
func GetTrackerNodeId() string {
	source := "tracker:" + tool.GetMacAddrs() + ":" + strconv.Itoa(cfg.TrackerServerPort)
	return tool.SignMsgWithSha256([]byte(source)) + cfg.ChainNode_Type
}
