package cmd

import (
	"context"
	"fmt"
	"net/http"
	opshub "vendingMaxine/packages/opsHub"
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

// checkAuthHandler checks if the user is authenticated by checking the presence of the id_token cookie.
// If the cookie is present, it validates the token with the OIDC provider.
// If the token is valid, it returns a JSON response with {"authenticated": true, "claims": {<claims>}}.
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

// wsHandler upgrades the HTTP connection to a websocket
// wsHandler will never return error - if there is a problem, it will send a json-body
// error message {"error": "error description"} and close the ws-connection.
func wsHandler(c echo.Context) error {
	var err error
	// Upgrade the connection to a websocket
	var ws *websocket.Conn
	{
		ws, err = upgrader.Upgrade(c.Response(), c.Request(), nil)
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
	}

	// Get idTokenClaims from the context
	idTokenClaims := c.Get("idTokenClaims").(map[string]interface{})
	// Set currentClient from idToken claims
	var currentClient opshub.CurrentClient
	{
		claimToUseForUser := "email"
		user, ok := idTokenClaims[claimToUseForUser].(string)
		if !ok {
			// If the claimToUseForUser is not found in the claims, send an error message and close the ws-connection
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("'%s' claim not found in IdToken", claimToUseForUser)})
			ws.Close()
			return nil
		}

		claimToUseForGroups := "groups"
		groups, ok := idTokenClaims[claimToUseForGroups].([]string)
		if !ok {
			// If the claimToUseForUser is not found in the claims, send an error message and close the ws-connection
			ws.WriteJSON(map[string]string{"error": fmt.Sprintf("'%s' claim not found in IdToken", claimToUseForGroups)})
			ws.Close()
			return nil
		}
		currentClient = opshub.CurrentClient{
			User:   user,
			Groups: groups,
		}
	}

	// Infinite loop of ws messages: listen for zzzRequest, reply with zzzResponse
	for {
		CONTINUE HERE
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
	privateApiGroup.GET("/ws", wsHandler)

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
