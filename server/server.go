package server

import (
	"bytes"
	"fmt"
	"gateway/model"
	"gateway/util"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

type GateWayHandle struct {
	NodesMap map[string]model.Node
}

func (gw GateWayHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.RequestURI)
	fmt.Println(r.Header.Get("Content-Type"))
	node := gw.NodesMap[r.Host]

	//fmt.Println(w)

	/*cookie, err := r.Cookie("GLSESSIONID")
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
	}*/

	if !strings.EqualFold(r.Header.Get("Sec-Websocket-Version"), "") || !strings.EqualFold(r.Header.Get("Sec-Websocket-Key"), "") {
		remote, err := url.Parse("ws://" + node.IP + ":" + strconv.Itoa(node.Port))
		util.CheckError(err)
		proxy := NewSingleHostReverseProxy(remote)
		proxy.ServeHTTP(w, r)
		//_currentNode.OperationTime = time.Now().Unix()
	} else {
		remote, err := url.Parse("http://" + node.IP + ":" + strconv.Itoa(node.Port))
		util.CheckError(err)
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.ModifyResponse = func(response *http.Response) error {
			fmt.Println(response.Request.Host)
			fmt.Println(response)

			buffer := bytes.NewBuffer(make([]byte, 0))
			n, err := io.Copy(buffer, response.Body)
			fmt.Println(n)
			fmt.Println(err)
			fmt.Println(string(buffer.Bytes()))
			return nil
		}
		proxy.ServeHTTP(w, r)
		//_currentNode.OperationTime = time.Now().Unix()
	}

}
