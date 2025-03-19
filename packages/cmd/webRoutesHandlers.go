package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"vendingMaxine/packages/webserver"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ------------- Handlers ------------------------

// / indexHandler serves the index page with links to other pages.
func indexHandler(c echo.Context) error {
	html := `<html>
		<body>
			<h1>Index</h1>
			<ul>
				<li><a href="/public">Public</a></li>
				<li><a href="/login">Login</a></li>
				<li><a href="/logout">Logout</a></li>
				<li><a href="/private">Private</a></li>
				<li><a href="/date_from_server">Server Date</a></li>
			</ul>
			</p><hr></p>
		</body>
	</html>`
	return c.HTML(http.StatusOK, html)
}

// privateHandler serves /private the private page, which is only accessible
// if the user is already authenticated (contains valid ID token as cookie).
func privateHandler(c echo.Context) error {

	// The cookie-and-claims code bellow is not necessary, as the authMiddleware already took care of verifying the ID token before reaching this handler.
	// This code is left here as an example of how to extract claims from the ID token.
	cookie, err := c.Cookie("id_token")
	if err != nil {
		return c.String(http.StatusUnauthorized, "Missing ID token")
	}

	idToken, err := webserver.OidcVerifier.Verify(context.Background(), cookie.Value)
	if err != nil {
		return c.String(http.StatusUnauthorized, "Invalid ID token")
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to parse claims: %v", err))
	}

	return c.HTML(http.StatusOK, fmt.Sprintf(`<html>
		<body>
			<h1>Private</h1>
			<p>Private access granted</p>
			<p>ID_Token: %s</p>
			<p>Claims: %v</p>
			<ul>
				<li><a href="/public">Public</a></li>
				<li><a href="/login">Login</a></li>
				<li><a href="/logout">Logout</a></li>
				<li><a href="/private">Private</a></li>
				<li><a href="/date_from_server">Server Date</a></li>
			</ul>
			</p><hr></p>
		</body>
	</html>`, cookie.Value, claims))
}

// dateFromServerHandler serves an HTML page with JavaScript that connects to the /ws_date websocket.
func dateFromServerHandler(c echo.Context) error {
	return c.HTML(http.StatusOK, `<html>
		<body>
			<ul>
				<li><a href="/public">Public</a></li>
				<li><a href="/login">Login</a></li>
				<li><a href="/logout">Logout</a></li>
				<li><a href="/private">Private</a></li>
				<li><a href="/date_from_server">Server Date</a></li>
			</ul>
			<h1>Server Date</h1>
			<div id="serverDate">Connecting...</div>
			<script>
				const ws = new WebSocket("ws://" + location.host + "/ws_date");
				ws.onmessage = (e) => {
					const data = JSON.parse(e.data);
					document.getElementById("serverDate").innerText += "\n" + data.timestamp;
				};
			</script>
		</body>
	</html>`)
}

type DateMessage struct {
	Timestamp string `json:"timestamp"`
}

// dateSocketHandler upgrades the HTTP connection to a websocket and periodically sends the current date/time to the client.
func dateSocketHandler(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		time.Sleep(time.Second)
		msg := DateMessage{
			Timestamp: time.Now().String(),
		}
		if err := ws.WriteJSON(msg); err != nil {
			return err
		}
	}
}

func webserverStart() {
	E := webserver.E

	// ------------- Middlewares ------------------------
	// Configure colored logging middleware
	E.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "\033[1;34m[${time_rfc3339}]\033[0m \033[1;37m${method}\033[0m \033[1;32m${status}\033[0m ${uri} \t\t\t\t${latency}ns\n",
	}))
	E.Use(middleware.Recover())

	// ------------- Routes -----------------------------
	// Add routes
	E.GET("/", indexHandler)
	// Public route: accessible by anyone.

	E.GET("/public", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<html>
			<body>
				<h1>Public Page</h1>
				<p>This is a public page.</p>
				<ul>
					<li><a href="/public">Public</a></li>
					<li><a href="/login">Login</a></li>
					<li><a href="/logout">Logout</a></li>
					<li><a href="/private">Private</a></li>
					<li><a href="/date_from_server">Server Date</a></li>
				</ul>
				</p><hr></p>
			</body>
		</html>`)
	})

	// Group routes that use the OauthIdTokenValidatorMiddleware
	privateGroup := E.Group("")
	privateGroup.Use(webserver.OauthIdTokenValidatorMiddleware)
	privateGroup.GET("/private", privateHandler)
	privateGroup.GET("/date_from_server", dateFromServerHandler)
	privateGroup.GET("/ws_date", dateSocketHandler)

	log.Info().Msg("Starting web server")
	webserverConfigOauth := &webserver.ConfigOauthClientDex{
		ClientID:          appConfig.DexConfig.ClientId,
		ClientSecret:      appConfig.DexConfig.ClientSecret,
		ClientRedirectURL: appConfig.DexConfig.ClientRedirectURL,
		ClientScopes:      []string{"openid", "profile", "email", "groups"},
		DexIssuer:         appConfig.DexConfig.DexIssuer,
	}
	webserver.Start(webserverConfigOauth)
}
