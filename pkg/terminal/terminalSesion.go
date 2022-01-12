// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package terminal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"go.uber.org/zap"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	"log"
	"net/http"
	"sync"

	"gopkg.in/igm/sockjs-go.v3/sockjs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const END_OF_TRANSMISSION = "\u0004"

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

// TerminalSession implements PtyHandler (using a SockJS connection)
type TerminalSession struct {
	id            string
	bound         chan error
	sockJSSession sockjs.Session
	sizeChan      chan remotecommand.TerminalSize
	doneChan      chan struct{}
}

// TerminalMessage is the messaging protocol between ShellController and TerminalSession.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// bind    fe->be     SessionID      Id sent back from TerminalResponse
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type TerminalMessage struct {
	Op, Data, SessionID string
	Rows, Cols          uint16
}

// TerminalSize handles pty->process resize events
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as the process is running
func (t TerminalSession) Read(p []byte) (int, error) {
	m, err := t.sockJSSession.Recv()
	if err != nil {
		// Send terminated signal to process to avoid resource leak
		return copy(p, END_OF_TRANSMISSION), err
	}

	var msg TerminalMessage
	if err := json.Unmarshal([]byte(m), &msg); err != nil {
		return copy(p, END_OF_TRANSMISSION), err
	}

	switch msg.Op {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		t.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t TerminalSession) Write(p []byte) (int, error) {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}

	if err = t.sockJSSession.Send(string(msg)); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t TerminalSession) Toast(p string) error {
	msg, err := json.Marshal(TerminalMessage{
		Op:   "toast",
		Data: p,
	})
	if err != nil {
		return err
	}

	if err = t.sockJSSession.Send(string(msg)); err != nil {
		return err
	}
	return nil
}

// SessionMap stores a map of all TerminalSession objects and a lock to avoid concurrent conflict
type SessionMap struct {
	Sessions map[string]TerminalSession
	Lock     sync.RWMutex
}

// Get return a given terminalSession by sessionId
func (sm *SessionMap) Get(sessionId string) TerminalSession {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	return sm.Sessions[sessionId]
}

// Set store a TerminalSession to SessionMap
func (sm *SessionMap) Set(sessionId string, session TerminalSession) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	sm.Sessions[sessionId] = session
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *SessionMap) Close(sessionId string, status uint32, reason string) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	err := sm.Sessions[sessionId].sockJSSession.Close(status, reason)
	if err != nil {
		log.Println(err)
	}

	delete(sm.Sessions, sessionId)
}

var terminalSessions = SessionMap{Sessions: make(map[string]TerminalSession)}

// handleTerminalSession is Called by net/http for any new /api/sockjs connections
func handleTerminalSession(session sockjs.Session) {
	var (
		buf             string
		err             error
		msg             TerminalMessage
		terminalSession TerminalSession
	)

	if buf, err = session.Recv(); err != nil {
		log.Printf("handleTerminalSession: can't Recv: %v", err)
		return
	}

	if err = json.Unmarshal([]byte(buf), &msg); err != nil {
		log.Printf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
		return
	}

	if msg.Op != "bind" {
		log.Printf("handleTerminalSession: expected 'bind' message, got: %s", buf)
		session.Close(http.StatusBadRequest, fmt.Sprintf("expected 'bind' message, got '%s'", buf))
		return
	}

	if terminalSession = terminalSessions.Get(msg.SessionID); terminalSession.id == "" {
		log.Printf("handleTerminalSession: can't find session '%s'", msg.SessionID)
		session.Close(http.StatusGone, fmt.Sprintf("handleTerminalSession: can't find session '%s'", msg.SessionID))
		return
	}

	terminalSession.sockJSSession = session
	terminalSessions.Set(msg.SessionID, terminalSession)
	terminalSession.bound <- nil
}

// CreateAttachHandler is called from main for /api/sockjs
func CreateAttachHandler(path string) http.Handler {
	return sockjs.NewHandler(path, sockjs.DefaultOptions, handleTerminalSession)
}

// startProcess is called by handleAttach
// Executed cmd in the container specified in request and connects it up with the ptyHandler (a session)
func startProcess(k8sClient kubernetes.Interface, cfg *rest.Config,
	cmd []string, ptyHandler PtyHandler, sessionRequest *TerminalSessionRequest) error {
	namespace := sessionRequest.Namespace
	podName := sessionRequest.PodName
	containerName := sessionRequest.ContainerName

	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil
}

// genTerminalSessionId generates a random session ID string. The format is not really interesting.
// This ID is used to identify the session when the client opens the SockJS connection.
// Not the same as the SockJS session id! We can't use that as that is generated
// on the client side and we don't have it yet at this point.
func genTerminalSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}

// isValidShell checks if the Shell is an allowed one
func isValidShell(validShells []string, shell string) bool {
	for _, validShell := range validShells {
		if validShell == shell {
			return true
		}
	}
	return false
}

type TerminalSessionRequest struct {
	Shell         string
	SessionId     string
	Namespace     string
	PodName       string
	ContainerName string
	//ApplicationId is helm app Id
	ApplicationId string
	EnvironmentId int
	AppId         int
	//ClusterId is optional
	ClusterId int
}

// WaitForTerminal is called from apihandler.handleAttach as a goroutine
// Waits for the SockJS connection to be opened by the client the session to be bound in handleTerminalSession
func WaitForTerminal(k8sClient kubernetes.Interface, cfg *rest.Config, request *TerminalSessionRequest) {

	select {
	case <-terminalSessions.Get(request.SessionId).bound:
		close(terminalSessions.Get(request.SessionId).bound)

		var err error
		validShells := []string{"bash", "sh", "powershell", "cmd"}

		if isValidShell(validShells, request.Shell) {
			cmd := []string{request.Shell}

			err = startProcess(k8sClient, cfg, cmd, terminalSessions.Get(request.SessionId), request)
		} else {
			// No Shell given or it was not valid: try some shells until one succeeds or all fail
			// FIXME: if the first Shell fails then the first keyboard event is lost
			for _, testShell := range validShells {
				cmd := []string{testShell}
				if err = startProcess(k8sClient, cfg, cmd, terminalSessions.Get(request.SessionId), request); err == nil {
					break
				}
			}
		}

		if err != nil {
			terminalSessions.Close(request.SessionId, 2, err.Error())
			return
		}

		terminalSessions.Close(request.SessionId, 1, "Process exited")
	}
}

type TerminalSessionHandler interface {
	GetTerminalSession(req *TerminalSessionRequest) (statusCode int, message *TerminalMessage, err error)
}
type TerminalSessionHandlerImpl struct {
	environmentService cluster.EnvironmentService
	clusterService     cluster.ClusterService
	logger             *zap.SugaredLogger
}

func NewTerminalSessionHandlerImpl(environmentService cluster.EnvironmentService, clusterService cluster.ClusterService,
	logger *zap.SugaredLogger) *TerminalSessionHandlerImpl {
	return &TerminalSessionHandlerImpl{
		environmentService: environmentService,
		clusterService:     clusterService,
		logger:             logger,
	}
}
func (impl *TerminalSessionHandlerImpl) GetTerminalSession(req *TerminalSessionRequest) (statusCode int, message *TerminalMessage, err error) {
	sessionID, err := genTerminalSessionId()
	if err != nil {
		statusCode := http.StatusInternalServerError
		statusError, ok := err.(*errors.StatusError)
		if ok && statusError.Status().Code > 0 {
			statusCode = int(statusError.Status().Code)
		}
		return statusCode, nil, err
	}
	req.SessionId = sessionID
	terminalSessions.Set(sessionID, TerminalSession{
		id:       sessionID,
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})
	config, client, err := impl.getClientConfig(req)
	if err != nil {
		impl.logger.Errorw("error in fetching config", "err", err)
		return http.StatusInternalServerError, nil, err
	}
	go WaitForTerminal(client, config, req)
	return http.StatusOK, &TerminalMessage{SessionID: sessionID}, nil
}

func (impl *TerminalSessionHandlerImpl) getClientConfig(req *TerminalSessionRequest) (*rest.Config, *kubernetes.Clientset, error) {
	var clusterBean *cluster.ClusterBean
	var err error
	if req.ClusterId != 0 {
		clusterBean, err = impl.clusterService.FindById(req.ClusterId)
		if err != nil {
			impl.logger.Errorw("error in fetching cluster detail", "envId", req.EnvironmentId, "err", err)
			return nil, nil, err
		}
	} else if req.EnvironmentId != 0 {
		clusterBean, err = impl.environmentService.FindClusterByEnvId(req.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error in fetching cluster detail", "envId", req.EnvironmentId, "err", err)
			return nil, nil, err
		}
	} else {
		return nil, nil, fmt.Errorf("not able to find cluster-config")
	}
	config, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		impl.logger.Errorw("error in config", "err", err)
		return nil, nil, err
	}
	cfg := &rest.Config{}
	cfg.Host = config.Host
	cfg.BearerToken = config.BearerToken
	cfg.Insecure = true
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		impl.logger.Errorw("error in clientSet", "err", err)
		return nil, nil, err
	}
	return cfg, clientSet, nil
}
