package ui

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	_ "github.com/arcspace/go-librespot/librespot/core" // bootstrapping
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-musicfox/spotifox/internal/configs"
	"github.com/go-musicfox/spotifox/internal/constants"
	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/internal/structs"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/go-musicfox/spotifox/utils/auth"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

const LoginPageType model.PageType = "login"

const (
	submitIndex = 2 // skip account and password input
	// authIndex   = 3
)

// login tick
type tickLoginMsg struct{}

func tickLogin(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return tickLoginMsg{}
	})
}

type LoginPage struct {
	spotifox *Spotifox

	menuTitle     *model.MenuItem
	index         int
	accountInput  textinput.Model
	passwordInput textinput.Model
	submitButton  string
	// authButton    string
	// authStep int
	tips string

	AfterLogin LoginCallback
}

func NewLoginPage(spotifox *Spotifox) (login *LoginPage) {
	accountInput := textinput.New()
	accountInput.Placeholder = " 账号"
	accountInput.Focus()
	accountInput.Prompt = model.GetFocusedPrompt()
	accountInput.TextStyle = util.GetPrimaryFontStyle()
	accountInput.CharLimit = 32

	passwordInput := textinput.New()
	passwordInput.Placeholder = " 密码"
	passwordInput.Prompt = "> "
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'
	passwordInput.CharLimit = 32

	login = &LoginPage{
		spotifox: spotifox,

		menuTitle:     &model.MenuItem{Title: "用户登录", Subtitle: "账号密码登录"},
		accountInput:  accountInput,
		passwordInput: passwordInput,
		submitButton:  model.GetBlurredSubmitButton(),
	}
	// login.authButton = model.GetBlurredButton(login.authButtonTextByStep())

	return
}

func (l *LoginPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return true
}

func (l *LoginPage) Type() model.PageType {
	return LoginPageType
}

func (l *LoginPage) Update(msg tea.Msg, _ *model.App) (model.Page, tea.Cmd) {
	inputs := []*textinput.Model{
		&l.accountInput,
		&l.passwordInput,
	}

	var (
		key tea.KeyMsg
		ok  bool
	)

	if _, ok = msg.(tickLoginMsg); ok {
		return l, nil
	}

	if key, ok = msg.(tea.KeyMsg); !ok {
		return l.updateLoginInputs(msg)
	}

	switch key.String() {
	case "b":
		// if l.index != submitIndex && l.index != authIndex {
		if l.index != submitIndex {
			return l.updateLoginInputs(msg)
		}
		fallthrough
	case "esc":
		l.tips = ""
		// l.authStep = 0
		// if l.index == authIndex {
		// 	l.authButton = model.GetFocusedButton(l.authButtonTextByStep())
		// } else {
		// 	l.authButton = model.GetBlurredButton(l.authButtonTextByStep())
		// }
		return l.spotifox.MustMain(), l.spotifox.RerenderCmd(true)
	case "tab", "shift+tab", "enter", "up", "down", "left", "right":
		s := key.String()

		// Did the user press enter while the submit button was focused?
		// If so, exit.
		if s == "enter" && l.index >= submitIndex {
			return l.enterHandler()
		}

		// 当focus在button上时，左右按键的特殊处理
		if s == "left" || s == "right" {
			if l.index < submitIndex {
				return l.updateLoginInputs(msg)
			}
			// if s == "left" && l.index == authIndex {
			// 	l.index--
			// } else if s == "right" && l.index == submitIndex {
			// 	l.index++
			// }
		} else if s == "up" || s == "shift+tab" {
			l.index--
		} else {
			l.index++
		}

		// if l.index > authIndex {
		// 	l.index = 0
		// } else if l.index < 0 {
		// 	l.index = authIndex
		// }

		for i := 0; i <= len(inputs)-1; i++ {
			if i != l.index {
				// Remove focused state
				inputs[i].Blur()
				inputs[i].Prompt = model.GetBlurredPrompt()
				inputs[i].TextStyle = lipgloss.NewStyle()
				continue
			}
			// Set focused state
			inputs[i].Focus()
			inputs[i].Prompt = model.GetFocusedPrompt()
			inputs[i].TextStyle = util.GetPrimaryFontStyle()
		}

		// l.accountInput = *inputs[0]
		// l.passwordInput = *inputs[1]

		if l.index == submitIndex {
			l.submitButton = model.GetFocusedSubmitButton()
		} else {
			l.submitButton = model.GetBlurredSubmitButton()
		}

		// if l.index == authIndex {
		// 	l.authButton = model.GetFocusedButton(l.authButtonTextByStep())
		// } else {
		// 	l.authButton = model.GetBlurredButton(l.authButtonTextByStep())
		// }

		return l, nil
	}

	// Handle character input and blinks
	return l.updateLoginInputs(msg)
}

func (l *LoginPage) View(a *model.App) string {
	var (
		builder  strings.Builder
		top      int // 距离顶部的行数
		mainPage = l.spotifox.MustMain()
	)

	// title
	if configs.ConfigRegistry.ShowTitle {
		builder.WriteString(mainPage.TitleView(a, &top))
	} else {
		top++
	}

	// menu title
	builder.WriteString(mainPage.MenuTitleView(a, &top, l.menuTitle))
	builder.WriteString("\n\n\n")
	top += 2

	inputs := []*textinput.Model{
		&l.accountInput,
		&l.passwordInput,
	}

	for i, input := range inputs {
		if mainPage.MenuStartColumn() > 0 {
			builder.WriteString(strings.Repeat(" ", mainPage.MenuStartColumn()))
		}

		builder.WriteString(input.View())

		var valueLen int
		if input.Value() == "" {
			valueLen = runewidth.StringWidth(input.Placeholder)
		} else {
			valueLen = runewidth.StringWidth(input.Value())
		}
		if spaceLen := l.spotifox.WindowWidth() - mainPage.MenuStartColumn() - valueLen - 3; spaceLen > 0 {
			builder.WriteString(strings.Repeat(" ", spaceLen))
		}

		top++

		if i < len(inputs)-1 {
			builder.WriteString("\n\n")
			top++
		}
	}

	builder.WriteString("\n\n")
	top++
	if mainPage.MenuStartColumn() > 0 {
		builder.WriteString(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}
	builder.WriteString(l.tips)
	builder.WriteString("\n\n")
	top++
	if mainPage.MenuStartColumn() > 0 {
		builder.WriteString(strings.Repeat(" ", mainPage.MenuStartColumn()))
	}
	builder.WriteString(l.submitButton)

	var btnBlank = "    "
	builder.WriteString(btnBlank)
	// builder.WriteString(l.authButton)

	// spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - runewidth.StringWidth(model.SubmitText) - runewidth.StringWidth(l.authButtonTextByStep()) - len(btnBlank)
	spaceLen := a.WindowWidth() - mainPage.MenuStartColumn() - runewidth.StringWidth(model.SubmitText) - len(btnBlank)
	if spaceLen > 0 {
		builder.WriteString(strings.Repeat(" ", spaceLen))
	}
	builder.WriteString("\n")

	if a.WindowHeight() > top+3 {
		builder.WriteString(strings.Repeat("\n", a.WindowHeight()-top-3))
	}

	return builder.String()
}

func (l *LoginPage) Msg() tea.Msg {
	return tickLoginMsg{}
}

func (l *LoginPage) updateLoginInputs(msg tea.Msg) (model.Page, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	l.accountInput, cmd = l.accountInput.Update(msg)
	cmds = append(cmds, cmd)

	l.passwordInput, cmd = l.passwordInput.Update(msg)
	cmds = append(cmds, cmd)

	return l, tea.Batch(cmds...)
}

// func (l *LoginPage) authButtonTextByStep() string {
// 	switch l.authStep {
// 	case 1:
// 		return "已在浏览器登录授权，继续"
// 	case 0:
// 		fallthrough
// 	default:
// 		return "OAuth授权"
// 	}
// }

func (l *LoginPage) enterHandler() (model.Page, tea.Cmd) {
	loading := model.NewLoading(l.spotifox.MustMain(), l.menuTitle)
	loading.Start()
	defer loading.Complete()

	switch l.index {
	case submitIndex:
		if len(l.accountInput.Value()) <= 0 || len(l.passwordInput.Value()) <= 0 {
			l.tips = util.SetFgStyle("请输入账号及密码", termenv.ANSIBrightRed)
			return l, nil
		}
		return l.loginByAccount()
		// case authIndex:
		//return l.loginByOAuth()
	}

	return l, tickLogin(time.Nanosecond)
}

func (l *LoginPage) loginByAccount() (model.Page, tea.Cmd) {
	login := &l.spotifox.sess.Context().Login
	login.Username = l.accountInput.Value()
	login.Password = l.passwordInput.Value()
	if err := l.spotifox.sess.Login(); err != nil {
		return l.handleLoginFail(err)
	}

	return l.handleLoginSuccess()
}

// func (l *LoginPage) loginByOAuth() (model.Page, tea.Cmd) {
//if l.authStep == 0 {
//	var authURL string
//	authURL, l.tokenChan = core.StartOAuth(constants.SpotifyClientId, constants.SpotifyClientSecret, constants.SpotifyOAuthScopes, constants.SpotifyOAuthPort)
//	_ = open.Start(authURL)
//	l.authStep++
//	return l, tickLogin(time.Nanosecond)
//}
//
//if l.tokenChan == nil {
//	utils.Logger().Print("auth failed: token chan is nil")
//	l.tips = util.SetFgStyle("Not Auth", termenv.ANSIBrightRed)
//	return l, tickLogin(time.Nanosecond)
//}
//var accessToken string
//select {
//case auth := <-l.tokenChan:
//	accessToken = auth.AccessToken
//case <-time.After(time.Second * 3):
//}
//if accessToken == "" {
//	utils.Logger().Print("auth failed: token is empty")
//	l.tips = util.SetFgStyle("Not Auth", termenv.ANSIBrightRed)
//	return l, tickLogin(time.Nanosecond)
//}
//session, err := core.LoginOAuthToken(accessToken, constants.SpotifyDeviceName)
//if err != nil {
//	utils.Logger().Printf("auth failed, get session err: %+v", err)
//	l.tips = util.SetFgStyle(err.Error(), termenv.ANSIBrightRed)
//	return l, tickLogin(time.Nanosecond)
//}
//
//l.tips = ""
//return l.loginSuccessHandle(structs.NewUserFromSession(session))
// }

func (l *LoginPage) handleLoginSuccess() (model.Page, tea.Cmd) {
	user := structs.NewUserFromSession(l.spotifox.sess.Context().Info)

	token, err := l.spotifox.sess.Mercury().GetToken(configs.ConfigRegistry.SpotifyClientId, constants.SpotifyOAuthScopes)
	if err != nil {
		return l.handleLoginFail(err)
	}
	if token == nil || token.AccessToken == "" {
		return l.handleLoginFail(errors.New("get access token failed"))
	}
	user.Token = *token

	httpClient := oauth2.NewClient(context.Background(), (*auth.TokenSourceWrapper)(&oauth2.Token{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Expiry:      time.Now().Add(time.Duration(token.ExpiresIn-30) * time.Second),
	}))
	httpClient.Timeout = constants.AppHttpTimeout
	l.spotifox.spotifyClient = spotify.New(httpClient)

	// get user profile
	u, err := l.spotifox.spotifyClient.CurrentUser(context.Background())
	if err != nil {
		return l.handleLoginFail(err)
	}
	user.User = u.User
	user.Email = u.Email
	user.Product = u.Product
	user.Birthdate = u.Birthdate

	l.spotifox.user = &user

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.User{}, user)

	// clean
	l.tips = ""
	l.accountInput.Reset()
	l.passwordInput.Reset()

	var newPage model.Page = l.spotifox.MustMain()
	if l.AfterLogin != nil {
		p := l.AfterLogin()
		if p != nil {
			newPage = p
		}
	}
	return newPage, tea.Tick(time.Nanosecond, func(t time.Time) tea.Msg { return newPage.Msg })
}

func (l *LoginPage) handleLoginFail(err error) (model.Page, tea.Cmd) {
	utils.Logger().Printf("login err, %+v", err)
	l.tips = util.SetFgStyle(err.Error(), termenv.ANSIBrightRed)
	return l, tickLogin(time.Nanosecond)
}
