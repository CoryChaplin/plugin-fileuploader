package config

const defaultConfig = `
[Server]
# Note: ListenAddress is not used when running as a webircgateway plugin.  In
# those situations, the listen addresses of the webircgateway will be used,
# including HTTPS if configured.
ListenAddress = "127.0.0.1:8088"

# When running as a webircgateway plugin, this path will be relative to the
# webircgateway domain, e.g. https://ws.irc.example.com/files
BasePath = "/files/"
# BasePath = "https://example.com/files" # external URL for use behind reverse proxy

# Cross-Origin Resource Sharing (CORS)
# 	If the server will be accessed from a different Origin than the KiwiIRC
# 	client, it is necessary to explicitly allow the KiwiIRC origin. See
# 	https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS for more detail
CorsOrigins = []
# CorsOrigins = [ "http://example.com" , "https://example.org" ]
# CorsOrigins = [ "*" ] # to allow all

# Requests from these networks will have their X-Forwarded-For headers trusted
TrustedReverseProxyRanges = [
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"fc00::/7",
	"127.0.0.0/8",
	"::1/128",
]

# Restrict server to users identified by valid JWT Account
RequireJwtAccount = false

[Storage]
Path = "./uploads"
ShardLayers = 6
MaximumUploadSize = "10 MB" # accepts units such as: MB, g, tB, peta, kilobytes, gigabyte

[Database]
Type = "sqlite3" # sqlite3 | mysql

# for sqlite3: a filesystem path
# for mysql: a DSN like "user:password@tcp(127.0.0.1:3306)/database". see https://github.com/go-sql-driver/mysql#dsn-data-source-name
Path = "./uploads.db"

[Expiration]
# Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
MaxAge = "24h" # 1 day
IdentifiedMaxAge = "168h" # 1 week
CheckInterval = "5m"

# If EXTJWT is supported by the gateway or network, a validated token with an account present (when
# the user is authenticated to an irc services account) will use the IdentifiedMaxAge setting above
# instead of the base MaxAge.
#
# The HMAC secret used to sign the token is needed here to be able to validate the token.
#
# When using a webircgateway, the issuer will be the network_common_address of the upstream server
# if set. Otherwise it will be the hostname used to connect to the network.
#
# You can set a fallback issuer as "*" which will be attempted if no other matching hosts exist
# this can be handy if webircgateway is in gateway mode and providing EXTJWT for unknown networks
[JwtSecretsByIssuer]
# "example.com" = "examplesecret"
# "169.254.0.0" = "anothersecret"

# Categories allow different retention policies and storage locations per upload category.
# Clients pass a "category" metadata field (e.g. Upload-Metadata: category <base64-value>).
# If a category is set and not listed here, the upload is rejected with HTTP 400.
# StoragePrefix places files under <Storage.Path>/<StoragePrefix>/complete/... instead of
# the default <Storage.Path>/complete/... — useful to isolate files from an external cleanup job.
#
# WARNING: the category field is client-supplied and is not cryptographically verified in v1.
# Any connected Kiwi user can request a category. Apply storage limits accordingly.
#
# [Categories.abuse-report]
# MaxAge           = "8760h" # 1 year
# IdentifiedMaxAge = "8760h" # 1 year
# StoragePrefix    = "reports" # stored at <Storage.Path>/reports/complete/...

# PreFinishCommands allows system commands to be run based on minetype once the file is fully uploaded
# but before it is hashed and moved from incomplete so the file can be rejected using RejectOnNoneZeroExit
# %FILE% will be replace with the full path to the file within [Storage.Path]/incomplete/
# There can be multiple definitions for [[PreFinishCommands]] and they will be evaluated in order
# To aid with debugging set RejectOnNoneZeroExit true and loglevel debug
# "Pattern" can include wildcards * and/or ?
#
# The below example uses exiv2 to strip GPSInfo from image files
# You would need exiv2 installed on the system for it to work
# [[PreFinishCommands]]
# Pattern = "image/*"
# Command = "/usr/bin/exiv2"
# Args = [
# 	"-M del Exif.GPSInfo.GPSVersionID",
# 	"-M del Exif.GPSInfo.GPSLatitudeRef",
# 	"-M del Exif.GPSInfo.GPSLatitude",
# 	"-M del Exif.GPSInfo.GPSLongitudeRef",
# 	"-M del Exif.GPSInfo.GPSLongitude",
# 	"-M del Exif.GPSInfo.GPSAltitudeRef",
# 	"-M del Exif.GPSInfo.GPSAltitude",
# 	"-M del Exif.GPSInfo.GPSTimeStamp",
# 	"-M del Exif.GPSInfo.GPSSatellites",
# 	"-M del Exif.GPSInfo.GPSStatus",
# 	"-M del Exif.GPSInfo.GPSMeasureMode",
# 	"-M del Exif.GPSInfo.GPSDOP",
# 	"-M del Exif.GPSInfo.GPSSpeedRef",
# 	"-M del Exif.GPSInfo.GPSSpeed",
# 	"-M del Exif.GPSInfo.GPSTrackRef",
# 	"-M del Exif.GPSInfo.GPSTrack",
# 	"-M del Exif.GPSInfo.GPSImgDirectionRef",
# 	"-M del Exif.GPSInfo.GPSImgDirection",
# 	"-M del Exif.GPSInfo.GPSMapDatum",
# 	"-M del Exif.GPSInfo.GPSDestLatitudeRef",
# 	"-M del Exif.GPSInfo.GPSDestLatitude",
# 	"-M del Exif.GPSInfo.GPSDestLongitudeRef",
# 	"-M del Exif.GPSInfo.GPSDestLongitude",
# 	"-M del Exif.GPSInfo.GPSDestBearingRef",
# 	"-M del Exif.GPSInfo.GPSDestBearing",
# 	"-M del Exif.GPSInfo.GPSDestDistanceRef",
# 	"-M del Exif.GPSInfo.GPSDestDistance",
# 	"-M del Exif.GPSInfo.GPSProcessingMethod",
# 	"-M del Exif.GPSInfo.GPSAreaInformation",
# 	"-M del Exif.GPSInfo.GPSDateStamp",
# 	"-M del Exif.GPSInfo.GPSDifferential",
# 	"-M del Exif.GPSInfo.GPSHPositioningError",
# 	"%FILE%"
# ]
# RejectOnNoneZeroExit = false

[Ads]
# Controls who sees advertisements on file preview pages.
# Mode = "none"      # Ads disabled for everyone (default)
# Mode = "all"       # Ads enabled for everyone
# Mode = "allowlist" # Ads enabled only for listed IP addresses
# AllowedIPs = ["192.168.1.10", "203.0.113.5"]
#
# Ad provider selection (defaults to "ezoic" when omitted).
# Provider = "ezoic"  # Ezoic standalone ads (zones 118–121)
# Provider = "google" # Google Auto Ads
#
# Ezoic: URL to managed ads.txt (updated daily). Leave empty to disable /ads.txt route.
# AdsTxtURL = "https://srv.adstxtmanager.com/19390/europnet.org"
#
# Google Ads (used when Provider = "google"):
# GooglePublisherId = "ca-pub-7308685930931350"
# GoogleConsentNonce = ""  # optional nonce for Content-Security-Policy

[[Loggers]]
Level = "info" # debug | info | warn | error | fatal | panic
Format = "pretty" # pretty | json
Output = "stderr:" # stderr: | stdout: | file:/path | udp:ip:port | unix:/path | locking-stderr: | locking-stdout:

# [[Loggers]]
# Level = "debug"
# Format = "json"
# Output = "file:./debug.log"

# [[Loggers]]
# # example receiver for testing: socat UDP-RECV:10101,bind=127.0.0.1 STDOUT
# Level = "info"
# Format = "json"
# Output = "udp:127.0.0.1:10101"

# [[Loggers]]
# # example receiver for testing: socat UNIX-LISTEN:./log.sock,fork STDOUT
# Level = "info"
# Format = "json"
# Output = "unix:./log.sock" # filesystem path to unix socket
`
