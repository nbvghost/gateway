package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gateway/util"
	"net"

	"gateway/model"
	"gateway/server"
	"io/ioutil"
	"log"

	"net/http"

	"strconv"

	"sync"
	"time"
)

var pool = &NodePool{Pool: make(map[string]model.Node)}
var Config model.Config

type NodePool struct {
	sync.RWMutex
	Pool map[string]model.Node
}

/*type UrlInfo struct {
	Url string
}*/

//var cacheData = make(map[string]map[string]UrlInfo)

//var infoPage []byte

func init() {
	var ConfigFile string
	flag.StringVar(&ConfigFile, "config", "config.json", "config")
	flag.Parse()

	content, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		panic("缺少配制文件：config.json")

	} else {
		//fmt.Println(string(content))
		err = json.Unmarshal(content, &Config)
		util.CheckError(err)
	}

	//infoPage, _ = ioutil.ReadFile("info.html")
}

func main() {
	http.DefaultClient.Timeout = time.Second * 1

	//go runInfo()
	//readAbleServer()
	//go startHttp()
	domainMap := Config.GateWay.GetNodesMap()

	h := &server.GateWayHandle{}
	h.NodesMap = domainMap

	go func() {

		tlsCfg := &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				node, ok := domainMap[info.ServerName]
				if ok {

					cert, err := tls.X509KeyPair([]byte(node.CertFile), []byte(node.KeyFile))
					if err != nil {
						return nil, err
					}

					return &cert, nil
				}

				return nil, errors.New("not exist ca")

			},
		}

		server := &http.Server{Addr: ":443", Handler: h, TLSConfig: tlsCfg}

		server.ConnState = func(conn net.Conn, state http.ConnState) {
			fmt.Println(state)
		}

		err := server.ListenAndServeTLS("", "")
		util.CheckError(err)

	}()

	util.Trace("gateway start to " + strconv.Itoa(Config.GateWay.Port))
	err := http.ListenAndServe(":"+strconv.Itoa(Config.GateWay.Port), h)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}

/*func startHttp() {
	//http.Handle("/status", FrameworkHttp.HttpObject{HandlerFunc: statusAction})
	http.HandleFunc("/status", statusAction)
	err := http.ListenAndServe(":9191", nil)
	log.Println(err)
}*/

/*func statusAction(Response http.ResponseWriter, Request *http.Request) {

	bs := ""
	for _, v := range Config.GateWay.Nodes {
		b, err := xml.Marshal(v)
		CheckError(err)
		bs = bs + string(b)
		//fmt.Println(bs)
	}

	bs = `<?xml version="1.0" encoding="UTF-8"?><Nodes>` + bs + `</Nodes>`

	Response.Write([]byte(bs))
}*/

/*func readAbleServer() {
	go func() {
		for {

			pool.Lock()
			for k, v := range pool.Pool {
				if v == nil {
					continue
				}

				if v.OperationTime+int64(30*60) < time.Now().Unix() {

					delete(pool.Pool, k)

				}
			}
			pool.Unlock()
			//fmt.Println(currentNode)
			time.Sleep(time.Second * 1)
		}
	}()
}*/

/*func runInfo() {

	for {
		for i := 0; i < len(Config.GateWay.Nodes); i++ {
			//go connectSocket(&gateway.Nodes[i])
			requestRuninfo(&Config.GateWay.Nodes[i])

		}
		time.Sleep(time.Second)
	}

}*/

/*func requestRuninfo(node *Node) {

	IPAddress := node.IP + ":" + strconv.Itoa(node.Port)

	request, err := http.NewRequest("GET", "http://"+IPAddress+"/info/run", nil)
	request.AddCookie(&http.Cookie{Name: "GLSESSIONID", Value: IPAddress, Path: "/"})
	resp, err := http.DefaultClient.Do(request)
	//resp, err := http.Get("http://" + node.IP + ":" + strconv.Itoa(node.Port) + "/info/run")
	if err == nil {
		body, _ := ioutil.ReadAll(resp.Body)
		d := make(map[string]interface{})
		json.Unmarshal(body, &d)

		if d["Data"] != nil {
			Data := d["Data"].(map[string]interface{})
			if Data["OnlineCount"] != nil {
				OnlineCount := Data["OnlineCount"].(float64)

				pool.Lock()
				node.ClientCount, _ = strconv.Atoi(strconv.FormatFloat(OnlineCount, 'f', 0, 64))
				node.Enable = true
				pool.Unlock()
			}
		}

	} else {
		node.Enable = false
		//CheckError(err)
	}

}*/
