package arc

import (
	"net/url"

	"github.com/arcspace/go-arc-sdk/stdlib/task"
)

// AppModule declares a 3rd-party module this is registered with an archost.
//
// An app can be invoked by:
//   - a client pinning a cell with a data model that the app handles
//   - a client or other app invoking its UID or URI directly
type AppModule struct {

	// AppID identifies this app with form "{AppNameID}.{FamilyID}.{PublisherID}" -- e.g. "filesys.amp.arcspace.systems"
	//   - PublisherID: typically the domain name of the publisher of this app -- e.g. "arcspace.systems"
	//   - FamilyID:    encompassing namespace ID used to group related apps and content (no spaces or punctuation)
	//   - AppNameID:   uniquely identifies this app within its parent family and domain (no spaces or punctuation)
	//
	// AppID form is consistent of a URL domain name (and subdomains).
	AppID        string
	UID          UID      // Universally unique and persistent ID for this module (and the module's "home" planet if present)
	Desc         string   // Human-readable description of this app
	Version      string   // "v{MajorVers}.{MinorID}.{RevID}"
	Dependencies []UID    // Module UIDs this app may access
	Invocations  []string // Additional aliases that invoke this app
	AttrDecl     []string // Attrs to be resolved and registered with the HostSession -- get the registered

	// Called when an App is invoked on an active User session and is not yet running.
	NewAppInstance func() AppInstance
}

// AppContext hosts is provided by the arc runtime and hosts an AppInstance.
//
// An AppModule retains the AppContext it is given via NewAppInstance() for:
//   - archost operations (e.g. resolve type schemas, publish assets for client consumption -- see AppContext)
//   - having a context to select{} against (for graceful shutdown)
type AppContext interface {
	task.Context
	AssetPublisher        // Allows an app to publish assets for client consumption
	Session() HostSession // Access to underlying Session

	// Returns the absolute fs path of the app's local state directory.
	// This directory is scoped by the app's UID and is unique to this app instance.
	LocalDataPath() string

	// Atomically issues a new and unique ID that will remain globally unique for the duration of this session.
	// An ID may still expire, go out of scope, or otherwise become meaningless.
	IssueCellID() CellID

	// Allows an app resolve attrs by name, etc

	// Gets the named cell and attribute from the user's home planet -- used high-level app settings.
	// The attr is scoped by both the app UID so key collision with other users or apps is not possible.
	GetAppCellAttr(attrSpec string, dst ElemVal) error

	// Write analog for GetAppCellAttr()
	PutAppCellAttr(attrSpec string, src ElemVal) error
}

// AppInstance is implemented by an arc app (AppModule)
type AppInstance interface {
	AppContext // An app instance implies an underlying host AppContext

	// Callback made immediately after AppModule.NewAppInstance() -- typically resolves app-specific type specs.
	OnNew(this AppContext) error

	// Celled when the app is pin the cell IAW with the given request.
	// If parent != nil, this is the context of the request.
	// If parent == nil, this app was invoked without out a parent cell / context.
	PinCell(parent PinnedCell, req PinReq) (PinnedCell, error)

	// Handles a meta message sent to this app, which could be any attr type.
	HandleURL(*url.URL) error

	// Called exactly once if / when an app is signaled to close.
	OnClosing()
}

// Cell is an interface for an app Cell
type Cell interface {

	// Returns this cell's immutable info
	Info() CellID
}

// PinnedCell is how your app encapsulates a pinned cell to the archost runtime and thus clients.
type PinnedCell interface {
	Cell

	// Apps spawn a PinnedCell as a child task.Context of arc.AppContext.Context or as a child of another PinnedCell.
	// This means an AppContext contains all its PinnedCells and thus Close() will close all PinnedCells.
	Context() task.Context

	// Pins the requested cell (typically a child cell).
	PinCell(req PinReq) (PinnedCell, error)

	// Pushes this cell and child cells to the client state is called.
	// Exits when any of the following occur:
	//   - ctx.Closing() is signaled,
	//   - a fatal error is encountered, or
	//   - state has been pushed to the client AND ctx.MaintainSync() == false
	ServeState(ctx PinContext) error

	// Merges a set of incoming changes into this pinned cell. -- "write" operation
	MergeUpdate(tx *Msg) error
}

type PbValue interface {
	Size() int
	MarshalToSizedBuffer(dAtA []byte) (int, error)
	Unmarshal(dAtA []byte) error
}

type ElemVal interface {

	// Returns the element type name (a degenerate AttrSpec).
	TypeName() string

	// Marshals this ElemVal to a buffer, reallocating if needed.
	MarshalToBuf(dst *[]byte) error

	// Unmarshals and merges value state from a buffer.
	Unmarshal(src []byte) error

	// Creates a default instance of this same ElemVal type
	New() ElemVal
}

// MultiTx is a state update for a pinned cell or a container of meta attrs.
type MultiTx struct {
	ReqID   uint64 // allows replies to be routed to the originator
	Status  ReqStatus
	CellTxs []CellTx
	//CellTxsPb []*CellTxPb // serialized version of CellTxs -- this goes away when MultiTx has its own serializer.
	//CellTxsBuf []byte  // TODO: use to denote serialization
}

// CellTx is a data super wrapper for arbitrary complexity and size data structures
type CellTx struct {
	Op         CellTxOp      // Op is the cell tx operation to perform
	TargetCell CellID        // Target ID of the cell being modified
	ElemsPb    []*AttrElemPb // Attr element run (serialized)
	//Elems      []AttrElem    // Attrs elements to/from target cell
}

type AttrElem struct {
	Val    ElemVal // Val is the abstraction interface allowing serialization and type string-ification
	SI     int64   // SI is the SeriesIndex, which is described in the AttrSpec.SeriesIndexType
	AttrID uint32  // AttrID is the native ID (AttrSpec.DefID) that fully names an AttrSpec
}

type AttrDef struct {
	Client AttrSpec
	Native AttrSpec
}

type CellDef struct {
	ClientDefID uint32    // READ-ONLY
	NativeDefID uint32    // READ-ONLY
	CommonAttrs []AttrDef // READ-ONLY
	PinnedAttrs []AttrDef // READ-ONLY
}
