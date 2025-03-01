// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.833
package game

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"fmt"
	"github.com/Megidy/k/static/templates/components"
)

// import "github.com/Megidy/k/static/components"
func Game(roomID string, isAlreadyPlaying bool, isFound bool, isOwner bool, isSpectator bool) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<head><title>Room</title><script src=\"https://unpkg.com/htmx.org@1.9.8\"></script><script src=\"https://unpkg.com/htmx.org/dist/ext/multi-swap.js\"></script><script src=\"https://unpkg.com/htmx.org/dist/ext/ws.js\"></script><meta name=\"htmx-config\" content=\"{&#34;wsReconnectDelay&#34;:&#34;10&#34;}\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><link rel=\"stylesheet\" href=\"/static/css/game/game.css\"><link rel=\"stylesheet\" href=\"/static/css/components/defaultQuestion.css\"><link rel=\"stylesheet\" href=\"/static/css/game/timeLoader.css\"></head><body>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = components.TopNavBar().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<div class=\"gameWrapper\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !isFound {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "<div>Кімната до якої ви намагаєтесь приєднатись , завершила своє існування або її початково не існувало:((</div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			if isOwner {
				if isSpectator {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "<div class=\"gameClass\" hx-ext=\"ws\" ws-connect=\"")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var2 string
					templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("/ws/room/handler/%s", roomID))
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `static/templates/game/game.templ`, Line: 29, Col: 96}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "\" ws-reconnect-delay=\"10000\"><div><h4>Room ID : ")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var3 string
					templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(roomID)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `static/templates/game/game.templ`, Line: 30, Col: 34}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "</h4></div><div id=\"leaderboard\"><div id=\"time\"><div id=\"spectator\"><p>Ви спостерігаєте, тож ви не можете відповідати , спостерігайте за поточними результатами</p><button type=\"submit\" id=\"beforeGameForceStart\" ws-send>примусовий початок</button><div id=\"beforeGameWait\"></div></div></div><div id=\"innerLeaderboard\"></div><div id=\"currQuestion\"></div><div id=\"currNotReadyPlayers\"></div></div></div>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					if isAlreadyPlaying {
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "<div>Якщо ви намагаєтесь перезайти , будь ласка перезагрузіть сторінку ще раз, якщо ж це не допомогло,це означає , що ви вже маєте відкриту вкладу з цією кімнатою</div>")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
					} else {
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, "<div class=\"gameClass\" hx-ext=\"ws\" ws-connect=\"")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						var templ_7745c5c3_Var4 string
						templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("/ws/room/handler/%s", roomID))
						if templ_7745c5c3_Err != nil {
							return templ.Error{Err: templ_7745c5c3_Err, FileName: `static/templates/game/game.templ`, Line: 48, Col: 97}
						}
						_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "\" ws-reconnect-delay=\"10000\"><div><h4>Room ID : ")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						var templ_7745c5c3_Var5 string
						templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(roomID)
						if templ_7745c5c3_Err != nil {
							return templ.Error{Err: templ_7745c5c3_Err, FileName: `static/templates/game/game.templ`, Line: 49, Col: 35}
						}
						_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
						templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "</h4></div><div id=\"leaderboard\"><div id=\"time\"></div><div id=\"game\"><p>Почекайте на інших користувачів</p><button type=\"submit\" ws-send>примусовий початок</button><div id=\"beforeGameWait\"></div></div></div></div>")
						if templ_7745c5c3_Err != nil {
							return templ_7745c5c3_Err
						}
					}
				}
			} else {
				if isAlreadyPlaying {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 11, "<div>Якщо ви намагаєтесь перезайти , будь ласка перезагрузіть сторінку ще раз, якщо ж це не допомогло,це означає , що ви вже маєте відкриту вкладу з цією кімнатою</div>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				} else {
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 12, "<div class=\"gameClass\" hx-ext=\"ws\" ws-connect=\"")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var6 string
					templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("/ws/room/handler/%s", roomID))
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `static/templates/game/game.templ`, Line: 65, Col: 96}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 13, "\" ws-reconnect-delay=\"10000\"><div><h4>Room ID : ")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var7 string
					templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(roomID)
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `static/templates/game/game.templ`, Line: 66, Col: 34}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 14, "</h4></div><div id=\"leaderboard\"><div id=\"time\"></div><div id=\"game\"><p>Почекайте на інших користувачів</p><div id=\"beforeGameWait\"></div></div></div></div>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
			}
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 15, "</div></body>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
