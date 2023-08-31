package arc

import (
	"fmt"
)

var (
	ErrNotConnected  = ErrCode_NotConnected.Error("not connected")
	ErrStreamClosed  = ErrCode_NotConnected.Error("stream closed")
	ErrCellNotFound  = ErrCode_CellNotFound.Error("cell not found")
	ErrPinCtxClosed  = ErrCode_PinContextClosed.Error("pin context closed")
	ErrNotPinnable   = ErrCode_PinFailed.Error("not pinnable")
	ErrUnimplemented = ErrCode_PinFailed.Error("not implemented")
	ErrBadCellTx     = ErrCode_MalformedTx.Error("missing cell spec / target cell")
	ErrNothingToPin  = ErrCode_PinFailed.Error("nothing to pin")
	ErrShuttingDown  = ErrCode_ShuttingDown.Error("shutting down")
	ErrTimeout       = ErrCode_Timeout.Error("timeout")
	ErrNoAuthToken   = ErrCode_AuthFailed.Error("no auth token")
)

// Error makes our custom error type conform to a standard Go error
func (err *Err) Error() string {
	codeStr, exists := ErrCode_name[int32(err.Code)]
	if !exists {
		codeStr = ErrCode_name[int32(ErrCode_UnnamedErr)]
	}

	if len(err.Msg) == 0 {
		return codeStr
	}

	return err.Msg + " / " + codeStr
}

// Error returns an *Err with the given error code
func (code ErrCode) Error(msg string) error {
	if code == ErrCode_NoErr {
		return nil
	}
	return &Err{
		Code: code,
		Msg:  msg,
	}
}

// Errorf returns an *Err with the given error code and msg.
// If one or more args are given, msg is used as a format string
func (code ErrCode) Errorf(format string, msgArgs ...interface{}) error {
	if code == ErrCode_NoErr {
		return nil
	}

	err := &Err{
		Code: code,
	}
	if len(msgArgs) == 0 {
		err.Msg = format
	} else {
		err.Msg = fmt.Sprintf(format, msgArgs...)
	}

	return err
}

// Wrap returns a ReqErr with the given error code and "cause" error
func (code ErrCode) Wrap(cause error) error {
	if cause == nil {
		return nil
	}
	return &Err{
		Code: code,
		Msg:  cause.Error(),
	}
}

func GetErrCode(err error) ErrCode {
	if err == nil {
		return ErrCode_NoErr
	}

	if arcErr, ok := err.(*Err); ok {
		return arcErr.Code
	}

	return ErrCode_UnnamedErr
}
