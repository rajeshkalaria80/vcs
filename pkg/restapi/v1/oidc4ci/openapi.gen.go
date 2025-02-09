// Package oidc4ci provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package oidc4ci

import (
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
)

// Model for Access Token Response.
type AccessTokenResponse struct {
	// The access token issued by the authorization server.
	AccessToken string `json:"access_token"`

	// String containing a nonce to be used to create a proof of possession of key material when requesting a credential.
	CNonce *string `json:"c_nonce,omitempty"`

	// Integer denoting the lifetime in seconds of the c_nonce.
	CNonceExpiresIn *int `json:"c_nonce_expires_in,omitempty"`

	// The lifetime in seconds of the access token.
	ExpiresIn *int `json:"expires_in,omitempty"`

	// The refresh token, which can be used to obtain new access tokens.
	RefreshToken *string `json:"refresh_token,omitempty"`

	// OPTIONAL, if identical to the scope requested by the client; otherwise, REQUIRED.
	Scope *string `json:"scope,omitempty"`

	// The type of the token issued.
	TokenType string `json:"token_type"`
}

// Model for OIDC Credential request.
type CredentialRequest struct {
	// Format of the credential being issued.
	Format *string   `json:"format,omitempty"`
	Proof  *JWTProof `json:"proof,omitempty"`

	// Array of types of the credential being issued.
	Types []string `json:"types"`
}

// Model for OIDC Credential response.
type CredentialResponse struct {
	// A JSON string containing a token subsequently used to obtain a Credential. MUST be present when credential is not returned.
	AcceptanceToken *string `json:"acceptance_token,omitempty"`

	// JSON string containing a nonce to be used to create a proof of possession of key material when requesting a Credential.
	CNonce *string `json:"c_nonce,omitempty"`

	// JSON integer denoting the lifetime in seconds of the c_nonce.
	CNonceExpiresIn *int        `json:"c_nonce_expires_in,omitempty"`
	Credential      interface{} `json:"credential"`

	// JSON string denoting the format of the issued Credential.
	Format string `json:"format"`
}

// JWTProof defines model for JWTProof.
type JWTProof struct {
	// REQUIRED. Signed JWT as proof of key possession.
	Jwt string `json:"jwt"`

	// REQUIRED. JSON String denoting the proof type. Currently the only supported proof type is 'jwt'.
	ProofType string `json:"proof_type"`
}

// Model for Pushed Authorization Response.
type PushedAuthorizationResponse struct {
	// A JSON number that represents the lifetime of the request URI in seconds as a positive integer. The request URI lifetime is at the discretion of the authorization server but will typically be relatively short (e.g., between 5 and 600 seconds).
	ExpiresIn int `json:"expires_in"`

	// The request URI corresponding to the authorization request posted. This URI is a single-use reference to the respective request data in the subsequent authorization request.
	RequestUri string `json:"request_uri"`
}

// OAuth 2.0 client registration error response.
type RegisterOAuthClientErrorResponse struct {
	// Single ASCII error code string.
	Error string `json:"error"`

	// Human-readable ASCII text description of the error used for debugging.
	ErrorDescription *string `json:"error_description,omitempty"`
}

// OAuth 2.0 client registration request.
type RegisterOAuthClientRequest struct {
	// Human-readable string name of the client to be presented to the end-user during authorization.
	ClientName *string `json:"client_name,omitempty"`

	// URL string of a web page providing information about the client.
	ClientUri *string `json:"client_uri,omitempty"`

	// Array of strings representing ways to contact people responsible for this client, typically email addresses.
	Contacts *[]string `json:"contacts,omitempty"`

	// Array of OAuth 2.0 grant types that the client is allowed to use. Supported values: authorization_code, urn:ietf:params:oauth:grant-type:pre-authorized_code.
	GrantTypes *[]string `json:"grant_types,omitempty"`

	// Client's JSON Web Key Set document value, which contains the client's public keys.
	Jwks *map[string]interface{} `json:"jwks,omitempty"`

	// URL string referencing the client's JSON Web Key (JWK) Set document, which contains the client's public keys.
	JwksUri *string `json:"jwks_uri,omitempty"`

	// URL string that references a logo for the client.
	LogoUri *string `json:"logo_uri,omitempty"`

	// URL string that points to a human-readable privacy policy document that describes how the deployment organization collects, uses, retains, and discloses personal data.
	PolicyUri *string `json:"policy_uri,omitempty"`

	// Array of allowed redirection URI strings for the client. Required if client supports authorization_code grant type.
	RedirectUris *[]string `json:"redirect_uris,omitempty"`

	// Array of OAuth 2.0 response types that the client can use at the authorization endpoint. Supported values: code.
	ResponseTypes *[]string `json:"response_types,omitempty"`

	// String containing a space-separated list of scope values that the client can use when requesting access tokens.
	Scope *string `json:"scope,omitempty"`

	// A unique identifier string (e.g. UUID) assigned by the client developer or software publisher used by registration endpoints to identify the client software to be dynamically registered.
	SoftwareId *string `json:"software_id,omitempty"`

	// A version identifier string for the client software identified by "software_id".
	SoftwareVersion *string `json:"software_version,omitempty"`

	// Requested client authentication method for the token endpoint. Supported values: none, client_secret_post, client_secret_basic. None is used for public clients (native apps, mobile apps) which can not have secrets. Default: client_secret_basic.
	TokenEndpointAuthMethod *string `json:"token_endpoint_auth_method,omitempty"`

	// URL string that points to a human-readable terms of service document for the client that describes a contractual relationship between the end-user and the client that the end-user accepts when authorizing the client.
	TosUri *string `json:"tos_uri,omitempty"`
}

// Response with registered metadata for created OAuth 2.0 client.
type RegisterOAuthClientResponse struct {
	// Client identifier.
	ClientId string `json:"client_id"`

	// Time at which the client identifier was issued.
	ClientIdIssuedAt int `json:"client_id_issued_at"`

	// Human-readable string name of the client to be presented to the end-user during authorization.
	ClientName *string `json:"client_name,omitempty"`

	// Client secret. This value is used by the confidential client to authenticate to the token endpoint.
	ClientSecret *string `json:"client_secret,omitempty"`

	// Time at which the client secret will expire or 0 if it will not expire.
	ClientSecretExpiresAt *int `json:"client_secret_expires_at,omitempty"`

	// URL string of a web page providing information about the client.
	ClientUri *string `json:"client_uri,omitempty"`

	// Array of strings representing ways to contact people responsible for this client, typically email addresses.
	Contacts *[]string `json:"contacts,omitempty"`

	// Array of OAuth 2.0 grant types that the client is allowed to use. Supported values: authorization_code, urn:ietf:params:oauth:grant-type:pre-authorized_code.
	GrantTypes []string `json:"grant_types"`

	// Client's JSON Web Key Set document value, which contains the client's public keys.
	Jwks *map[string]interface{} `json:"jwks,omitempty"`

	// URL string referencing the client's JSON Web Key (JWK) Set document, which contains the client's public keys.
	JwksUri *string `json:"jwks_uri,omitempty"`

	// URL string that references a logo for the client.
	LogoUri *string `json:"logo_uri,omitempty"`

	// URL string that points to a human-readable privacy policy document that describes how the deployment organization collects, uses, retains, and discloses personal data.
	PolicyUri *string `json:"policy_uri,omitempty"`

	// Array of allowed redirection URI strings for the client. Required if client supports authorization_code grant type.
	RedirectUris *[]string `json:"redirect_uris,omitempty"`

	// Array of OAuth 2.0 response types that the client can use at the authorization endpoint. Supported values: code.
	ResponseTypes *[]string `json:"response_types,omitempty"`

	// String containing a space-separated list of scope values that the client can use when requesting access tokens.
	Scope *string `json:"scope,omitempty"`

	// A unique identifier string (e.g. UUID) assigned by the client developer or software publisher used by registration endpoints to identify the client software to be dynamically registered.
	SoftwareId *string `json:"software_id,omitempty"`

	// A version identifier string for the client software identified by "software_id".
	SoftwareVersion *string `json:"software_version,omitempty"`

	// Requested client authentication method for the token endpoint. Supported values: none, client_secret_post, client_secret_basic. None is used for public clients (native apps, mobile apps) which can not have secrets. Default: client_secret_basic.
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method"`

	// URL string that points to a human-readable terms of service document for the client that describes a contractual relationship between the end-user and the client that the end-user accepts when authorizing the client.
	TosUri *string `json:"tos_uri,omitempty"`
}

// OidcAuthorizeParams defines parameters for OidcAuthorize.
type OidcAuthorizeParams struct {
	// Value MUST be set to "code".
	ResponseType string `form:"response_type" json:"response_type"`

	// The client identifier.
	ClientId string `form:"client_id" json:"client_id"`

	// A challenge derived from the code verifier that is sent in the authorization request, to be verified against later.
	CodeChallenge string `form:"code_challenge" json:"code_challenge"`

	// A method that was used to derive code challenge.
	CodeChallengeMethod *string `form:"code_challenge_method,omitempty" json:"code_challenge_method,omitempty"`

	// The authorization server redirects the user-agent to the client's redirection endpoint previously established with the authorization server during the client registration process or when making the authorization request.
	RedirectUri *string `form:"redirect_uri,omitempty" json:"redirect_uri,omitempty"`

	// The scope of the access request.
	Scope *string `form:"scope,omitempty" json:"scope,omitempty"`

	// An opaque value used by the client to maintain state between the request and callback. The authorization server includes this value when redirecting the user-agent back to the client. The parameter SHOULD be used for preventing cross-site request forgery.
	State *string `form:"state,omitempty" json:"state,omitempty"`

	// The authorization_details conveys the details about the credentials the wallet wants to obtain. Multiple authorization_details can be used with type openid_credential to request authorization in case of multiple credentials.
	AuthorizationDetails *string `form:"authorization_details,omitempty" json:"authorization_details,omitempty"`

	// Wallet's OpenID Connect Issuer URL. The Issuer will use the discovery process to determine the wallet's capabilities and endpoints. RECOMMENDED in Dynamic Credential Request.
	WalletIssuer *string `form:"wallet_issuer,omitempty" json:"wallet_issuer,omitempty"`

	// An opaque user hint the wallet MAY use in subsequent callbacks to optimize the user's experience. RECOMMENDED in Dynamic Credential Request.
	UserHint *string `form:"user_hint,omitempty" json:"user_hint,omitempty"`

	// String value identifying a certain processing context at the credential issuer. A value for this parameter is typically passed in an issuance initiation request from the issuer to the wallet. This request parameter is used to pass the  issuer_state value back to the credential issuer. The issuer must take into account that op_state is not guaranteed to originate from this issuer, could be an attack.
	IssuerState *string `form:"issuer_state,omitempty" json:"issuer_state,omitempty"`

	// String indicating that client is using an identifier not assigned by the authorization server. The only supported value "urn:ietf:params:oauth:client-id-scheme:oauth-discoverable-client" specifies "client_id" parameter in the request as an HTTPS based URL corresponding to the "client_uri". If the authorization server does not already have the metadata for the identified client, it can retrieve the metadata from client’s well-known location.
	ClientIdScheme *string `form:"client_id_scheme,omitempty" json:"client_id_scheme,omitempty"`
}

// OidcCredentialJSONBody defines parameters for OidcCredential.
type OidcCredentialJSONBody = CredentialRequest

// OidcRedirectParams defines parameters for OidcRedirect.
type OidcRedirectParams struct {
	// auth code for issuer provider
	Code string `form:"code" json:"code"`

	// state
	State string `form:"state" json:"state"`
}

// OidcRegisterClientJSONBody defines parameters for OidcRegisterClient.
type OidcRegisterClientJSONBody = RegisterOAuthClientRequest

// OidcCredentialJSONRequestBody defines body for OidcCredential for application/json ContentType.
type OidcCredentialJSONRequestBody = OidcCredentialJSONBody

// OidcRegisterClientJSONRequestBody defines body for OidcRegisterClient for application/json ContentType.
type OidcRegisterClientJSONRequestBody = OidcRegisterClientJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// OIDC Authorization Request
	// (GET /oidc/authorize)
	OidcAuthorize(ctx echo.Context, params OidcAuthorizeParams) error
	// OIDC Credential
	// (POST /oidc/credential)
	OidcCredential(ctx echo.Context) error
	// OIDC Pushed Authorization Request
	// (POST /oidc/par)
	OidcPushedAuthorizationRequest(ctx echo.Context) error
	// OIDC Redirect
	// (GET /oidc/redirect)
	OidcRedirect(ctx echo.Context, params OidcRedirectParams) error
	// OIDC Token Request
	// (POST /oidc/token)
	OidcToken(ctx echo.Context) error
	// OIDC Register OAuth Client
	// (POST /oidc/{profileID}/{profileVersion}/register)
	OidcRegisterClient(ctx echo.Context, profileID string, profileVersion string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// OidcAuthorize converts echo context to params.
func (w *ServerInterfaceWrapper) OidcAuthorize(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params OidcAuthorizeParams
	// ------------- Required query parameter "response_type" -------------

	err = runtime.BindQueryParameter("form", true, true, "response_type", ctx.QueryParams(), &params.ResponseType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter response_type: %s", err))
	}

	// ------------- Required query parameter "client_id" -------------

	err = runtime.BindQueryParameter("form", true, true, "client_id", ctx.QueryParams(), &params.ClientId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter client_id: %s", err))
	}

	// ------------- Required query parameter "code_challenge" -------------

	err = runtime.BindQueryParameter("form", true, true, "code_challenge", ctx.QueryParams(), &params.CodeChallenge)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code_challenge: %s", err))
	}

	// ------------- Optional query parameter "code_challenge_method" -------------

	err = runtime.BindQueryParameter("form", true, false, "code_challenge_method", ctx.QueryParams(), &params.CodeChallengeMethod)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code_challenge_method: %s", err))
	}

	// ------------- Optional query parameter "redirect_uri" -------------

	err = runtime.BindQueryParameter("form", true, false, "redirect_uri", ctx.QueryParams(), &params.RedirectUri)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter redirect_uri: %s", err))
	}

	// ------------- Optional query parameter "scope" -------------

	err = runtime.BindQueryParameter("form", true, false, "scope", ctx.QueryParams(), &params.Scope)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter scope: %s", err))
	}

	// ------------- Optional query parameter "state" -------------

	err = runtime.BindQueryParameter("form", true, false, "state", ctx.QueryParams(), &params.State)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter state: %s", err))
	}

	// ------------- Optional query parameter "authorization_details" -------------

	err = runtime.BindQueryParameter("form", true, false, "authorization_details", ctx.QueryParams(), &params.AuthorizationDetails)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter authorization_details: %s", err))
	}

	// ------------- Optional query parameter "wallet_issuer" -------------

	err = runtime.BindQueryParameter("form", true, false, "wallet_issuer", ctx.QueryParams(), &params.WalletIssuer)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter wallet_issuer: %s", err))
	}

	// ------------- Optional query parameter "user_hint" -------------

	err = runtime.BindQueryParameter("form", true, false, "user_hint", ctx.QueryParams(), &params.UserHint)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter user_hint: %s", err))
	}

	// ------------- Optional query parameter "issuer_state" -------------

	err = runtime.BindQueryParameter("form", true, false, "issuer_state", ctx.QueryParams(), &params.IssuerState)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter issuer_state: %s", err))
	}

	// ------------- Optional query parameter "client_id_scheme" -------------

	err = runtime.BindQueryParameter("form", true, false, "client_id_scheme", ctx.QueryParams(), &params.ClientIdScheme)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter client_id_scheme: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.OidcAuthorize(ctx, params)
	return err
}

// OidcCredential converts echo context to params.
func (w *ServerInterfaceWrapper) OidcCredential(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.OidcCredential(ctx)
	return err
}

// OidcPushedAuthorizationRequest converts echo context to params.
func (w *ServerInterfaceWrapper) OidcPushedAuthorizationRequest(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.OidcPushedAuthorizationRequest(ctx)
	return err
}

// OidcRedirect converts echo context to params.
func (w *ServerInterfaceWrapper) OidcRedirect(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params OidcRedirectParams
	// ------------- Required query parameter "code" -------------

	err = runtime.BindQueryParameter("form", true, true, "code", ctx.QueryParams(), &params.Code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter code: %s", err))
	}

	// ------------- Required query parameter "state" -------------

	err = runtime.BindQueryParameter("form", true, true, "state", ctx.QueryParams(), &params.State)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter state: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.OidcRedirect(ctx, params)
	return err
}

// OidcToken converts echo context to params.
func (w *ServerInterfaceWrapper) OidcToken(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.OidcToken(ctx)
	return err
}

// OidcRegisterClient converts echo context to params.
func (w *ServerInterfaceWrapper) OidcRegisterClient(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "profileID" -------------
	var profileID string

	err = runtime.BindStyledParameterWithLocation("simple", false, "profileID", runtime.ParamLocationPath, ctx.Param("profileID"), &profileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter profileID: %s", err))
	}

	// ------------- Path parameter "profileVersion" -------------
	var profileVersion string

	err = runtime.BindStyledParameterWithLocation("simple", false, "profileVersion", runtime.ParamLocationPath, ctx.Param("profileVersion"), &profileVersion)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter profileVersion: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.OidcRegisterClient(ctx, profileID, profileVersion)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/oidc/authorize", wrapper.OidcAuthorize)
	router.POST(baseURL+"/oidc/credential", wrapper.OidcCredential)
	router.POST(baseURL+"/oidc/par", wrapper.OidcPushedAuthorizationRequest)
	router.GET(baseURL+"/oidc/redirect", wrapper.OidcRedirect)
	router.POST(baseURL+"/oidc/token", wrapper.OidcToken)
	router.POST(baseURL+"/oidc/:profileID/:profileVersion/register", wrapper.OidcRegisterClient)

}
