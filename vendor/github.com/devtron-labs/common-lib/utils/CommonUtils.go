/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"
	"github.com/devtron-labs/common-lib/git-manager/util"
	"github.com/devtron-labs/common-lib/utils/bean"
	"log"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

const (
	DOCKER_REGISTRY_TYPE_DOCKERHUB = "docker-hub"
	DEVTRON_SELF_POD_UID           = "DEVTRON_SELF_POD_UID"
	DEVTRON_SELF_POD_NAME          = "DEVTRON_SELF_POD_NAME"
)

// Generates random string
func Generate(size int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}

func hasScheme(url string) bool {
	return len(url) >= 7 && (url[:7] == "http://" || url[:8] == "https://")
}

func GetUrlWithScheme(url string) (urlWithScheme string) {
	urlWithScheme = url
	if !hasScheme(url) {
		urlWithScheme = fmt.Sprintf("https://%s", url)
	}
	return urlWithScheme
}

func IsValidDockerTagName(tagName string) bool {
	regString := "^[a-zA-Z0-9_][a-zA-Z0-9_.-]{0,127}$"
	regexpCompile := regexp.MustCompile(regString)
	if regexpCompile.MatchString(tagName) {
		return true
	} else {
		return false
	}
}

func BuildDockerImagePath(dockerInfo bean.DockerRegistryInfo) (string, error) {
	dest := ""
	if DOCKER_REGISTRY_TYPE_DOCKERHUB == dockerInfo.DockerRegistryType {
		dest = dockerInfo.DockerRepository + ":" + dockerInfo.DockerImageTag
	} else {
		registryUrl := dockerInfo.DockerRegistryURL
		u, err := util.ParseUrl(registryUrl)
		if err != nil {
			log.Println("not a valid docker repository url")
			return "", err
		}
		u.Path = path.Join(u.Path, "/", dockerInfo.DockerRepository)
		dockerRegistryURL := u.Host + u.Path
		dest = dockerRegistryURL + ":" + dockerInfo.DockerImageTag
	}
	return dest, nil
}

func GetSelfK8sUID() string {
	return os.Getenv(DEVTRON_SELF_POD_UID)
}

func GetSelfK8sPodName() string {
	return os.Getenv(DEVTRON_SELF_POD_NAME)
}
