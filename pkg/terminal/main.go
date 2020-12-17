package terminal

import (
	"flag"
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"path/filepath"
)

const podName = "queenly-numbat-nginx-ingress-controller-8648f4b785-7j9z4"
const namespace = "default"

func serveSocJs() {
	wsContainer := restful.NewContainer()

	apiV1Ws := new(restful.WebService)

	apiV1Ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	wsContainer.Add(apiV1Ws)
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}/{pod}/Shell/{container}").
			To(handleExecShell))

	http.Handle("/api/", wsContainer)

	http.Handle("/api/sockjs/", CreateAttachHandler("/api/sockjs"))

	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("./UI"))))

	http.ListenAndServe(":8080", nil)

}

func handleExecShell(request *restful.Request, response *restful.Response) {
	/*sessionID, err := genTerminalSessionId()
	if err != nil {
		statusCode := http.StatusInternalServerError
		statusError, ok := err.(*errors.StatusError)
		if ok && statusError.Status().Code > 0 {
			statusCode = int(statusError.Status().Code)
		}
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(statusCode, err.Error()+"\n")
	}

	terminalSessions.Set(sessionID, TerminalSession{
		id:       sessionID,
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})
	go WaitForTerminal(clientset, restcfg, request, sessionID)
	response.WriteHeaderAndEntity(http.StatusOK, TerminalMessage{SessionID: sessionID})
*/
}
func main() {
	fmt.Println("starting main")
	clientset, restcfg = GetClientConfig()
	serveSocJs()
}

var clientset *kubernetes.Clientset
var restcfg *rest.Config

func GetClientConfig() (*kubernetes.Clientset, *rest.Config) {

	kubeconfig := flag.String("kubeconfig_cd", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}
	return clientset, config
}
