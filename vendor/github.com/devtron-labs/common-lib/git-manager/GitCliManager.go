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

package git_manager

import (
	"fmt"
	"github.com/devtron-labs/common-lib/git-manager/util"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitCliManager interface {
	Fetch(gitContext GitContext, rootDir string) (response, errMsg string, err error)
	Checkout(gitContext GitContext, rootDir string, checkout string) (response, errMsg string, err error)
	RunCommandWithCred(cmd *exec.Cmd, userName, password string, tlsPathInfo *TlsPathInfo) (response, errMsg string, err error)
	RunCommand(cmd *exec.Cmd) (response, errMsg string, err error)
	runCommandForSuppliedNullifiedEnv(cmd *exec.Cmd, setHomeEnvToNull bool) (response, errMsg string, err error)
	Init(rootDir string, remoteUrl string, isBare bool) error
	Clone(gitContext GitContext, prj CiProjectDetails) (response, errMsg string, err error)
	Merge(rootDir string, commit string) (response, errMsg string, err error)
	RecursiveFetchSubmodules(rootDir string) (response, errMsg string, error error)
	UpdateCredentialHelper(rootDir string) (response, errMsg string, error error)
	UnsetCredentialHelper(rootDir string) (response, errMsg string, error error)
	GitCheckout(gitContext GitContext, checkoutPath string, targetCheckout string, authMode AuthMode, fetchSubmodules bool, gitRepository string) (errMsg string, error error)
}

type GitCliManagerImpl struct {
}

func NewGitCliManager() *GitCliManagerImpl {
	return &GitCliManagerImpl{}
}

const GIT_AKS_PASS = "/git-ask-pass.sh"
const DefaultRemoteName = "origin"

func (impl *GitCliManagerImpl) Fetch(gitContext GitContext, rootDir string) (response, errMsg string, err error) {
	log.Println(util.DEVTRON, "git fetch ", "location", rootDir)
	tlsPathInfo, err := CreateFilesForTlsData(gitContext.TLSData, gitContext.WorkingDir)
	if err != nil {
		//making it non-blocking
		fmt.Println("error encountered in createFilesForTlsData", "err", err)
	}
	defer DeleteTlsFiles(tlsPathInfo)
	cmd := exec.Command("git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	output, errMsg, err := impl.RunCommandWithCred(cmd, gitContext.Auth.Username, gitContext.Auth.Password, tlsPathInfo)
	log.Println(util.DEVTRON, "fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, "", nil
}

func (impl *GitCliManagerImpl) Checkout(gitContext GitContext, rootDir string, checkout string) (response, errMsg string, err error) {
	log.Println(util.DEVTRON, "git checkout ", "location", rootDir)
	tlsPathInfo, err := CreateFilesForTlsData(gitContext.TLSData, gitContext.WorkingDir)
	if err != nil {
		//making it non-blocking
		fmt.Println("error encountered in createFilesForTlsData", "err", err)
	}
	defer DeleteTlsFiles(tlsPathInfo)
	cmd := exec.Command("git", "-C", rootDir, "checkout", checkout, "--force")
	output, errMsg, err := impl.RunCommandWithCred(cmd, gitContext.Auth.Username, gitContext.Auth.Password, tlsPathInfo)
	log.Println(util.DEVTRON, "checkout output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, "", nil
}

func (impl *GitCliManagerImpl) RunCommandWithCred(cmd *exec.Cmd, userName, password string, tlsPathInfo *TlsPathInfo) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_AKS_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", userName), // ignored
		fmt.Sprintf("GIT_PASSWORD=%s", password), // this value is used
	)
	if tlsPathInfo != nil {
		if tlsPathInfo.TlsKeyPath != "" && tlsPathInfo.TlsCertPath != "" {
			cmd.Env = append(cmd.Env,
				fmt.Sprintf("GIT_SSL_KEY=%s", tlsPathInfo.TlsKeyPath),
				fmt.Sprintf("GIT_SSL_CERT=%s", tlsPathInfo.TlsCertPath))
		}
		if tlsPathInfo.CaCertPath != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_SSL_CAINFO=%s", tlsPathInfo.CaCertPath))
		}
	}
	return impl.RunCommand(cmd)
}

func (impl *GitCliManagerImpl) RunCommand(cmd *exec.Cmd) (response, errMsg string, err error) {
	return impl.runCommandForSuppliedNullifiedEnv(cmd, true)
}

func (impl *GitCliManagerImpl) runCommandForSuppliedNullifiedEnv(cmd *exec.Cmd, setHomeEnvToNull bool) (response, errMsg string, err error) {
	if setHomeEnvToNull {
		cmd.Env = append(cmd.Env, "HOME=/dev/null")
	}
	// https://stackoverflow.com/questions/18159704/how-to-debug-exit-status-1-error-when-running-exec-command-in-golang
	// in CombinedOutput, both stdOut and stdError are returned in single output
	outBytes, err := cmd.CombinedOutput()
	output := string(outBytes)
	output = strings.Replace(output, "\n", "", -1)
	output = strings.TrimSpace(output)
	if err != nil {
		exErr, ok := err.(*exec.ExitError)
		if !ok {
			return "", output, err
		}
		errOutput := string(exErr.Stderr)
		return "", fmt.Sprintf("%s\n%s", output, errOutput), err
	}
	return output, "", nil
}

func (impl *GitCliManagerImpl) Init(rootDir string, remoteUrl string, isBare bool) error {

	//-----------------

	err := os.MkdirAll(rootDir, 0755)
	if err != nil {
		return err
	}
	err = impl.AddRepo(rootDir, remoteUrl)
	return err
}
func (impl *GitCliManagerImpl) AddRepo(rootDir string, remoteUrl string) error {
	err := impl.gitInit(rootDir)
	if err != nil {
		return err
	}
	return impl.gitCreateRemote(rootDir, remoteUrl)
}

func (impl *GitCliManagerImpl) gitInit(rootDir string) error {
	log.Println(util.DEVTRON, "git", "-C", rootDir, "init")
	cmd := exec.Command("git", "-C", rootDir, "init")
	output, errMsg, err := impl.RunCommand(cmd)
	log.Println(util.DEVTRON, "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return err
}

func (impl *GitCliManagerImpl) gitCreateRemote(rootDir string, url string) error {
	log.Println(util.DEVTRON, "git", "-C", rootDir, "remote", "add", DefaultRemoteName, url)
	cmd := exec.Command("git", "-C", rootDir, "remote", "add", DefaultRemoteName, url)
	output, errMsg, err := impl.RunCommand(cmd)
	log.Println(util.DEVTRON, "url", url, "opt", output, "errMsg", errMsg, "error", err)
	return err
}

// setting user.name and user.email as for non-fast-forward merge, git ask for user.name and email
func (impl *GitCliManagerImpl) Merge(rootDir string, commit string) (response, errMsg string, err error) {
	log.Println(util.DEVTRON, "git merge ", "location", rootDir)
	command := "cd " + rootDir + " && git config user.email git@devtron.com && git config user.name Devtron && git merge " + commit + " --no-commit"
	cmd := exec.Command("/bin/sh", "-c", command)
	output, errMsg, err := impl.RunCommand(cmd)
	log.Println(util.DEVTRON, "merge output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) RecursiveFetchSubmodules(rootDir string) (response, errMsg string, error error) {
	log.Println(util.DEVTRON, "git recursive fetch submodules ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "submodule", "update", "--init", "--recursive")
	output, eMsg, err := impl.runCommandForSuppliedNullifiedEnv(cmd, false)
	log.Println(util.DEVTRON, "recursive fetch submodules output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, eMsg, err
}

func (impl *GitCliManagerImpl) UpdateCredentialHelper(rootDir string) (response, errMsg string, error error) {
	log.Println(util.DEVTRON, "git credential helper store ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "config", "--global", "credential.helper", "store")
	output, eMsg, err := impl.runCommandForSuppliedNullifiedEnv(cmd, false)
	log.Println(util.DEVTRON, "git credential helper store output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, eMsg, err
}

func (impl *GitCliManagerImpl) UnsetCredentialHelper(rootDir string) (response, errMsg string, error error) {
	log.Println(util.DEVTRON, "git credential helper unset ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "config", "--global", "--unset", "credential.helper")
	output, eMsg, err := impl.runCommandForSuppliedNullifiedEnv(cmd, false)
	log.Println(util.DEVTRON, "git credential helper unset output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, eMsg, err
}

func (impl *GitCliManagerImpl) GitCheckout(gitContext GitContext, checkoutPath string, targetCheckout string, authMode AuthMode, fetchSubmodules bool, gitRepository string) (errMsg string, error error) {

	rootDir := filepath.Join(gitContext.WorkingDir, checkoutPath)

	// checkout target hash
	_, eMsg, cErr := impl.Checkout(gitContext, rootDir, targetCheckout)
	if cErr != nil {
		return eMsg, cErr
	}

	log.Println(util.DEVTRON, " fetchSubmodules ", fetchSubmodules, " authMode ", authMode)

	if fetchSubmodules {
		httpsAuth := (authMode == AUTH_MODE_USERNAME_PASSWORD) || (authMode == AUTH_MODE_ACCESS_TOKEN)
		if httpsAuth {
			// first remove protocol
			modifiedUrl := strings.ReplaceAll(gitRepository, "https://", "")
			// for bitbucket - if git repo url is started with username, then we need to remove username
			if strings.Contains(modifiedUrl, "bitbucket.org") && !strings.HasPrefix(modifiedUrl, "bitbucket.org") {
				modifiedUrl = modifiedUrl[strings.Index(modifiedUrl, "bitbucket.org"):]
			}
			// build url
			modifiedUrl = "https://" + gitContext.Auth.Username + ":" + gitContext.Auth.Password + "@" + modifiedUrl

			_, errMsg, cErr = impl.UpdateCredentialHelper(rootDir)
			if cErr != nil {
				return errMsg, cErr
			}

			cErr = util.CreateGitCredentialFileAndWriteData(modifiedUrl)
			if cErr != nil {
				return "Error in creating git credential file", cErr
			}

		}

		_, errMsg, cErr = impl.RecursiveFetchSubmodules(rootDir)
		if cErr != nil {
			return errMsg, cErr
		}

		// cleanup

		if httpsAuth {
			_, errMsg, cErr = impl.UnsetCredentialHelper(rootDir)
			if cErr != nil {
				return errMsg, cErr
			}

			// delete file (~/.git-credentials) (which was created above)
			cErr = util.CleanupAfterFetchingHttpsSubmodules()
			if cErr != nil {
				return "", cErr
			}
		}
	}

	return "", nil

}

func (impl *GitCliManagerImpl) shallowClone(gitContext GitContext, rootDir string, remoteUrl string, sourceBranch string) (response, errMsg string, err error) {
	log.Println(util.DEVTRON, "git shallow clone ", "location", rootDir)
	tlsPathInfo, err := CreateFilesForTlsData(gitContext.TLSData, gitContext.WorkingDir)
	if err != nil {
		//making it non-blocking
		fmt.Println("error encountered in createFilesForTlsData", "err", err)
	}
	defer DeleteTlsFiles(tlsPathInfo)
	cmd := exec.Command("git", "-C", rootDir, "clone", "--filter=tree:0", "--single-branch", "-b", sourceBranch, remoteUrl, "--no-checkout")
	output, errMsg, err := impl.RunCommandWithCred(cmd, gitContext.Auth.Username, gitContext.Auth.Password, tlsPathInfo)
	log.Println(util.DEVTRON, "shallow clone output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) moveFilesFromSourceToDestination(scrDir, dest string) (response, errMsg string, err error) {
	cmd := exec.Command("mv", scrDir+"/.git", dest+"/")
	output, errMsg, err := impl.RunCommand(cmd)
	log.Println(util.DEVTRON, "moving files from: ", scrDir+"/.git", " to: ", dest+"/", "opt: ", output, "errMsg: ", errMsg, "error: ", err)
	return output, errMsg, err
}

func (impl *GitCliManagerImpl) Clone(gitContext GitContext, prj CiProjectDetails) (response, errMsg string, err error) {
	var msgMsg string
	var checkoutPath string
	var cErr error
	checkoutBranch := prj.GetCheckoutBranchName()
	checkoutPath = filepath.Join(gitContext.WorkingDir, prj.CheckoutPath)
	err = os.MkdirAll(checkoutPath, 0755)
	if err != nil {
		return "", "", err
	}
	_, msgMsg, cErr = impl.shallowClone(gitContext, checkoutPath, prj.GitRepository, checkoutBranch)
	if cErr != nil {
		logrus.Error("could not clone repo ", "msgMsg: ", msgMsg, " err: ", cErr)
		return "", msgMsg, cErr
	}
	projectName := util.GetProjectName(prj.GitRepository)
	projRootDir := filepath.Join(checkoutPath, projectName)

	_, msgMsg, cErr = impl.moveFilesFromSourceToDestination(projRootDir, checkoutPath)
	if cErr != nil {
		logrus.Error("could not move files between files ", "msgMsg: ", msgMsg, "err: ", cErr)
		return "", msgMsg, cErr
	}
	return response, msgMsg, cErr
}
