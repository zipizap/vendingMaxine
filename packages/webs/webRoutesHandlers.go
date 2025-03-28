package cmd

import (
	"context"
	"net/http"
	"time"
	"vendingMaxine/packages/sharedTypes"

	"github.com/zipizap/goEchoWebOauth2Dex/webserver"

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
			<p>This is a temp placeholder - the iindex webpage will be included in SPA</p>
			</p><hr></p>
		</body>
	</html>`
	return c.HTML(http.StatusOK, html)
}

/*
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
*/

// checkAuthHandler checks if the user is authenticated by checking the presence of the id_token cookie.
// If the cookie is present, it validates the token with the OIDC provider.
// If the token is valid, it returns a JSON response {"authenticated": true}.
// If the token is invalid or missing, it returns a JSON response {"authenticated": false}.
func checkAuthHandler(c echo.Context) error {
	cookie, err := c.Cookie("id_token")
	if err != nil {
		return c.JSON(http.StatusOK, map[string]bool{"authenticated": false})
	}
	// Extract claims from the ID token
	idToken, err := webserver.OidcVerifier.Verify(context.Background(), cookie.Value)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]bool{"authenticated": false})
	}

	// Parse the claims
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		// If claims parsing fails, return authenticated without claims
		return c.JSON(http.StatusOK, map[string]bool{"authenticated": true})
	}

	// Return authentication status with claims
	return c.JSON(http.StatusOK, map[string]interface{}{
		"authenticated": true,
		"claims":        claims,
	})
}

type DateMessage struct {
	Timestamp string `json:"timestamp"`
}

// dateSocketHandlerApi upgrades the HTTP connection to a websocket and periodically sends the current date/time to the client.
// dateSocketHandlerApi will never return error - if there is a problem, it will send a json-body
// error message {"error": "error description"} and close the ws-connection.
func dateSocketHandlerApi(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		// DO NOT RETURN ERROR, as the client is already expecting a websocket
		// If upgrading the connection fails AFTER the authentication middleware,
		// it's a server error, not an authentication error.
		// Send a close message to the client.
		ws.WriteJSON(map[string]string{"error": "Server error upgrading websocket"})
		ws.Close()
		return nil
	}
	defer ws.Close()

	for {
		time.Sleep(time.Second)
		msg := DateMessage{
			Timestamp: time.Now().String(),
		}
		if err := ws.WriteJSON(msg); err != nil {
			// DO NOT RETURN ERROR, as the client is already expecting a websocket
			// If writing the date fails AFTER the websocket was established,
			// it's a server error, not an authentication error.
			// Send a close message to the client.
			ws.WriteJSON(map[string]string{"error": "Server error sending date"})
			ws.Close()
			return nil
		}
	}
}

func WebserverStart(appConfig *sharedTypes.AppConfigType) {
	E := webserver.E

	// ------------- Middlewares ------------------------
	// Configure colored logging middleware
	E.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "\033[1;34m[${time_rfc3339}]\033[0m \033[1;37m${method}\033[0m \033[1;32m${status}\033[0m ${uri} \t\t\t\t${latency}ns\n",
	}))
	E.Use(middleware.Recover())

	// ------------- Routes -----------------------------
	// Group routes for public html pages
	publicGroup4HtmlWebs := E.Group("")
	publicGroup4HtmlWebs.GET("/", indexHandler)

	// Group routes for public API endpoints
	publicApiGroup := E.Group("/api/public")
	publicApiGroup.GET("/check_auth", checkAuthHandler)

	// Group routes for private html pages (use the OauthIdTokenValidatorMiddleware)
	privateGroup4HtmlWebs := E.Group("")
	privateGroup4HtmlWebs.Use(webserver.OauthIdTokenValidatorMiddleware)

	// Group routes for private API endpoints (use the OauthIdTokenValidatorApiMiddleware)
	privateApiGroup := E.Group("/api/private")
	privateApiGroup.Use(webserver.OauthIdTokenValidatorApiMiddleware)
	privateApiGroup.GET("/ws", dateSocketHandlerApi)

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
