package bean

type ErrorResponse struct {
	Kind    string `json:"kind"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
	//Details struct {
	//	ClusterName  string `json:"cluster_name"`
	//	Group        string `json:"group"`
	//	Version      string `json:"version"`
	//	Kind         string `json:"kind"`
	//	ResourceName string `json:"resource_name"`
	//	Namespace    string `json:"namespace"`
	//	Role         string `json:"role"`
	//	Causes       []struct {
	//		Reason  string `json:"reason"`
	//		Message string `json:"message"`
	//	} `json:"causes"`
	//} `json:"details"`

	//errorResponse.Details.ClusterName = clusterRequested.ClusterName
	//errorResponse.Details.Group = gvk.Group
	//errorResponse.Details.Version = gvk.Version
	//errorResponse.Details.Kind = gvk.Kind
	//errorResponse.Details.Namespace = namespace
	//errorResponse.Details.ResourceName = resourceName
	//errorResponse.Details.Causes = []struct {
	//	Reason  string `json:"reason"`
	//	Message string `json:"message"`
	//}{
	//	{
	//		Reason:  "You do not have permission.",
	//		Message: "Please ask admin for specific permission. * represents all.",
	//	},
	//}
	//w.Header().Set("Content-Type", "application/json")
	//if r.Method == http.MethodPost {
	//	errorResponse.Message = fmt.Sprintf("You need %s access on cluster %s, namespace %s, group %s, version %s, kind %s, resource-name %s. * represents all.", errorResponse.Details.Role, clusterRequested.ClusterName, namespace, gvk.Group, gvk.Version, gvk.Kind, resourceName)
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(http.StatusForbidden)
	//	_ = json.NewEncoder(w).Encode(errorResponse)
	//	return
	//}
}
