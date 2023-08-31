package arc

import (
	"sync"
)

// MsgBatch is an ordered list os Msgs
// See NewMsgBatch()
type MsgBatch struct {
	Msgs []*Msg
}

/*
var gMsgBatchPool = sync.Pool{
	New: func() interface{} {
		return &MsgBatch{
			Msgs: make([]*Msg, 0, 16),
		}
	},
}

func NewMsgBatch() *MsgBatch {
	return gMsgBatchPool.Get().(*MsgBatch)
}

func (batch *MsgBatch) Reset(count int) []*Msg {
	if count > cap(batch.Msgs) {
		msgs := make([]*Msg, count)
		copy(msgs, batch.Msgs)
		batch.Msgs = msgs
	} else {
		batch.Msgs = batch.Msgs[:count]
	}

	// Alloc or init  each msg
	for i, msg := range batch.Msgs {
		if msg == nil {
			batch.Msgs[i] = NewMsg()
		} else {
			msg.Init()
		}
	}

	return batch.Msgs
}

func (batch *MsgBatch) AddNew(count int) []*Msg {
	N := len(batch.Msgs)
	for i := 0; i < count; i++ {
		batch.Msgs = append(batch.Msgs, NewMsg())
	}
	return batch.Msgs[N:]
}

func (batch *MsgBatch) AddMsgs(msgs []*Msg) {
	batch.Msgs = append(batch.Msgs, msgs...)
}

func (batch *MsgBatch) AddMsg() *Msg {
	m := NewMsg()
	batch.Msgs = append(batch.Msgs, m)
	return m
}

func (batch *MsgBatch) Reclaim() {
	for i, msg := range batch.Msgs {
		msg.Reclaim()
		batch.Msgs[i] = nil
	}
	batch.Msgs = batch.Msgs[:0]
	gMsgBatchPool.Put(batch)
}

func (batch *MsgBatch) PushCopyToClient(dst PinContext) bool {
	for _, src := range batch.Msgs {
		msg := CopyMsg(src)
		if !dst.PushUpdate(msg) {
			return false
		}
	}
	return true
}
*/

func NewMsg() *Msg {
	msg := gMsgPool.Get().(*Msg)
	return msg
}

/*
	func CopyMsg(src *Msg) *Msg {
		msg := NewMsg()

		if src != nil {
			valBuf := append(msg.ValBuf[:0], src.ValBuf...)
			*msg = *src
			msg.ValBuf = valBuf

		}
		return msg
	}
*/
func (msg *Msg) Init() {
	*msg = Msg{
		CellTxs: msg.CellTxs[:0],
	}
}

func (msg *Msg) Reclaim() {
	if msg != nil {
		msg.Init()
		gMsgPool.Put(msg)
	}
}

/*
	func (msg *Msg) MarshalAttrElem(attrID uint32, src PbValue) error {
		msg.AttrID = attrID
		sz := src.Size()
		if sz > cap(msg.ValBuf) {
			msg.ValBuf = make([]byte, sz, (sz+0x3FF)&^0x3FF)
		} else {
			msg.ValBuf = msg.ValBuf[:sz]
		}
		_, err := src.MarshalToSizedBuffer(msg.ValBuf)
		return err
	}

	func (msg *Msg) UnmarshalValue(dst PbValue) error {
		return dst.Unmarshal(msg.ValBuf)
	}

	func (attr AttrElem) MarshalToMsg(id CellID) (*Msg, error) {
		msg := NewMsg()
		msg.Op = MsgOp_PushAttrElem
		msg.AttrID = attr.AttrID
		msg.SI = attr.SI
		msg.CellID = int64(id)
		err := attr.Val.MarshalToBuf(&msg.ValBuf)
		return msg, err
	}
*/

// type CellMarshaller struct {
// 	Txs []*CellTxPb

// 	marshalBuf []byte
// 	fatalErr   error
// }

var gMsgPool = sync.Pool{
	New: func() interface{} {
		return &Msg{}
	},
}

func (tx *CellTx) Marshal(attrID uint32, SI int64, val ElemVal) {
	if val == nil {
		return
	}
	if attrID == 0 {
		panic("attrID == 0")
	}

	pb := &AttrElemPb{
		AttrID: uint64(attrID),
		SI:     SI,
	}
	err := val.MarshalToBuf(&pb.ValBuf)
	if err != nil {
		panic(err)
	}

	tx.ElemsPb = append(tx.ElemsPb, pb)
}

func (tx *CellTx) Clear(op CellTxOp) {
	tx.Op = op
	tx.TargetCell = 0
	//tx.Elems = tx.Elems[:0]
	tx.ElemsPb = tx.ElemsPb[:0]
}

/*
func (tx *CellTx) MarshalAttrs() error {
	if cap(tx.ElemsPb) < len(tx.Elems) {
		tx.ElemsPb = make([]*AttrElemPb, len(tx.Elems))
	} else {
		tx.ElemsPb = tx.ElemsPb[:len(tx.Elems)]
	}
	for j, srcELem := range tx.Elems {
		elem := tx.ElemsPb[j]
		if elem == nil {
			elem = &AttrElemPb{}
			tx.ElemsPb[j] = elem
		}
		elem.SI = srcELem.SI
		elem.AttrID = uint64(srcELem.AttrID)
		if err := srcELem.Val.MarshalToBuf(&elem.ValBuf); err != nil {
			return err
		}
	}
	return nil
}


func (tx *CellTx) MarshalToPb(dst *CellTxPb) error {
	tx.MarshalAttrs()
	dst.Op = tx.Op
	dst.CellSpec = tx.CellSpec
	dst.TargetCell = int64(tx.TargetCell)
	dst.Elems = tx.ElemsPb
	return nil
}
*/

// If reqID == 0, then this sends an attr to the client's session controller (vs a specific request)
func SendClientMetaAttr(sess HostSession, reqID uint64, val ElemVal) error {
	msg, err := FormClientMetaAttrMsg(sess, val.TypeName(), val)
	msg.ReqID = reqID
	if err != nil {
		return err
	}
	return sess.SendMsg(msg)
}

func FormClientMetaAttrMsg(reg SessionRegistry, attrSpec string, val ElemVal) (*Msg, error) {
	spec, err := reg.ResolveAttrSpec(attrSpec, false)
	if err != nil {
		return nil, err
	}

	return FormMetaAttrTx(spec, val)
}

func FormMetaAttrTx(attrSpec AttrSpec, val ElemVal) (*Msg, error) {
	elemPb := &AttrElemPb{
		AttrID: uint64(attrSpec.DefID),
	}
	if err := val.MarshalToBuf(&elemPb.ValBuf); err != nil {
		return nil, err
	}

	tx := &CellTxPb{
		Op: CellTxOp_MetaAttr,
		Elems: []*AttrElemPb{
			elemPb,
		},
	}

	msg := NewMsg()
	msg.ReqID = 0 // signals a meta message
	msg.Status = ReqStatus_Synced
	msg.CellTxs = append(msg.CellTxs, tx)
	return msg, nil
}

func (msg *Msg) GetMetaAttr() (attr *AttrElemPb, err error) {
	if len(msg.CellTxs) == 0 || msg.CellTxs[0].Op != CellTxOp_MetaAttr || msg.CellTxs[0].Elems == nil || len(msg.CellTxs[0].Elems) == 0 {
		return nil, ErrCode_MalformedTx.Error("missing meta attr")
	}

	return msg.CellTxs[0].Elems[0], nil
}

func (tx *MultiTx) UnmarshalFrom(msg *Msg, reg SessionRegistry, native bool) error {
	tx.ReqID = msg.ReqID
	tx.Status = msg.Status
	tx.CellTxs = tx.CellTxs[:0]

	elemCount := 0

	srcTxs := msg.CellTxs
	if cap(tx.CellTxs) < len(srcTxs) {
		tx.CellTxs = make([]CellTx, len(srcTxs))
	} else {
		tx.CellTxs = tx.CellTxs[:len(srcTxs)]
	}
	for i, cellTx := range srcTxs {
		elems := make([]AttrElem, len(cellTx.Elems))
		for j, srcElem := range cellTx.Elems {
			attrID := uint32(srcElem.AttrID)
			elem := AttrElem{
				SI:     srcElem.SI,
				AttrID: attrID,
			}
			var err error
			elem.Val, err = reg.NewAttrElem(attrID, native)
			if err == nil {
				err = elem.Val.Unmarshal(srcElem.ValBuf)
			}
			if err != nil {
				return err
			}
			elems[j] = elem
			elemCount++
		}

		tx.CellTxs[i] = CellTx{
			Op:         cellTx.Op,
			TargetCell: CellID(cellTx.TargetCell),
			//Elems:      elems,
		}
	}

	if elemCount == 0 {
		return ErrBadCellTx
	}
	return nil
}

/*
// Pushes a attr mutation to the client, returning true if the msg was sent (false if the client has been closed).
func (bat *CellTx) PushBatch(ctx PinContext) error {

	for _, attr := range bat.Attrs {
		msg, err := attr.MarshalToMsg(bat.Target)
		if err != nil {
			ctx.Warnf("MarshalToMsg() err: %v", err)
			continue
		}

		// if i == len(bat.Attrs)-1 {
		// 	msg.Flags |= MsgFlags_CellCheckpoint
		// }

		if !ctx.PushUpdate(msg) {
			return ErrPinCtxClosed
		}
	}

	return nil

}


func (tx *MultiTx) MarshalToBuf(dst *[]byte) error {
	pb := MultiTxPb{
		ReqID:   tx.ReqID,
		CellTxs: make([]*CellTxPb, len(tx.CellTxs)),
	}
	for i, srcTx := range tx.CellTxs {
		cellTx := &CellTxPb{
			Op:           srcTx.Op,
			CellSpec:     srcTx.CellSpec,
			TargetCellID: int64(srcTx.Target),
			Elems:        make([]*AttrElemPb, len(srcTx.Elems)),
		}
		for j, attrElem := range srcTx.Elems {
			//attrElem.ValBuf = make([]byte, attrElem.Val.Marhal
			elem := &AttrElemPb{
				SI:     attrElem.SI,
				AttrID: attrElem.AttrID,
			}
			attrElem.Val.MarshalToBuf(&elem.ValBuf)
			cellTx.Elems[j] = elem
		}
		pb.CellTxs[i] = cellTx
	}
	sz := pb.Size()
	if cap(*dst) < sz {
		*dst = make([]byte, sz)
	} else {
		*dst = (*dst)[:sz]
	}
	_, err := pb.MarshalToSizedBuffer(*dst)
	return err
}



func (v *MultiTxPb) MarshalToBuf(dst *[]byte) error {
	return MarshalPbValueToBuf(v, dst)
}

func (v *MultiTxPb) TypeName() string {
	return "MultiTx"
}

func (v *MultiTxPb) New() ElemVal {
	return &MultiTxPb{}
}

*/
