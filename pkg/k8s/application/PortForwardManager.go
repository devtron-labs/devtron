package application

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"go.uber.org/zap"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	proxy "k8s.io/kubectl/pkg/proxy"
	"k8s.io/kubectl/pkg/util/podutils"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"
)

type PortForwardManager interface {
	ForwardPort(ctx context.Context, request bean.PortForwardRequest) (int, error)
	StopPortForwarding(ctx context.Context, port int) bool
	StartK8sProxy(ctx context.Context, clusterId int, handleK8sApiProxyError func(clusterId int)) (int, error)
}

type PortForwardManagerImpl struct {
	logger           *zap.SugaredLogger
	k8sCommonService k8s.K8sCommonService
	stopChannels     map[int]chan struct{}
}

func NewPortForwardManagerImpl(logger *zap.SugaredLogger, k8sCommonService k8s.K8sCommonService) (*PortForwardManagerImpl, error) {
	return &PortForwardManagerImpl{
		logger:           logger,
		k8sCommonService: k8sCommonService,
		stopChannels:     make(map[int]chan struct{}),
	}, nil
}

func (impl *PortForwardManagerImpl) StopPortForwarding(ctx context.Context, port int) bool {
	if stopChannel, ok := impl.stopChannels[port]; ok {
		close(stopChannel)
		delete(impl.stopChannels, port)
		return true
	}
	return false
}

func (impl *PortForwardManagerImpl) StartK8sProxy(ctx context.Context, clusterId int, handleK8sApiProxyError func(clusterId int)) (int, error) {
	config, err, _ := impl.k8sCommonService.GetRestConfigByClusterId(ctx, clusterId)
	if err != nil {
		return 0, err
	}
	var keepAlive time.Duration
	apiProxyServer, err := proxy.NewServer("", "/", "", nil, config, keepAlive, false)
	if err != nil {
		return 0, err
	}
	unUsedPort, err := impl.getUnUsedPort()
	if err != nil {
		return 0, err
	}
	// Separate listening from serving so we can report the bound port
	// when it is chosen by os (eg: port == 0)
	var listener net.Listener
	listener, err = apiProxyServer.Listen("localhost", unUsedPort)
	if err != nil {
		return 0, err
	}
	impl.logger.Infow("Starting to serve k8s api proxy server", "addr", listener.Addr().String())
	stopChannel := make(chan struct{})
	impl.stopChannels[unUsedPort] = stopChannel
	go func() {
		err = apiProxyServer.ServeOnListener(listener)
		select {
		case <-stopChannel: // In case proxy server is stopped due to inactivity by calling stop channel, we simply return and do not log the error
			handleK8sApiProxyError(clusterId)
			return
		default:
			if err != nil {
				handleK8sApiProxyError(clusterId)
				impl.logger.Errorw("An error occurred while listening to k8s api proxy server", "err", err)
			}
			return
		}
	}()
	go func() {
		select {
		case <-stopChannel:
			err := listener.Close()
			if err != nil {
				impl.logger.Errorw("An error occurred while closing k8s api server", "err", err)
			}
			return
		}
	}()
	return unUsedPort, nil
}

func (impl *PortForwardManagerImpl) ForwardPort(ctx context.Context, request bean.PortForwardRequest) (int, error) {
	pod, service, err := impl.getServicePod(ctx, request)
	if err != nil {
		return 0, err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return 0, fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	portForwarderOps := NewDefaultPortForwardOptions()
	portForwarderOps.Address = []string{"localhost"}

	portString, portToAllocate := impl.getPortString(request.TargetPort)
	portForwarderOps.Ports, err = translateServicePortToTargetPort([]string{portString}, *service, *pod)

	//returnCtx, returnCtxCancel := context.WithCancel(ctx)
	//defer returnCtxCancel()

	go func() {
		select {
		case <-signals:
			//case <-returnCtx.Done():
		}
		if portForwarderOps.StopChannel != nil {
			close(portForwarderOps.StopChannel)
		}
	}()

	config, err, _ := impl.k8sCommonService.GetRestConfigByClusterId(ctx, request.ClusterId)
	if err != nil {
		return 0, err
	}
	err = setKubernetesDefaults(config)
	if err != nil {
		return 0, err
	}
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return 0, err
	}
	req := restClient.Post().
		Resource("pods").
		Namespace(request.Namespace).
		Name(pod.Name).
		SubResource("portforward")
	portForwarderOps.Config = config
	portFwdErrChan := make(chan error, 1) //using buffered channel
	go func() {
		portForwardErr := portForwarderOps.PortForwarder.ForwardPorts("POST", req.URL(), portForwarderOps)
		if portForwardErr != nil {
			impl.logger.Errorw("error occurred while forwarding port request", "err", err)
			portFwdErrChan <- portForwardErr
		}
	}()
	var portForwardErr error
	select {
	case <-portForwarderOps.ReadyChannel:
	case portForwardErr = <-portFwdErrChan:
	}
	impl.stopChannels[portToAllocate] = portForwarderOps.StopChannel
	portForwarderOps.forwardedPort = portToAllocate
	return portToAllocate, portForwardErr
}

func (impl *PortForwardManagerImpl) getServicePod(ctx context.Context, request bean.PortForwardRequest) (*corev1.Pod, *corev1.Service, error) {
	_, v1Client, err := impl.k8sCommonService.GetCoreClientByClusterId(request.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting coreV1 client by clusterId", "clusterId", request.ClusterId, "err", err)
		return nil, nil, err
	}
	service, err := v1Client.Services(request.Namespace).Get(ctx, request.ServiceName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	namespace, selector, err := polymorphichelpers.SelectorsForObject(service)
	if err != nil {
		return nil, nil, err
	}
	sortBy := func(pods []*corev1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	getPodTimeout := 60 * time.Second
	firstPod, _, err := polymorphichelpers.GetFirstPod(v1Client, namespace, selector.String(), getPodTimeout, sortBy)
	pod, err := v1Client.Pods(request.Namespace).Get(ctx, firstPod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return pod, service, nil
}

func (impl *PortForwardManagerImpl) getPortString(servicePort string) (string, int) {
	unUsedport, _ := impl.getUnUsedPort()
	return strconv.Itoa(unUsedport) + ":" + servicePort, unUsedport
}

func (impl *PortForwardManagerImpl) getUnUsedPort() (int, error) {
	// handle case where all ports are used
	//if len(impl.stopChannels) >= 1000 {
	//	return -1
	//}
	//randomPort := randRange(7000, 8000)
	//if _, ok := impl.stopChannels[randomPort]; !ok {
	//	return randomPort
	//}
	//return impl.getUnUsedPort()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		impl.logger.Errorw("error occurred while generating random port", "err", err)
		return -1, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err = ln.Close(); err != nil {
		impl.logger.Warnf("failed to close %v: %v", ln, err)
	}
	return port, nil
}

type portForwarder interface {
	ForwardPorts(method string, url *url.URL, opts *PortForwardOptions) error
}

func NewDefaultPortForwardOptions() *PortForwardOptions {
	portFwdOptions := &PortForwardOptions{
		StopChannel:  make(chan struct{}, 1),
		ReadyChannel: make(chan struct{}),
	}
	streams := IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	portFwdOptions.PortForwarder = &defaultPortForwarder{IOStreams: streams}
	return portFwdOptions
}

type defaultPortForwarder struct {
	IOStreams
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts *PortForwardOptions) error {
	transport, upgrader, err := spdy.RoundTripperFor(opts.Config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforward.NewOnAddresses(dialer, opts.Address, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

//type ActivityLogger interface {
//	logActivity(response []byte)
//}

// PortForwardOptions contains all the options for running the port-forward cli command.
type PortForwardOptions struct {
	//ActivityLogger
	Namespace             string
	PodName               string
	Config                *rest.Config
	Address               []string // []string{"localhost"}
	Ports                 []string
	PortForwarder         portForwarder
	StopChannel           chan struct{}
	ReadyChannel          chan struct{}
	lastActivityTimestamp time.Time
	forwardedPort         int
}

func (opts *PortForwardOptions) logActivity(response []byte) {
	stdResponse := string(response)
	fmt.Println("response", stdResponse)
	if strings.Contains(stdResponse, "Handling connection") {
		opts.lastActivityTimestamp = time.Now()
	}
}

// IOStreams provides the standard names for iostreams.  This is useful for embedding and for unit testing.
// Inconsistent and different names make it hard to read and review code
type IOStreams struct {
	// In think, os.Stdin
	In io.Reader
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

type customWriter struct {
	io.Writer
	//activityLogger ActivityLogger
}

func NewCustomWriter() io.Writer {
	return &customWriter{}
}

func (writer *customWriter) Write(p []byte) (n int, err error) {
	//writer.activityLogger.logActivity(p)
	return len(p), nil
}
