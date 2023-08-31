package arc

import (
	"net/url"

	"github.com/arcspace/go-arc-sdk/stdlib/symbol"
	"github.com/arcspace/go-arc-sdk/stdlib/task"
)

// Host is the highest level controller.
// Child processes attach to it and start new host sessions as needed.
type Host interface {
	task.Context

	Registry() Registry

	// StartNewSession creates a new HostSession and binds its Msg transport to a stream.
	StartNewSession(parent HostService, via Transport) (HostSession, error)
}

// Transport wraps a Msg transport abstraction, allowing a Host to connect over any data transport layer.
// This is intended to be implemented by a grpc and other transport layers.
type Transport interface {

	// Describes this stream
	Desc() string

	// Called when this stream to be closed because the associated parent host session is closing or has closed.
	Close()

	// SendMsg sends a Msg to the remote client.
	// ErrStreamClosed is used to denote normal stream close.
	// Like grpc.Transport.SendMsg(), on exit, the Msg has been copied and so can be reused.
	SendMsg(m *Msg) error

	// RecvMsg blocks until it receives a Msg or the stream is done.
	// ErrStreamClosed is used to denote normal stream close.
	RecvMsg() (*Msg, error)
}

// HostService attaches to a arc.Host as a child task, extending host functionality.
// For example. it wraps a Grpc-based Msg transport as well as a dll-based Msg transport implementation.
type HostService interface {
	task.Context

	// Returns short string identifying this service
	ServiceURI() string

	// Returns the parent Host this extension is attached to.
	Host() Host

	// StartService attaches a child task to a Host and starts this HostService.
	StartService(on Host) error

	// GracefulStop initiates a polite stop of this extension and blocks until it's in a "soft" closed state,
	//    meaning that its service has effectively stopped but its Context is still open.
	// Note this could any amount of time (e.g. until all open requests are closed)
	// Typically, GracefulStop() is called (blocking) and then Context.Close().
	// To stop immediately, Context.Close() is always available.
	GracefulStop()
}

// HostSession in an open session instance with a Host.
// Closing is initiated via Context.Close().
type HostSession interface {
	task.Context    // Underlying task context
	SessionRegistry // How an AppInstance resolves symbols and types

	// Called when this session is newly opened to set up the SessionRegistry
	InitSessionRegistry(symTable symbol.Table)

	// Returns the running AssetPublisher instance for this session.
	AssetPublisher() AssetPublisher

	// Returns info about this user and session
	LoginInfo() Login

	// Sends a readied Msg to the client for handling.
	// If msg.ReqID == 0, the attr is sent to the client's session controller (for sending session meta messages).
	// On exit, the given msg should not be referenced further.
	SendMsg(msg *Msg) error

	// PinCell resolves and pins a requested cell.
	PinCell(req PinReq) (PinContext, error)

	// Gets the currently running AppInstance for an AppID.
	// If the requested app is not running and autoCreate is set, a new instance is created and started.
	GetAppInstance(appID UID, autoCreate bool) (AppInstance, error)
}

// SessionRegistry manages a HostSession's symbol and type definitions.
// All calls are safe to call from multiple goroutines.
type SessionRegistry interface {

	// Returns the symbol table for a session.
	ClientSymbols() symbol.Table

	// Issues a monotonically increasing UTC16 timestamp (guaranteed never to have been issued before).
	// This is often just the current timestamp, but when multiple timestamps are rapidly issued then the next available UTC16 is issued by adding a tick.
	// This means even during intense TimeID issuance, TimeID will be unique and "caught up" after a negligible period of time.
	IssueTimeID() TimeID

	// Translates a native symbol ID to a client symbol ID, returning false if not found.
	NativeToClientID(nativeID uint32) (clientID uint32, found bool)

	// Registers an ElemVal as a prototype under its element type name..
	// This and ResolveAttrSpec() allow NewElemVal() to work.
	RegisterElemType(prototype ElemVal) error

	// Registers a block of symbol, attr, cell, and selector definitions for a client.
	RegisterDefs(defs *RegisterDefs) error

	// Resolves an AttrSpec into useful symbols, auto-registering the AttrSpec as needed.
	// Typically used during AppInstance.OnNew() to get the AttrIDs that correspond to the AttrSpecs it will send later.
	//
	// If native === true, the spec is resolved with native symbols (vs client symbols).
	//
	// See AttrSpec docs.
	ResolveAttrSpec(attrSpec string, native bool) (AttrSpec, error)

	// Resolves a CellSpec (a cell attr schema) into symbol IDs, auto-registering each as needed.
	// Called by apps to resolve cell types it supports, getting a CellSpec ID to stamp cells it pushes to clients.
	//
	// See CellSpec docs.
	ResolveCellSpec(cellSpec string) (CellDef, error)

	// Instantiates an attr element value for an AttrID -- typically followed by ElemVal.Unmarshal()
	NewAttrElem(attrDefID uint32, native bool) (ElemVal, error)
}

// Registry maps an app ID to an AppModule.    It is safe to access from multiple goroutines.
type Registry interface {

	// Registers an ElemVal as a prototype under its AttrElemType (also a valid AttrSpec type expression).
	// If an entry already exists (common for a type used by multiple apps), an error is returned and is a no-op.
	// This and ResolveAttrSpec() allow NewElemVal() to work.
	RegisterElemType(prototype ElemVal)

	// When a HostSession creates a new SessionRegistry(), this populates it with its registered ElemTypes.
	ExportTo(dst SessionRegistry) error

	// Registers an app by its UUID, URI, and schemas it supports.
	RegisterApp(app *AppModule) error

	// Looks-up an app by UUID
	GetAppByUID(appUID UID) (*AppModule, error)

	// Selects the app that best matches an invocation string.
	GetAppForInvocation(invocation string) (*AppModule, error)
}

// NewRegistry returns a new Registry
func NewRegistry() Registry {
	return newRegistry()
}

// PinContext wraps a client request to receive a cell's state / updates.
type PinContext interface {
	task.Context // Started as a CHILD of the arc.PinnedCell returned by AppInstance.PinCell()

	PinReq // Originating request info

	// PushTx pushes the given tx to this PinContext
	PushUpdate(tx *Msg) error

	App() AppContext // Parent app of the cell associated with this context
}

type PinReq interface {
	Params() *PinReqParams
	URLPath() []string
}

// PinReqParams implements PinReq
type PinReqParams struct {
	PinReq   PinRequest
	URL      *url.URL
	Target   CellID
	ReqID    uint64    // Request ID needed to route to the originator
	LogLabel string    // info string for logging and debugging
	Outlet   chan *Msg // send to this channel to send to the originator

}
