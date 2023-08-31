package arc

import "net/url"

// This file contains types and interfaces intended to ease an arc app development.
// These types are not required to be used, but are provided as a convenience.

// AppBase is a helper for implementing AppInstance.
// It is typically extended by embedding it into a struct that builds on top of it.
type AppBase struct {
	AppContext
	CellTextAttr   uint32
	CellHeaderAttr uint32
	CellPosAttr    uint32
}

func (app *AppBase) OnNew(ctx AppContext) error {
	app.AppContext = ctx

	var err error
	if app.CellHeaderAttr, err = app.ResolveAppAttr("CellHeader"); err != nil {
		return err
	}
	if app.CellTextAttr, err = app.ResolveAppAttr("[Locale.Name]CellText"); err != nil {
		return err
	}
	if app.CellPosAttr, err = app.ResolveAppAttr("[Surface.Name]Position"); err != nil {
		return err
	}
	return nil
}

func (app *AppBase) HandleURL(*url.URL) error {
	return ErrUnimplemented
}

func (app *AppBase) OnClosing() {

}

// ResolveAppCell is a convenience function that resolves a cell spec into a CellSpec def ID.
func (app *AppBase) ResolveAppCell(cellSpec string) (cellSpecID uint32, err error) {
	cellDef, err := app.AppContext.Session().ResolveCellSpec(cellSpec)
	if err != nil {
		return
	}
	return cellDef.ClientDefID, nil
}

// ResolveAppAttr is a convenience function that resolves an attr spec intended to be sent to the client.
func (app *AppBase) ResolveAppAttr(attrSpec string) (uint32, error) {
	spec, err := app.AppContext.Session().ResolveAttrSpec(attrSpec, false)
	if err != nil {
		return 0, err
	}
	return spec.DefID, nil
}

func (app *AppBase) RegisterElemType(prototype ElemVal) error {
	err := app.AppContext.Session().RegisterElemType(prototype)
	if err != nil {
		return err
	}
	return nil
}
