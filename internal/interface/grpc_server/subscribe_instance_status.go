package grpc_server

type InstanceStatusResponse struct {
	Name    string `json:"name"`
	Version int32  `json:"version"`
	Online  bool   `json:"online"`

	Databases map[string]*DbStatusResponse `json:"all_db_statuses"`
}

type SubscribeInstanceStatusFunc func(*InstanceStatusResponse) bool

type SubscribeInstanceStatus interface {
	SubscribeInstanceStatus(callback SubscribeInstanceStatusFunc)
}
