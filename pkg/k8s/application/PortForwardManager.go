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
	"k8s.io/kubectl/pkg/util/podutils"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"time"
)

type PortForwardManager interface {
	ForwardPort(ctx context.Context, request bean.PortForwardRequest) error
}

type PortForwardManagerImpl struct {
	logger           *zap.SugaredLogger
	k8sCommonService k8s.K8sCommonService
}

func NewPortForwardManagerImpl(logger *zap.SugaredLogger, k8sCommonService k8s.K8sCommonService) (PortForwardManager, error) {
	return &PortForwardManagerImpl{
		logger:           logger,
		k8sCommonService: k8sCommonService,
	}, nil
}

func (impl *PortForwardManagerImpl) ForwardPort(ctx context.Context, request bean.PortForwardRequest) error {
	pod, service, err := impl.getServicePod(ctx, request)
	if err != nil {
		return err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	portForwarderOps := NewDefaultPortForwardOptions()
	portForwarderOps.Address = []string{"localhost"}
	portForwarderOps.Ports, err = translateServicePortToTargetPort(request.PortString, *service, *pod)

	returnCtx, returnCtxCancel := context.WithCancel(ctx)
	defer returnCtxCancel()

	go func() {
		select {
		case <-signals:
		case <-returnCtx.Done():
		}
		if portForwarderOps.StopChannel != nil {
			close(portForwarderOps.StopChannel)
		}
	}()

	config, err, _ := impl.k8sCommonService.GetRestConfigByClusterId(ctx, request.ClusterId)
	if err != nil {
		return err
	}
	err = setKubernetesDefaults(config)
	if err != nil {
		return err
	}
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return err
	}
	req := restClient.Post().
		Resource("pods").
		Namespace(request.Namespace).
		Name(pod.Name).
		SubResource("portforward")
	portForwarderOps.Config = config
	return portForwarderOps.PortForwarder.ForwardPorts("POST", req.URL(), portForwarderOps)
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

type portForwarder interface {
	ForwardPorts(method string, url *url.URL, opts *PortForwardOptions) error
}

func NewDefaultPortForwardOptions() *PortForwardOptions {
	streams := IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	return &PortForwardOptions{
		PortForwarder: &defaultPortForwarder{
			IOStreams: streams,
		},
	}
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

// PortForwardOptions contains all the options for running the port-forward cli command.
type PortForwardOptions struct {
	Namespace     string
	PodName       string
	Config        *rest.Config
	Address       []string // []string{"localhost"}
	Ports         []string
	PortForwarder portForwarder
	StopChannel   chan struct{}
	ReadyChannel  chan struct{}
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
