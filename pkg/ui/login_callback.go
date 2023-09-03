package ui

import "github.com/anhoder/foxful-cli/model"

type LoginCallback func() model.Page

func EnterMenuCallback(m *model.Main) LoginCallback {
	return func() model.Page {
		return m.EnterMenu(nil, nil)
	}
}

func BottomOutHookCallback(main *model.Main, m model.Menu) LoginCallback {
	return func() model.Page {
		b := m.BottomOutHook()
		if b == nil {
			return nil
		}
		res, newPage := b(main)
		if !res {
			return nil
		}
		main.RefreshMenuList()
		return newPage
	}
}
