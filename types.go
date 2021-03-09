package shared

// RepairRequest definition
type RepairRequest struct {
	Namespace string `json:"namespace,omitempty"` // e.g. default
	Name      string `json:"name,omitempty"`      // e.g. deployment name
	Type      string `json:"type,omitempty"`      // deployment, statefulset etc.
	Replicas  int32  `json:"replicas,omitempty"`
}
