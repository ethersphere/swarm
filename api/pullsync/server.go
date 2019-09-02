package pullsync

// pullSync.Server implements stream.Provider
// uses localstore SubscribePull for the  bins
// server is node-wide
type Server struct {
	// ...
	*stream.LocalProvider
}
