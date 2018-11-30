package model

type Node struct {
	Domain   string `json:"Domain"`
	IP       string `json:"IP"`
	Port     int    `json:"Port"`
	Enable   bool   `json:"Enable"`
	CertFile string `json:"CertFile"`
	KeyFile  string `json:"KeyFile"`
}
type GateWay struct {
	Port  int    `json:"Port"`
	Nodes []Node `json:"Nodes"`
}

func (gw GateWay) GetNodesMap() map[string]Node {
	dataMap := make(map[string]Node)
	for index := range gw.Nodes {
		//fmt.Println(&gw.Nodes[index])
		dataMap[gw.Nodes[index].Domain] = gw.Nodes[index]
		//afd := dataMap[gw.Nodes[index].Domain]
		//fmt.Println(&afd)
		//fmt.Println(dataMap[gw.Nodes[index].Domain])
	}

	return dataMap
}

type Config struct {
	GateWay GateWay `json:"GateWay"`
}
