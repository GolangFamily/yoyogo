package Middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

type HttpContext struct {
	Req        *http.Request
	Resp       *responseWriter
	store      map[string]interface{}
	storeMutex *sync.RWMutex
}

func NewContext(w http.ResponseWriter, r *http.Request) *HttpContext {
	ctx := &HttpContext{}
	ctx.init(w, r)
	return ctx
}

func (ctx *HttpContext) init(w http.ResponseWriter, r *http.Request) {
	ctx.storeMutex = new(sync.RWMutex)
	ctx.Resp = &responseWriter{w, 0, 0, nil}
	ctx.Req = r
	ctx.storeMutex.Lock()
	ctx.store = nil
	ctx.storeMutex.Unlock()
}

//Set data in context.
func (ctx *HttpContext) SetItem(key string, val interface{}) {
	ctx.storeMutex.Lock()
	if ctx.store == nil {
		ctx.store = make(map[string]interface{})
	}
	ctx.store[key] = val
	ctx.storeMutex.Unlock()
}

// Get data in context.
func (ctx *HttpContext) GetItem(key string) interface{} {
	ctx.storeMutex.RLock()
	v := ctx.store[key]
	ctx.storeMutex.RUnlock()
	return v
}

//Set Cookie value
func (ctx *HttpContext) SetCookie(name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
	}
	ctx.Resp.Header().Add("Set-Cookie", cookie.String())
}

//Get Cookie by Name
func (ctx *HttpContext) GetCookie(name string) string {
	cookie, err := ctx.Req.Cookie(name)
	if err != nil {
		return ""
	}
	return url.QueryEscape(cookie.Value)
}

//Get Post Params
func (ctx *HttpContext) PostForm() url.Values {
	_ = ctx.Req.ParseForm()
	return ctx.Req.PostForm
}

func (ctx *HttpContext) PostMultipartForm() url.Values {
	_ = ctx.Req.ParseMultipartForm(32 << 20)
	return ctx.Req.MultipartForm.Value
}

func (ctx *HttpContext) PostJsonForm() url.Values {
	ret := url.Values{}
	var jsonMap map[string]interface{}
	body, _ := ioutil.ReadAll(ctx.Req.Body)
	_ = json.Unmarshal(body, &jsonMap)
	var strVal string
	for key, value := range jsonMap {
		switch value.(type) {
		case int32:
		case int64:
			strVal = strconv.Itoa(value.(int))
			break
		case float64:
			strVal = strconv.FormatFloat(value.(float64), 'f', -1, 64)
			break
		default:
			strVal = fmt.Sprint(value)
		}
		ret.Add(key, strVal)
	}
	return ret
}

//Get Post Param
func (ctx *HttpContext) Param(name string) string {
	var form url.Values

	content_type := strings.ToLower(ctx.Req.Header.Get("Content-Type"))
	if content_type == "application/x-www-form-urlencoded" {
		form = ctx.PostForm()
	} else if strings.Contains(content_type, "multipart/form-data") {
		form = ctx.PostMultipartForm()
	} else if strings.Contains(content_type, "application/json") {
		form = ctx.PostJsonForm()
	}

	if form[name] != nil {
		return form[name][0]
	}
	return ""
}

// Get Query string
func (ctx *HttpContext) QueryStrings() url.Values {

	queryForm, err := url.ParseQuery(ctx.Req.URL.RawQuery)
	if err == nil {
		return queryForm
	}
	return nil
}

// Get Query String By Key
func (ctx *HttpContext) Query(key string) string {
	return ctx.QueryStrings().Get(key)
}

// Redirect redirects the request
func (ctx *HttpContext) Redirect(code int, url string) {
	http.Redirect(ctx.Resp, ctx.Req, url, code)
}

// Path returns URL Path string.
func (ctx *HttpContext) Path() string {
	return ctx.Req.URL.Path
}

// Referer returns request referer.
func (ctx *HttpContext) Referer() string {
	return ctx.Req.Header.Get("Referer")
}

// UserAgent returns http request UserAgent
func (ctx *HttpContext) UserAgent() string {
	return ctx.Req.Header.Get("User-Agent")
}

//Get Http Method.
func (ctx *HttpContext) Method() string {
	return ctx.Req.Method
}

//Get Http Status Code.
func (ctx *HttpContext) Status() int {
	return ctx.Resp.status
}

// FormFile gets file from request.
func (ctx *HttpContext) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.Req.FormFile(key)
}

// SaveFile saves the form file and
// returns the filename.
func (ctx *HttpContext) SaveFile(name, saveDir string) (string, error) {
	fr, handle, err := ctx.FormFile(name)
	if err != nil {
		return "", err
	}
	defer fr.Close()
	fw, err := os.OpenFile(path.Join(saveDir, handle.Filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	defer fw.Close()

	_, err = io.Copy(fw, fr)
	return handle.Filename, err
}

// Write Error Response.
func (ctx *HttpContext) Error(code int, error string) {
	http.Error(ctx.Resp, error, code)
}

// Write Byte[] Response.
func (ctx *HttpContext) Write(data []byte) (n int, err error) {
	return ctx.Resp.Write(data)
}

// Text response text format data .
func (ctx *HttpContext) String(code int, format string, datas ...interface{}) error {
	text := fmt.Sprintf(format, datas)
	return ctx.Text(code, text)
}

// Text response text data.
func (ctx *HttpContext) Text(code int, body string) error {
	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.WriteHeader(code)
	_, err := ctx.Resp.Write([]byte(body))
	return err
}

// Write Json Response.
func (ctx *HttpContext) JSON(data interface{}) {
	ctx.Resp.Header().Set("Content-Type", "application/json")
	jsons, _ := json.Marshal(data)
	_, _ = ctx.Resp.Write(jsons)
}

// JSONP return JSONP data.
func (ctx *HttpContext) JSONP(code int, callback string, data interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	ctx.Resp.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	ctx.Resp.WriteHeader(code)
	_, _ = ctx.Resp.Write([]byte(callback + "("))
	_, _ = ctx.Resp.Write(j)
	_, _ = ctx.Resp.Write([]byte(");"))
	return nil
}
