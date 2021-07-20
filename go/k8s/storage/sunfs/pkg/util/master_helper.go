package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

const (
	requestTimeout = 30 * time.Second
)

var (
	ErrNoValidMaster = errors.New("no valid master")
)

// MasterHelper defines the helper struct to manage the master.
type MasterHelper interface {
	AddNode(address string)
	Nodes() []string
	Leader() string
	Request(method, path string, param map[string]string, body []byte) (data []byte, err error)
}

type masterHelper struct {
	sync.RWMutex
	masters    []string
	leaderAddr string
}

func (helper *masterHelper) AddNode(address string) {
	helper.Lock()
	defer helper.Unlock()

	helper.updateMaster(address)
}

func (helper *masterHelper) Nodes() []string {
	panic("implement me")
}

func (helper *masterHelper) Leader() string {
	panic("implement me")
}

// prepareRequest returns the leader address and all master addresses.
func (helper *masterHelper) prepareRequest() (addr string, nodes []string) {
	helper.RLock()
	defer helper.RUnlock()
	addr = helper.leaderAddr
	nodes = helper.masters
	return
}

func (helper *masterHelper) Request(method, path string, param map[string]string, reqData []byte) (repsData []byte, err error) {
	leaderAddr, nodes := helper.prepareRequest()
	host := leaderAddr
	for i := -1; i < len(nodes); i++ {
		if i == -1 {
			if host == "" {
				continue
			}
		} else {
			host = nodes[i]
		}
		var resp *http.Response
		resp, err = helper.httpRequest(method, fmt.Sprintf("http://%s%s", host,
			path), param, reqData)
		if err != nil {
			klog.Errorf("[masterHelper] %s", err)
			continue
		}
		stateCode := resp.StatusCode
		repsData, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			klog.Errorf("[masterHelper] %s", err)
			continue
		}
		switch stateCode {
		case http.StatusForbidden:
			curMasterAddr := strings.TrimSpace(string(repsData))
			curMasterAddr = strings.Replace(curMasterAddr, "\n", "", -1)
			if len(curMasterAddr) == 0 {
				klog.Errorf("[masterHelper] request[%s] response statudCode"+
					"[403], respBody is empty", host)
				err = ErrNoValidMaster
				return
			}
			repsData, err = helper.request(method, path, param, reqData)
			return
		case http.StatusOK:
			if leaderAddr != host {
				helper.setLeader(host)
			}
			var body = &struct {
				Code int32  `json:"code"`
				Msg  string `json:"msg"`
				Data json.RawMessage
			}{}
			if err := json.Unmarshal(repsData, body); err != nil {
				return nil, fmt.Errorf("unmarshal response body err:%v", err)

			}
			// o represent proto.ErrCodeSuccess
			if body.Code != 0 {
				return nil, fmt.Errorf("request error, code[%d], msg[%s]", body.Code, body.Msg)
			}
			return []byte(body.Data), nil
		default:
			klog.Errorf("[masterHelper] master[%v] uri[%v] statusCode[%v] respBody[%v].",
				resp.Request.URL.String(), host, stateCode, string(repsData))
			continue
		}
	}
	err = ErrNoValidMaster
	return
}

func (helper *masterHelper) request(method, path string, param map[string]string, reqData []byte) (repsData []byte, err error) {
	leaderAddr, nodes := helper.prepareRequest()
	host := leaderAddr
	for i := -1; i < len(nodes); i++ {
		if i == -1 {
			if host == "" {
				continue
			}
		} else {
			host = nodes[i]
		}
		var resp *http.Response
		resp, err = helper.httpRequest(method, fmt.Sprintf("http://%s%s", host,
			path), param, reqData)
		if err != nil {
			klog.Errorf("[masterHelper] %s", err)
			continue
		}
		stateCode := resp.StatusCode
		repsData, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			klog.Errorf("[masterHelper] %s", err)
			continue
		}
		switch stateCode {
		case http.StatusForbidden:
			curMasterAddr := strings.TrimSpace(string(repsData))
			curMasterAddr = strings.Replace(curMasterAddr, "\n", "", -1)
			if len(curMasterAddr) == 0 {
				klog.Errorf("[masterHelper] request[%s] response statudCode"+
					"[403], respBody is empty", host)
				err = ErrNoValidMaster
				return
			}
			repsData, err = helper.request(method, path, param, reqData)
			return
		case http.StatusOK:
			if leaderAddr != host {
				helper.setLeader(host)
			}
			var body = &struct {
				Code int32  `json:"code"`
				Msg  string `json:"msg"`
				Data json.RawMessage
			}{}
			if err := json.Unmarshal(repsData, body); err != nil {
				return nil, fmt.Errorf("unmarshal response body err:%v", err)

			}
			// o represent proto.ErrCodeSuccess
			if body.Code != 0 {
				return nil, fmt.Errorf("request error, code[%d], msg[%s]", body.Code, body.Msg)
			}
			return []byte(body.Data), nil
		default:
			klog.Errorf("[masterHelper] master[%v] uri[%v] statusCode[%v] respBody[%v].",
				resp.Request.URL.String(), host, stateCode, string(repsData))
			continue
		}
	}
	err = ErrNoValidMaster
	return
}

func (helper *masterHelper) httpRequest(method, url string, param map[string]string, reqData []byte) (resp *http.Response, err error) {
	client := &http.Client{}
	reader := bytes.NewReader(reqData)
	client.Timeout = requestTimeout
	var req *http.Request
	fullUrl := helper.mergeRequestUrl(url, param)
	klog.V(5).Infof("action[httpRequest] method[%v] url[%v] reqBodyLen[%v].", method, fullUrl, len(reqData))
	if req, err = http.NewRequest(method, fullUrl, reader); err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	resp, err = client.Do(req)
	return
}

func (helper *masterHelper) mergeRequestUrl(url string, params map[string]string) string {
	if params != nil && len(params) > 0 {
		buff := bytes.NewBuffer([]byte(url))
		isFirstParam := true
		for k, v := range params {
			if isFirstParam {
				buff.WriteString("?")
				isFirstParam = false
			} else {
				buff.WriteString("&")
			}
			buff.WriteString(k)
			buff.WriteString("=")
			buff.WriteString(v)
		}
		return buff.String()
	}
	return url
}

func (helper *masterHelper) updateMaster(address string) {
	contains := false
	for _, master := range helper.masters {
		if master == address {
			contains = true
			break
		}
	}
	if !contains {
		helper.masters = append(helper.masters, address)
	}

	// INFO: 这个逻辑岂不是最后一个 address 是 leader
	helper.leaderAddr = address
}

// Change the leader address.
func (helper *masterHelper) setLeader(addr string) {
	helper.Lock()
	defer helper.Unlock()
	helper.leaderAddr = addr
}

// NewMasterHelper returns a new MasterHelper instance.
func NewMasterHelper() MasterHelper {
	return &masterHelper{}
}
