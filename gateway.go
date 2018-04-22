package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var pool = &NodePool{Pool: make(map[string]*Node)}
var Config Configer

type Node struct {
	IP             string `xml:"IP"`
	Port           int    `xml:"Port"`
	ClientCount    int    `xml:"ClientCount"`
	MaxClientCount int    `xml:"MaxClientCount"`
	Enable         bool   `xml:"Enable"`
	OperationTime  int64
}
type GateWay struct {
	Port  string `xml:"Port"`
	Nodes []Node `xml:"Node"`
}
type Configer struct {
	GateWay GateWay `xml:"GateWay"`
}

type GateWayHandle struct {
}

type NodePool struct {
	sync.RWMutex
	Pool map[string]*Node
}

func (n *NodePool) getMaxAble() *Node {

	minValueIndex := -1
	minValue := math.MaxFloat64

	for i := 0; i < len(Config.GateWay.Nodes); i++ {
		mitem := Config.GateWay.Nodes[i]

		if mitem.Enable == true {
			cminValue := math.Min(minValue, float64(mitem.ClientCount)/float64(mitem.MaxClientCount))
			if cminValue < minValue {
				minValueIndex = i
				minValue = cminValue
			}
		}
	}
	if minValueIndex == -1 {
		return &Config.GateWay.Nodes[0] //如果没有可用的服务器时，默认使用第一个
	} else {
		return &Config.GateWay.Nodes[minValueIndex]
	}

}
func (n *NodePool) getNode(glsessionid string) *Node {
	n.RLock()
	defer n.RUnlock()
	return n.Pool[glsessionid]
}
func (n *NodePool) setNode(glsessionid string, node *Node) {
	n.Lock()
	defer n.Unlock()
	n.Pool[glsessionid] = node
}

func (this *GateWayHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("GLSESSIONID")
	var _currentNode *Node
	if err != nil {
		_currentNode = pool.getMaxAble()
		glsessionid := UUID().String() + "." + strings.Replace(_currentNode.IP, ".", "", -1) + ":" + strconv.Itoa(_currentNode.Port)
		fmt.Println(glsessionid)
		pool.setNode(glsessionid, _currentNode)
		http.SetCookie(w, &http.Cookie{Name: "GLSESSIONID", Value: glsessionid, Path: "/"})
		r.AddCookie(&http.Cookie{Name: "GLSESSIONID", Value: glsessionid, Path: "/"})
	} else {
		glsessionid := cookie.Value
		_currentNode = pool.getNode(glsessionid)

		if _currentNode == nil || _currentNode.Enable == false {
			_currentNode = pool.getMaxAble()
			pool.setNode(glsessionid, _currentNode)
		}
	}

	if _currentNode == nil || _currentNode.Enable == false {

		w.Header().Add("Content-Type", "text/html")
		w.Write(infoPage)
		return
	}

	if !strings.EqualFold(r.Header.Get("Sec-Websocket-Version"), "") || !strings.EqualFold(r.Header.Get("Sec-Websocket-Key"), "") {
		remote, err := url.Parse("ws://" + _currentNode.IP + ":" + strconv.Itoa(_currentNode.Port))
		CheckError(err)
		proxy := NewSingleHostReverseProxy(remote)
		proxy.ServeHTTP(w, r)
		_currentNode.OperationTime = time.Now().Unix()
	} else {
		remote, err := url.Parse("http://" + _currentNode.IP + ":" + strconv.Itoa(_currentNode.Port))
		CheckError(err)
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.ServeHTTP(w, r)
		_currentNode.OperationTime = time.Now().Unix()
	}

}

var infoPage []byte

func init() {
	var ConfigFile string
	flag.StringVar(&ConfigFile, "config", "Configer.xml", "config")
	flag.Parse()

	content, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		panic("缺少配制文件：Configer.xml")

	} else {
		err = xml.Unmarshal(content, &Config)
		CheckError(err)
	}
	infoPage, _ = ioutil.ReadFile("info.html")
}
func main() {
	http.DefaultClient.Timeout = time.Second * 1

	go runInfo()
	readAbleServer()
	go startHttp()

	h := &GateWayHandle{}

	go func() {

		err := http.ListenAndServeTLS(":443", "cas/cert.pem", "cas/key.key", h)
		CheckError(err)

	}()

	Trace("gateway start to " + Config.GateWay.Port)
	err := http.ListenAndServe(":"+Config.GateWay.Port, h)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}
func startHttp() {
	//http.Handle("/status", FrameworkHttp.HttpObject{HandlerFunc: statusAction})
	http.HandleFunc("/status", statusAction)
	err := http.ListenAndServe(":9191", nil)
	log.Println(err)
}
func statusAction(Response http.ResponseWriter, Request *http.Request) {

	bs := ""
	for _, v := range Config.GateWay.Nodes {
		b, err := xml.Marshal(v)
		CheckError(err)
		bs = bs + string(b)
		//fmt.Println(bs)
	}

	bs = `<?xml version="1.0" encoding="UTF-8"?><Nodes>` + bs + `</Nodes>`

	Response.Write([]byte(bs))
}

func readAbleServer() {
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
}

func runInfo() {

	for {
		for i := 0; i < len(Config.GateWay.Nodes); i++ {
			//go connectSocket(&gateway.Nodes[i])
			requestRuninfo(&Config.GateWay.Nodes[i])

		}
		time.Sleep(time.Second)
	}

}

func requestRuninfo(node *Node) {

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

}
