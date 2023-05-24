package hdfs

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Property is the struct representation of hadoop configuration
// key value pair.
type Property struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type propertyList struct {
	Property []Property `xml:"property"`
}

// HadoopConf represents a map of all the key value configutation
// pairs found in a user's hadoop configuration files.
type HadoopConf map[string]string

var errNoNamenodesInConf = errors.New("no namenode address(es) in configuration")

// LoadHadoopConf returns a HadoopConf object representing configuration from
// the specified path, or finds the correct path in the environment. If
// path or the env variable HADOOP_CONF_DIR is specified, it should point
// directly to the directory where the xml files are. If neither is specified,
// ${HADOOP_HOME}/conf will be used.
func LoadHadoopConf(path string) HadoopConf {
	if path == "" {
		path = os.Getenv("HADOOP_CONF_DIR")
		if path == "" {
			path = filepath.Join(os.Getenv("HADOOP_HOME"), "conf")
		}
	}

	hadoopConf := make(HadoopConf)
	for _, file := range []string{"core-site.xml", "hdfs-site.xml"} {
		pList := propertyList{}
		f, err := ioutil.ReadFile(filepath.Join(path, file))
		if err != nil {
			continue
		}

		err = xml.Unmarshal(f, &pList)
		if err != nil {
			continue
		}

		for _, prop := range pList.Property {
			hadoopConf[prop.Name] = prop.Value
		}
	}

	return hadoopConf
}

// Namenodes returns the namenode hosts present in the configuration. The
// returned slice will be sorted and deduped. The values are loaded from
// fs.defaultFS (or the deprecated fs.default.name), or fields beginning with
// dfs.namenode.rpc-address.
func (conf HadoopConf) Namenodes() ([]string, error) {
	nns := make(map[string]bool)
	for key, value := range conf {
		if strings.Contains(key, "fs.default") {
			nnUrl, _ := url.Parse(value)
			nns[nnUrl.Host] = true
		} else if strings.HasPrefix(key, "dfs.namenode.rpc-address") {
			nns[value] = true
		}
	}

	if len(nns) == 0 {
		return nil, errNoNamenodesInConf
	}

	keys := make([]string, 0, len(nns))
	for k, _ := range nns {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys, nil
}
