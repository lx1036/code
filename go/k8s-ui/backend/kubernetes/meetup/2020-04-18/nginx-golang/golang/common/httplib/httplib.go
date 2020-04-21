package httplib

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"k8s-lx1036/k8s-ui/backend/kubernetes/meetup/2020-04-18/nginx-golang/golang/prometheus"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var defaultSetting = HTTPSettings{
	UserAgent:        "UtilHttpLibServer",
	ConnectTimeout:   1 * time.Second,
	ReadWriteTimeout: 3 * time.Second,
	Gzip:             true,
	DumpBody:         true,
}
var defaultCookieJar http.CookieJar
var settingMutex sync.Mutex

// createDefaultCookie creates a global cookiejar to store cookies.
func createDefaultCookie() {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultCookieJar, _ = cookiejar.New(nil)
}
func SetDefaultSetting(setting HTTPSettings) {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	defaultSetting = setting
}
func NewRequest(rawurl, method string) *HTTPRequest {
	var resp http.Response
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Println("HttpLib:", err)
	}
	req := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &HTTPRequest{
		url:     rawurl,
		req:     &req,
		params:  map[string][]string{},
		files:   map[string]string{},
		setting: defaultSetting,
		resp:    &resp,
	}
}
func Get(url string) *HTTPRequest {
	return NewRequest(url, "GET")
}
func Post(url string) *HTTPRequest {
	return NewRequest(url, "POST")
}
func Put(url string) *HTTPRequest {
	return NewRequest(url, "PUT")
}
func Delete(url string) *HTTPRequest {
	return NewRequest(url, "DELETE")
}
func Head(url string) *HTTPRequest {
	return NewRequest(url, "HEAD")
}

type HTTPSettings struct {
	ShowDebug        bool
	UserAgent        string
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	TLSClientConfig  *tls.Config
	Proxy            func(*http.Request) (*url.URL, error)
	Transport        http.RoundTripper
	CheckRedirect    func(req *http.Request, via []*http.Request) error
	EnableCookie     bool
	Gzip             bool
	DumpBody         bool
	Retries          int
}
type HTTPRequest struct {
	url     string
	req     *http.Request
	params  map[string][]string
	files   map[string]string
	setting HTTPSettings
	resp    *http.Response
	body    []byte
	dump    []byte
}

func (r *HTTPRequest) GetRequest() *http.Request {
	r.buildURL(r.getRequestParams())
	return r.req
}
func (r *HTTPRequest) Setting(setting HTTPSettings) *HTTPRequest {
	r.setting = setting
	return r
}
func (r *HTTPRequest) SetBasicAuth(username, password string) *HTTPRequest {
	r.req.SetBasicAuth(username, password)
	return r
}
func (r *HTTPRequest) SetEnableCookie(enable bool) *HTTPRequest {
	r.setting.EnableCookie = enable
	return r
}
func (r *HTTPRequest) SetUserAgent(useragent string) *HTTPRequest {
	r.setting.UserAgent = useragent
	return r
}
func (r *HTTPRequest) Debug(isdebug bool) *HTTPRequest {
	r.setting.ShowDebug = isdebug
	return r
}
func (r *HTTPRequest) Retries(times int) *HTTPRequest {
	r.setting.Retries = times
	return r
}
func (r *HTTPRequest) DumpBody(isdump bool) *HTTPRequest {
	r.setting.DumpBody = isdump
	return r
}
func (r *HTTPRequest) DumpRequest() []byte {
	return r.dump
}
func (r *HTTPRequest) SetTimeout(connectTimeout, readWriteTimout time.Duration) *HTTPRequest {
	r.setting.ConnectTimeout = connectTimeout
	r.setting.ReadWriteTimeout = readWriteTimout
	return r
}
func (r *HTTPRequest) SetTLSClientConfig(config *tls.Config) *HTTPRequest {
	r.setting.TLSClientConfig = config
	return r
}
func (r *HTTPRequest) Header(key, value string) *HTTPRequest {
	r.req.Header.Set(key, value)
	return r
}
func (r *HTTPRequest) SetHost(host string) *HTTPRequest {
	r.req.Host = host
	return r
}
func (r *HTTPRequest) SetProtocolVersion(vers string) *HTTPRequest {
	if len(vers) == 0 {
		vers = "HTTP/1.1"
	}
	major, minor, ok := http.ParseHTTPVersion(vers)
	if ok {
		r.req.Proto = vers
		r.req.ProtoMajor = major
		r.req.ProtoMinor = minor
	}
	return r
}
func (r *HTTPRequest) SetCookie(cookie *http.Cookie) *HTTPRequest {
	r.req.Header.Add("Cookie", cookie.String())
	return r
}
func (r *HTTPRequest) SetTransport(transport http.RoundTripper) *HTTPRequest {
	r.setting.Transport = transport
	return r
}
func (r *HTTPRequest) SetCheckRedirect(redirect func(req *http.Request, via []*http.Request) error) *HTTPRequest {
	r.setting.CheckRedirect = redirect
	return r
}
func (r *HTTPRequest) Param(key, value string) *HTTPRequest {
	if param, ok := r.params[key]; ok {
		r.params[key] = append(param, value)
	} else {
		r.params[key] = []string{value}
	}
	return r
}
func (r *HTTPRequest) PostFile(formname, filename string) *HTTPRequest {
	r.files[formname] = filename
	return r
}
func (r *HTTPRequest) Body(data interface{}) *HTTPRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(t))
	}
	return r
}
func (r *HTTPRequest) XmlBody(obj interface{}) (*HTTPRequest, error) {
	if r.req.Body == nil && obj != nil {
		byts, err := xml.Marshal(obj)
		if err != nil {
			return r, err
		}
		r.req.Body = ioutil.NopCloser(bytes.NewBuffer(byts))
		r.req.ContentLength = int64(len(byts))
		r.req.Header.Set("Content-Type", "application/xml")
	}
	return r, nil
}
func (r *HTTPRequest) JSONBody(obj interface{}) (*HTTPRequest, error) {
	if r.req.Body == nil && obj != nil {
		byts, err := json.Marshal(obj)
		if err != nil {
			return r, err
		}
		r.req.Body = ioutil.NopCloser(bytes.NewBuffer(byts))
		r.req.ContentLength = int64(len(byts))
		r.req.Header.Set("Content-Type", "application/json")
	}
	return r, nil
}
func (r *HTTPRequest) buildURL(paramBody string) {
	if r.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(r.url, "?") {
			r.url += "&" + paramBody
		} else {
			r.url = r.url + "?" + paramBody
		}
		return
	}
	if (r.req.Method == "POST" || r.req.Method == "PUT" || r.req.Method == "DELETE" || r.req.Method == "PATCH") && r.req.Body == nil {
		// with files
		if len(r.files) > 0 {
			pr, pw := io.Pipe()
			defer pw.Close()
			bodyWriter := multipart.NewWriter(pw)
			defer bodyWriter.Close()
			go func() {
				for formname, filename := range r.files {
					fileWriter, err := bodyWriter.CreateFormFile(formname, filename)
					if err != nil {
						log.Println("UtilHttpLib:", err)
					}
					fh, err := os.Open(filename)
					if err != nil {
						log.Println("UtilHttpLib:", err)
					}
					_, err = io.Copy(fileWriter, fh)
					fh.Close()
					if err != nil {
						log.Println("UtilHttpLib:", err)
					}
				}
				for k, v := range r.params {
					for _, vv := range v {
						bodyWriter.WriteField(k, vv)
					}
				}
			}()
			r.Header("Content-Type", bodyWriter.FormDataContentType())
			r.req.Body = ioutil.NopCloser(pr)
			return
		}
		// with params
		if len(paramBody) > 0 {
			r.Header("Content-Type", "application/x-www-form-urlencoded")
			r.Body(paramBody)
		}
	}
}
func (r *HTTPRequest) getResponse() (*http.Response, error) {
	if r.resp.StatusCode != 0 {
		return r.resp, nil
	}

	now := time.Now()
	response, err := r.DoRequest()
	latency := float64(time.Since(now).Milliseconds())

	go func() {
		if prometheus.GetWrapper() != nil {
			prometheus.GetWrapper().QpsCounterLog(prometheus.QpsRecord{
				Times:  1,
				Api:    r.req.URL.Path,
				Module: prometheus.Module,
				Method: r.req.Method,
				Code:   response.StatusCode,
			})

			prometheus.GetWrapper().LatencyLog(prometheus.LatencyRecord{
				Time:   latency,
				Api:    r.req.URL.Path,
				Module: prometheus.Module,
				Method: r.req.Method,
			})

			log.WithFields(log.Fields{
				"method":  r.req.Method,
				"path":    r.req.URL.Path,
				"status":  response.StatusCode,
				"latency": latency,
			}).Info("[api level]prometheus access logger")
		}
	}()

	if err != nil {
		return nil, err
	}

	r.resp = response
	return response, nil
}

func (r *HTTPRequest) getRequestParams() string {
	var paramBody string
	if len(r.params) > 0 {
		var buf bytes.Buffer
		for k, v := range r.params {
			for _, vv := range v {
				buf.WriteString(url.QueryEscape(k))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(vv))
				buf.WriteByte('&')
			}
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
	}

	return paramBody
}

func (r *HTTPRequest) DoRequest() (resp *http.Response, err error) {
	r.buildURL(r.getRequestParams())
	parseUrl, err := url.Parse(r.url)
	if err != nil {
		return nil, err
	}
	r.req.URL = parseUrl
	trans := r.setting.Transport
	if trans == nil {
		trans = &http.Transport{
			TLSClientConfig:     r.setting.TLSClientConfig,
			Proxy:               r.setting.Proxy,
			Dial:                TimeoutDialer(r.setting.ConnectTimeout, r.setting.ReadWriteTimeout),
			MaxIdleConnsPerHost: -1,
		}
	} else {
		if t, ok := trans.(*http.Transport); ok {
			if t.TLSClientConfig == nil {
				t.TLSClientConfig = r.setting.TLSClientConfig
			}
			if t.Proxy == nil {
				t.Proxy = r.setting.Proxy
			}
			if t.Dial == nil {
				t.Dial = TimeoutDialer(r.setting.ConnectTimeout, r.setting.ReadWriteTimeout)
			}
		}
	}
	var jar http.CookieJar
	if r.setting.EnableCookie {
		if defaultCookieJar == nil {
			createDefaultCookie()
		}
		jar = defaultCookieJar
	}
	client := &http.Client{
		Transport: trans,
		Jar:       jar,
	}
	if r.setting.UserAgent != "" && r.req.Header.Get("User-Agent") == "" {
		r.req.Header.Set("User-Agent", r.setting.UserAgent)
	}
	if r.setting.CheckRedirect != nil {
		client.CheckRedirect = r.setting.CheckRedirect
	}
	if r.setting.ShowDebug {
		dump, err := httputil.DumpRequest(r.req, r.setting.DumpBody)
		if err != nil {
			log.Println(err)
		}
		r.dump = dump
	}
	for i := 0; r.setting.Retries == -1 || i <= r.setting.Retries; i++ {
		resp, err = client.Do(r.req)
		if err == nil {
			break
		}
	}
	return resp, err
}
func (r *HTTPRequest) String() (string, error) {
	data, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}
func (r *HTTPRequest) Bytes() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}
	resp, err := r.getResponse()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	if r.setting.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		r.body, err = ioutil.ReadAll(reader)
		return r.body, err
	}
	r.body, err = ioutil.ReadAll(resp.Body)
	return r.body, err
}
func (r *HTTPRequest) ToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	resp, err := r.getResponse()
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
func (r *HTTPRequest) ToJSON(v interface{}) error {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
func (r *HTTPRequest) ToXml(v interface{}) error {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}
func (r *HTTPRequest) Response() (*http.Response, error) {
	return r.getResponse()
}

// TimeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}
