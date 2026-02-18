package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	goLog "log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kiwiirc/plugin-fileuploader/events"
	"github.com/kiwiirc/plugin-fileuploader/logging"
	"github.com/kiwiirc/plugin-fileuploader/shardedfilestore"
	tusd "github.com/tus/tusd/pkg/handler"
)

func customizedCors(serv *UploadServer) gin.HandlerFunc {
	// convert slice values to keys of map for "contains" test
	originSet := make(map[string]struct{}, len(serv.cfg.Server.CorsOrigins))
	allowAll := false
	exists := struct{}{}
	for _, origin := range serv.cfg.Server.CorsOrigins {
		if origin == "*" {
			allowAll = true
			continue
		}
		originSet[origin] = exists
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		respHeader := c.Writer.Header()

		// only allow the origin if it's in the list from the config
		if allowAll && origin != "" {
			respHeader.Set("Access-Control-Allow-Origin", origin)
		} else if _, ok := originSet[origin]; ok {
			respHeader.Set("Access-Control-Allow-Origin", origin)
		} else {
			respHeader.Del("Access-Control-Allow-Origin")
			if c.Request.Method != "HEAD" && c.Request.Method != "GET" {
				// Don't log unknown cors origin for HEAD or GET requests
				serv.log.Warn().Str("origin", origin).Msg("Unknown cors origin")
			}
		}

		// lets the user-agent know the response can vary depending on the origin of the request.
		// ensures correct behaviour of browser cache.
		respHeader.Add("Vary", "Origin")
	}
}

func (serv *UploadServer) fileuploaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" && c.Request.Method != "DELETE" {
			// Metadata is only required for POST and DELETE requests
			return
		}

		metadata := tusd.ParseMetadataHeader(c.Request.Header.Get("Upload-Metadata"))

		// ensure the user does not try to provide their own RemoteIP
		delete(metadata, "RemoteIP")

		// determine the originating IP
		remoteIP, err := serv.getDirectOrForwardedRemoteIP(c.Request)
		if err != nil {
			if addrErr, ok := err.(*net.AddrError); ok {
				c.AbortWithError(http.StatusInternalServerError, addrErr).SetType(gin.ErrorTypePrivate)
			} else {
				c.AbortWithError(http.StatusNotAcceptable, err)
			}
			return
		}

		// add RemoteIP to metadata
		metadata["RemoteIP"] = remoteIP

		err = serv.processJwt(metadata)
		if err != nil {
			// Jwt failures are none fatal, but will result in the uploaded being treated as anonymous
			// Stick a warning in the log to help with debugging
			serv.log.Warn().
				Err(err).
				Str("extjwt", metadata["extjwt"]).
				Msg("Failed to process EXTJWT")
			err = nil
		}

		// extjwt is no longer needed, remove so it does not get stored with the file info
		delete(metadata, "extjwt")

		// Update metadata with any changes that have been made
		c.Request.Header.Set("Upload-Metadata", tusd.SerializeMetadataHeader(metadata))

		// store metadata in gin context so it does not need to be parsed again
		c.Set("metadata", metadata)
	}
}

func (serv *UploadServer) registerTusHandlers(r *gin.Engine, store *shardedfilestore.ShardedFileStore) error {
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	maximumUploadSize := serv.cfg.Storage.MaximumUploadSize
	serv.log.Debug().Str("size", maximumUploadSize.String()).Msg("Using upload limit")

	config := tusd.Config{
		BasePath:                serv.cfg.Server.BasePath,
		StoreComposer:           composer,
		MaxSize:                 int64(maximumUploadSize.Bytes()),
		Logger:                  goLog.New(ioutil.Discard, "", 0),
		NotifyCompleteUploads:   true,
		NotifyCreatedUploads:    true,
		NotifyTerminatedUploads: true,
		NotifyUploadProgress:    true,
	}

	routePrefix, err := routePrefixFromBasePath(serv.cfg.Server.BasePath)
	if err != nil {
		return err
	}

	// Serve favicon.ico directly so browsers don't get a 404 for it
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/x-icon", faviconIcoBytes)
	})

	// iOS/Safari requests these at the root regardless of <link> tags in HTML
	appleIconHandler := func(c *gin.Context) {
		c.Data(http.StatusOK, "image/png", appleTouchIconBytes)
	}
	r.GET("/apple-touch-icon.png", appleIconHandler)
	r.GET("/apple-touch-icon-precomposed.png", appleIconHandler)

	handler, err := tusd.NewUnroutedHandler(config)
	if err != nil {
		return err
	}

	// create event broadcaster
	serv.tusEventBroadcaster = events.NewTusEventBroadcaster(handler)

	// attach logger
	go logging.TusdLogger(serv.log, serv.tusEventBroadcaster)

	noopHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	tusdMiddleware := gin.WrapH(handler.Middleware(noopHandler))

	rg := r.Group(routePrefix)
	rg.Use(tusdMiddleware)
	rg.Use(customizedCors(serv))
	rg.Use(serv.fileuploaderMiddleware())
	// Reject IDs shorter than the shard layer count to prevent panics in ShardedFileStore
	minIDLen := serv.cfg.Storage.ShardLayers
	rg.Use(func(c *gin.Context) {
		if id := c.Param("id"); id != "" && len(id) < minIDLen {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	})
	rg.POST("", serv.postFile(handler))

	// Register a dummy handler for OPTIONS, without this the middleware's would not be called
	rg.OPTIONS("*any", gin.WrapH(noopHandler))

	headFile := gin.WrapF(handler.HeadFile)
	rg.HEAD(":id", headFile)
	rg.HEAD(":id/:filename", rewritePath(headFile, routePrefix))

	// Use new content negotiation handler for GET requests
	rg.GET(":id", serv.getFileOrHtml(handler))
	rg.GET(":id/:filename", serv.getFileOrHtml(handler))

	patchFile := gin.WrapF(handler.PatchFile)
	rg.PATCH(":id", patchFile)
	rg.PATCH(":id/:filename", rewritePath(patchFile, routePrefix))

	// Only attach the DELETE handler if the Terminate() method is provided
	if config.StoreComposer.UsesTerminater {
		delFile := serv.delFile(handler)
		rg.DELETE(":id", delFile)
		rg.DELETE(":id/:filename", rewritePath(delFile, routePrefix))
	}

	return nil
}

func (serv *UploadServer) postFile(handler *tusd.UnroutedHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		metadata := c.MustGet("metadata").(map[string]string)

		if serv.cfg.Server.RequireJwtAccount {
			if metadata["account"] == "" {
				c.Error(errors.New("Missing JWT account")).SetType(gin.ErrorTypePublic)
				c.AbortWithStatusJSON(http.StatusUnauthorized, "Account required")
				return
			}
		}

		handler.PostFile(c.Writer, c.Request)
	}
}

func (serv *UploadServer) delFile(handler *tusd.UnroutedHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		metadata := c.MustGet("metadata").(map[string]string)

		var uploaderIP, jwtAccount, jwtIssuer string
		row := serv.DBConn.DB.QueryRow(`SELECT uploader_ip, jwt_account, jwt_issuer FROM uploads WHERE id = ?`, id)
		err := row.Scan(&uploaderIP, &jwtAccount, &jwtIssuer)

		// no finalized upload exists
		if err == sql.ErrNoRows {
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err).SetType(gin.ErrorTypePrivate)
			return
		}

		if jwtAccount != "" && (jwtAccount != metadata["account"] || jwtIssuer != metadata["issuer"]) {
			// The upload was created by an identified account that does not match this requests account
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		} else if uploaderIP == "" || uploaderIP != metadata["RemoteIP"] {
			// The upload was created by an anonymous user that does not match this requests ip address
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		handler.DelFile(c.Writer, c.Request)
	}
}

// getFileOrHtml handles GET requests with content negotiation
// Browsers receive HTML wrapper, tools (curl/wget) receive binary
// Query parameter ?raw=1 forces binary download
func (serv *UploadServer) getFileOrHtml(handler *tusd.UnroutedHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		filename := c.Param("filename")

		// serveRaw rewrites the URL to strip any filename before forwarding to the
		// TUS handler. tusd extracts the file ID from the last URL path segment, so
		// leaving the filename in place causes it to use the filename as the ID,
		// which can be too short for the ShardedFileStore and trigger a panic.
		routePrefix, _ := routePrefixFromBasePath(serv.cfg.Server.BasePath)
		serveRaw := func() {
			c.Request.URL.Path = path.Join(routePrefix, id)
			handler.GetFile(c.Writer, c.Request)
		}

		// If ?raw=1 query parameter is present, always serve binary
		if c.Query("raw") == "1" {
			serveRaw()
			return
		}

		// Load file metadata from storage
		upload, err := serv.store.GetUpload(c.Request.Context(), id)
		if err != nil {
			// File not found or error loading — serve HTML 404 for browsers
			accept := c.GetHeader("Accept")
			userAgent := c.GetHeader("User-Agent")
			wantsHtml := strings.Contains(accept, "text/html")
			isBrowser := strings.Contains(userAgent, "Mozilla") ||
				strings.Contains(userAgent, "Chrome") ||
				strings.Contains(userAgent, "Safari")
			if wantsHtml || (isBrowser && accept == "*/*") {
				serv.serveHtml404(c)
				return
			}
			serveRaw()
			return
		}

		info, err := upload.GetInfo(c.Request.Context())
		if err != nil {
			// Error getting file info, fall back to binary
			serveRaw()
			return
		}

		// Get MIME type from metadata
		mimeType := info.MetaData["filetype"]
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		// Optional: Validate filename matches if provided
		if filename != "" {
			actualFilename := info.MetaData["filename"]
			decodedFilename, _ := url.PathUnescape(filename)
			if actualFilename != "" && decodedFilename != actualFilename {
				// Redirect to canonical URL with correct filename
				routePrefix, _ := routePrefixFromBasePath(serv.cfg.Server.BasePath)
				correctPath := path.Join(routePrefix, id, url.PathEscape(actualFilename))
				// Preserve ?raw=1 query parameter in redirect if present
				if c.Query("raw") != "" {
					correctPath += "?raw=1"
				}
				c.Redirect(http.StatusMovedPermanently, correctPath)
				return
			}
		}

		// Determine if file is viewable
		isViewable := strings.HasPrefix(mimeType, "image/") ||
			strings.HasPrefix(mimeType, "video/") ||
			strings.HasPrefix(mimeType, "audio/") ||
			strings.HasPrefix(mimeType, "text/") ||
			mimeType == "application/pdf"

		if !isViewable {
			// Always serve as binary for non-viewable files
			serveRaw()
			return
		}

		// Detect client type via Accept header
		accept := c.GetHeader("Accept")
		userAgent := c.GetHeader("User-Agent")

		wantsHtml := strings.Contains(accept, "text/html")
		isBrowser := strings.Contains(userAgent, "Mozilla") ||
			strings.Contains(userAgent, "Chrome") ||
			strings.Contains(userAgent, "Safari")

		if wantsHtml || (isBrowser && accept == "*/*") {
			serv.serveHtmlWrapper(c, handler, upload, info)
		} else {
			// Serve raw file via TUS handler
			serveRaw()
		}
	}
}

// serveHtml404 renders the 404 error page for browsers
func (serv *UploadServer) serveHtml404(c *gin.Context) {
	t := detectLanguage(c.GetHeader("Accept-Language"))
	view := NotFoundView{
		MaxAge:           humanizeDuration(serv.cfg.Expiration.MaxAge.Duration, t),
		IdentifiedMaxAge: humanizeDuration(serv.cfg.Expiration.IdentifiedMaxAge.Duration, t),
		T:                t,
	}
	c.Status(http.StatusNotFound)
	if err := serv.notFoundTemplate.Execute(c.Writer, view); err != nil {
		serv.log.Error().Err(err).Msg("Failed to render 404 template")
	}
}

// serveHtmlWrapper renders the HTML preview page for a file
func (serv *UploadServer) serveHtmlWrapper(c *gin.Context, handler *tusd.UnroutedHandler, upload tusd.Upload, info tusd.FileInfo) {
	mimeType := info.MetaData["filetype"]
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	filename := info.MetaData["filename"]
	if filename == "" {
		filename = info.ID
	}

	// Build absolute URLs for Open Graph (crawlers need full URLs)
	// Detect scheme (check X-Forwarded-Proto header for reverse proxy)
	scheme := "https"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if c.Request.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	// Build absolute URLs using current request path to preserve URL encoding
	// PageURL: HTML page URL for Open Graph (no query params)
	// DirectURL: Binary file URL with ?raw=1 for img/video/audio tags
	host := c.Request.Host
	pageURL := scheme + "://" + host + c.Request.URL.Path
	directURL := scheme + "://" + host + c.Request.URL.Path + "?raw=1"

	// Build view model
	view := FileView{
		Filename:      filename,
		FileSize:      info.Size,
		FileSizeHuman: humanizeBytes(info.Size),
		MimeType:      mimeType,
		PageURL:       pageURL,
		DirectURL:     directURL,
		IsImage:       strings.HasPrefix(mimeType, "image/"),
		IsVideo:       strings.HasPrefix(mimeType, "video/"),
		IsAudio:       strings.HasPrefix(mimeType, "audio/"),
		IsPDF:         mimeType == "application/pdf",
		IsText:        strings.HasPrefix(mimeType, "text/"),
		T:             detectLanguage(c.GetHeader("Accept-Language")),
	}

	// Parse expiration from metadata
	if expiresStr, ok := info.MetaData["expires"]; ok {
		expiresUnix, err := strconv.ParseInt(expiresStr, 10, 64)
		if err == nil {
			view.ExpiresAt = time.Unix(expiresUnix, 0)
		}
	}

	// For text files, load content
	if view.IsText {
		textContent, err := serv.readTextContent(upload, info.Size)
		if err != nil {
			// If error (file too large, non-UTF8), fallback to binary download
			serv.log.Debug().
				Err(err).
				Str("id", info.ID).
				Msg("Failed to read text content, serving as binary")
			routePrefix, _ := routePrefixFromBasePath(serv.cfg.Server.BasePath)
			c.Request.URL.Path = path.Join(routePrefix, info.ID)
			handler.GetFile(c.Writer, c.Request)
			return
		}
		view.TextContent = textContent
	}

	// Set headers (CSP is defined in HTML template meta tag)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)

	// Execute template
	err := serv.htmlTemplate.Execute(c.Writer, view)
	if err != nil {
		serv.log.Error().Err(err).Msg("Failed to render HTML template")
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// readTextContent reads and validates UTF-8 text file content
func (serv *UploadServer) readTextContent(upload tusd.Upload, fileSize int64) (string, error) {
	// Limit size to avoid loading huge files into memory
	const maxTextSize = 1024 * 1024 // 1 MB max for text display
	if fileSize > maxTextSize {
		return "", fmt.Errorf("file too large for text display (%d bytes)", fileSize)
	}

	// Get file reader
	reader, err := upload.GetReader(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get reader: %w", err)
	}

	// Close reader if it implements io.Closer
	if closer, ok := reader.(io.Closer); ok {
		defer closer.Close()
	}

	// Read complete content
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}

	// Validate UTF-8
	if !utf8.Valid(content) {
		return "", fmt.Errorf("file is not valid UTF-8")
	}

	return string(content), nil
}

func (serv *UploadServer) getSecretForToken(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Failed to get claims")
	}

	issuer, ok := claims["iss"]
	if !ok {
		return nil, errors.New("Issuer field 'iss' missing from JWT")
	}

	issuerStr, ok := issuer.(string)
	if !ok {
		return nil, errors.New("Failed to coerce issuer to string")
	}

	secret, ok := serv.cfg.JwtSecretsByIssuer[issuerStr]
	if !ok {
		// Attempt to get fallback issuer
		secret, ok = serv.cfg.JwtSecretsByIssuer["*"]

		if !ok {
			return nil, fmt.Errorf("Issuer %#v not configured", issuerStr)
		} else {
			serv.log.Warn().
				Msg(fmt.Sprintf("Issuer %#v not configured, used fallback", issuerStr))
		}
	}

	return []byte(secret), nil
}

func (serv *UploadServer) processJwt(metadata map[string]string) (err error) {
	// ensure the client doesn't attempt to specify their own account/issuer fields
	delete(metadata, "issuer")
	delete(metadata, "account")

	tokenString := metadata["extjwt"]
	if tokenString == "" {
		return nil
	}

	token, err := jwt.Parse(tokenString, serv.getSecretForToken)
	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("invalid jwt")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("no jwt claims")
	}

	if issuer, ok := claims["iss"].(string); ok {
		metadata["issuer"] = issuer
	}
	if account, ok := claims["account"].(string); ok {
		metadata["account"] = account
	}

	return nil
}

// ErrInvalidXForwardedFor occurs if the X-Forwarded-For header is trusted but invalid
var ErrInvalidXForwardedFor = errors.New("Failed to parse IP from X-Forwarded-For header")

func (serv *UploadServer) getDirectOrForwardedRemoteIP(req *http.Request) (string, error) {
	// extract direct IP
	remoteIP, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		serv.log.Error().
			Err(err).
			Msg("Could not split address into host and port")
		return "", err
	}

	// use X-Forwarded-For header if direct IP is a trusted reverse proxy
	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if serv.remoteIPisTrusted(net.ParseIP(remoteIP)) {
			// We do not check intermediary proxies against the whitelist.
			// If a trusted proxy is appending to and forwarding the value of the
			// header it is receiving, that is an implicit expression of trust
			// which we will honor transitively.

			// take the first comma delimited address
			// this is the original client address
			parts := strings.Split(forwardedFor, ",")
			forwardedForClient := strings.TrimSpace(parts[0])
			forwardedForIP := net.ParseIP(forwardedForClient)
			if forwardedForIP == nil {
				err := ErrInvalidXForwardedFor
				serv.log.Error().
					Err(err).
					Str("client", forwardedForClient).
					Str("remoteIP", remoteIP).
					Msg("Couldn't use trusted X-Forwarded-For header")
				return "", err
			}
			return forwardedForIP.String(), nil
		}
		serv.log.Warn().
			Str("X-Forwarded-For", forwardedFor).
			Str("remoteIP", remoteIP).
			Msg("Untrusted remote attempted to override stored IP")
	}

	// otherwise use direct IP
	return remoteIP, nil
}

func (serv *UploadServer) remoteIPisTrusted(remoteIP net.IP) bool {
	// check if remote IP is a trusted reverse proxy
	for _, trustedNet := range serv.cfg.Server.TrustedReverseProxyRanges {
		if trustedNet.Contains(remoteIP) {
			return true
		}
	}
	return false
}

func routePrefixFromBasePath(basePath string) (string, error) {
	url, err := url.Parse(basePath)
	if err != nil {
		return "", err
	}

	return url.Path, nil
}

func rewritePath(handler gin.HandlerFunc, routePrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// rewrite request path to ":id" route pattern
		c.Request.URL.Path = path.Join(routePrefix, url.PathEscape(c.Param("id")))

		// call the normal handler
		handler(c)
	}
}
