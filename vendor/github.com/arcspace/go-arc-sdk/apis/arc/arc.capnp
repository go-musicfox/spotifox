@0x9aff325096b39f47;

using Go = import "/go.capnp";
$Go.package("arc");
$Go.import("github.com/arcspace/go-arc-sdk");

using CSharp = import "/csharp.capnp";
$CSharp.namespace("Arcspace");

struct AttrDefTest {
    typeName @0 :Text;
    typeID   @1 :Int32;   
}


enum CellTxOp2 {
    noOp @0;
    insertChild @1;
    upsertChild @2;
    deleteChild @3;
    deleteCell  @4;
    checkpoint  @5;
}


struct MultiTxCp {
    reqID          @0 :Int64;
    cellTxs        @1 :List(CellTxCp);
    
    struct CellTxCp {
        op          @0 :CellTxOp2;
        cellSpec    @1 :UInt32;
        cellID      @2 :Int64;
        elems       @3 :AttrElemCp;
        
        struct AttrElemCp {
            attrID       @0 :UInt32;
            seriesIndex  @1 :Int64;
            valBuf       @2 :Data;
        }
    }
}

