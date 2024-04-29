package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"math/rand"
	"net/http"
	u "net/url"
	"strings"
	"text/template"
	"time"
)

const (
	CONTENT_TYPE     = "Content-Type"
	APPLICATION_JSON = "application/json"
)

// Tprintf passed template string is formatted usign its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
func Tprintf(tmpl string, data interface{}) (string, error) {
	t := template.Must(template.New("example").Funcs(template.FuncMap{
		"add": func(a, b int64) int64 { return a + b },
	}).Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func DoHttpPOSTRequest(url string, queryParams, headers map[string]string, payload interface{}) (bool, error) {
	client := http.Client{}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		klog.Errorln("error while marshaling event request ", "err", err)
		return false, err
	}

	if len(queryParams) > 0 {
		params := u.Values{}
		for key, val := range queryParams {
			params.Set(key, val)
		}
		url = fmt.Sprintf("%s?%s", url, params.Encode())
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		klog.Errorln("error while writing event", "err", err)
		return false, err
	}

	if len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		klog.Errorln("error in sending notification rest request ", "err", err)
		return false, err
	}
	klog.Infof("notification response %s", resp.Status)
	defer resp.Body.Close()
	return true, err
}

func ObjectMapAdapter(obj interface{}) map[string]interface{} {
	mp, _ := obj.(map[string]interface{})
	return mp
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Generate random string
func Generate(size int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}

func GetConfigOrDie(kubeconfig string) *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}
	return config
}

type Response struct {
	Code   int         `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Errors interface{} `json:"errors,omitempty"`
}
