package util

import (
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"github.com/gertd/go-pluralize"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

func searchInArray(array []string, value string) int {
	for idx, element := range array {
		if element == value {
			return idx
		}
	}
	return -1
}

func ParseK8sProxyURL(url string) (string, schema.GroupVersionKind, string) {
	urlParts := strings.Split(url, "/")
	arrLen := len(urlParts)
	if urlParts[arrLen-1] == bean.Empty {
		urlParts = urlParts[:arrLen-1]
		arrLen--
	}
	grammar := pluralize.NewClient()
	namespace := bean.ALL
	group := bean.ALL
	version := bean.ALL
	kind := bean.ALL
	resourceName := bean.ALL

	if arrLen < 2 {
		return namespace, schema.GroupVersionKind{Group: group, Version: version, Kind: kind}, resourceName
	}

	switch urlParts[1] {
	case bean.API:
		group = bean.K8sEmpty
		if arrLen > 2 {
			version = urlParts[2]
		}
		if arrLen > 5 {
			kind = grammar.Singular(urlParts[5])
			if arrLen > 6 {
				resourceName = urlParts[6]
			}
		}
	case bean.APIs:
		if arrLen > 2 {
			group = urlParts[2]
			if arrLen > 3 {
				version = urlParts[3]
			}
			if arrLen > 6 {
				kind = grammar.Singular(urlParts[6])
				if arrLen > 7 {
					resourceName = urlParts[7]
				}
			}
		}
	}

	if idx := searchInArray(urlParts, bean.NAMESPACES); idx != -1 && arrLen > idx+1 {
		namespace = urlParts[idx+1]
	} else if idx := searchInArray(urlParts, bean.NODES); idx != -1 {
		// what if command is of nodes
		// will check for super admin access
		return bean.ALL, schema.GroupVersionKind{Group: bean.ALL, Version: bean.ALL, Kind: bean.ALL}, bean.ALL
	}

	return namespace, schema.GroupVersionKind{Group: group, Version: version, Kind: kind}, resourceName
}
