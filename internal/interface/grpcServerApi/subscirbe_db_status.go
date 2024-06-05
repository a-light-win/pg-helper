package grpcServerApi

const (
	ContinueSubscribe bool = true
	StopSubscribe     bool = false
)

// SubscribeDbStatusFunc is a callback function to be called when the status of the database changes
// The function should return true if it wants to continue receiving notifications
// and false if it wants to unsubscribe
type SubscribeDbStatusFunc func(*DbStatusResponse) bool

type SubscribeDbStatus interface {
	// Subscribe to database status changes, the callback will be called when the status changes
	// The callback should return true if it wants to continue receiving notifications
	// and false if it wants to unsubscribe
	//
	// This function will report on following stage changed:
	// - Ready
	// - Idle
	// - DropCompleted
	SubscribeDbStatus(callback SubscribeDbStatusFunc)
}
