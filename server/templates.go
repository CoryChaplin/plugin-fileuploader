package server

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"
	"time"
)

// faviconIcoBytes holds the decoded favicon.ico binary, served at /favicon.ico
var faviconIcoBytes []byte

// appleTouchIconBytes holds the decoded 180x180 PNG, served at /apple-touch-icon*.png
var appleTouchIconBytes []byte

func init() {
	// Extract base64 payload from the data URI in the shortcut icon link tag
	const dataURI = `data:image/x-icon;base64,AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABMLAAATCwAAAAAAAAAAAAB7ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP99URj/gFQc/4BVHf+AVR3/gFUd/4BVHf9/VBz/fFAW/3tOFP97ThT/fVAX/3xQFv97ThT/e04U/3tOFP96TRL/sJRx/9rOvv/WyLb/1si2/9bItv/WyLb/28+//6aIYf95TBH/ekwS/62Rbv+ig1r/eUwR/3tOFP97ThT/ekwS/8WxmP/MuqT/mnhL/5h1SP+YdUj/mXZK/93Rw/+7o4X/eUsQ/3lMEf/CrZL/sph2/3lLEP97ThT/e04U/3pMEv/FsZf/7efg/7OZd/95SxH/eUsQ/3pNEv/Sw6//u6SG/3hLEP95SxD/wayR/7KYdv95SxD/e04U/3tOFP96TBL/xbGY/8+/qf+UcEH/g1kj/4NZIv9+Uhr/0sOv/8Crj/+BVx//glcg/8aymf+ymHb/eUsQ/3tOFP97ThT/ekwS/72mif/l3dH/18q4/9nMu//Yy7n/n39U/8WymP/o4Nb/2cy7/9nLu//n39X/q45p/3lLEP97ThT/e04U/3tOFP+GXCf/kmw8/5JtPf+SbT3/kmw8/4NZIv+GXSj/km09/5NuPv+SbT7/kWs7/4JYIf97ThP/e04U/3tOFP97ThT/ek0T/3pMEv96TBL/ekwS/3pMEv97TRP/ek0T/3pMEf96TBL/ekwS/3pMEv97TRP/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==`
	b64 := strings.TrimPrefix(dataURI, "data:image/x-icon;base64,")
	var err error
	faviconIcoBytes, err = base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic("invalid favicon.ico base64: " + err.Error())
	}

	// Apple touch icon (180x180 PNG) — iOS/Safari requests /apple-touch-icon*.png at root
	const appleTouchDataURI = `data:image/png;base64,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`
	atb64 := strings.TrimPrefix(appleTouchDataURI, "data:image/png;base64,")
	appleTouchIconBytes, err = base64.StdEncoding.DecodeString(atb64)
	if err != nil {
		panic("invalid apple-touch-icon base64: " + err.Error())
	}
}

// FileView contains data for rendering file preview HTML
type FileView struct {
	Filename      string    // Original filename
	FileSize      int64     // Size in bytes
	FileSizeHuman string    // Human-readable size (e.g., "2.5 MB")
	MimeType      string    // MIME type
	UploadDate    time.Time // Upload timestamp
	ExpiresAt     time.Time // Expiration timestamp
	PageURL       string    // URL of the HTML page (for og:url)
	DirectURL     string    // URL for direct download with ?raw=1

	// Flags for conditional display
	IsImage    bool
	IsVideo    bool
	IsAudio    bool
	IsPDF      bool
	IsText     bool
	IsMarkdown bool // Markdown-formatted text (.md, .markdown, .txt)

	// Content for text files
	TextContent string // Text file content (if IsText == true)

	// Ads
	ShowAds            bool   // true if ads should be shown to this visitor
	AdProvider         string // "ezoic" | "google"
	GooglePublisherId  string // Google publisher ID (Google Ads only)
	GoogleConsentNonce string // optional CSP nonce (Google Ads only)

	// Localisation
	T *Translations
}

// humanizeBytes converts bytes to human-readable format
func humanizeBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// fileViewTemplate is the HTML template for file preview pages
var fileViewTemplateHTML = `<!DOCTYPE html>
<html lang="{{.T.Lang}}">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Filename}}</title>

	<!-- Favicon -->
	<link rel="shortcut icon" href="data:image/x-icon;base64,AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABMLAAATCwAAAAAAAAAAAAB7ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP99URj/gFQc/4BVHf+AVR3/gFUd/4BVHf9/VBz/fFAW/3tOFP97ThT/fVAX/3xQFv97ThT/e04U/3tOFP96TRL/sJRx/9rOvv/WyLb/1si2/9bItv/WyLb/28+//6aIYf95TBH/ekwS/62Rbv+ig1r/eUwR/3tOFP97ThT/ekwS/8WxmP/MuqT/mnhL/5h1SP+YdUj/mXZK/93Rw/+7o4X/eUsQ/3lMEf/CrZL/sph2/3lLEP97ThT/e04U/3pMEv/FsZf/7efg/7OZd/95SxH/eUsQ/3pNEv/Sw6//u6SG/3hLEP95SxD/wayR/7KYdv95SxD/e04U/3tOFP96TBL/xbGY/8+/qf+UcEH/g1kj/4NZIv9+Uhr/0sOv/8Crj/+BVx//glcg/8aymf+ymHb/eUsQ/3tOFP97ThT/ekwS/72mif/l3dH/18q4/9nMu//Yy7n/n39U/8WymP/o4Nb/2cy7/9nLu//n39X/q45p/3lLEP97ThT/e04U/3tOFP+GXCf/kmw8/5JtPf+SbT3/kmw8/4NZIv+GXSj/km09/5NuPv+SbT7/kWs7/4JYIf97ThP/e04U/3tOFP97ThT/ek0T/3pMEv96TBL/ekwS/3pMEv97TRP/ek0T/3pMEf96TBL/ekwS/3pMEv97TRP/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==" type="image/x-icon">
	<link rel="icon" type="image/png" sizes="16x16" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAA0lBMVEUUTnsTTXoSTHoTTXsRTHonXIY9bZI+bpMiWYMoXYY+bZI/bpM8bJEiWIMTTnuJpr3S3ea6y9i8zdq6y9lUfp+XscXW4Oi8zdnV4OhpjqsQS3mYssWpvs9BcJQjWoQjWYMaUn6uwtKPq8AgV4IhWIKZssZ1l7GYscXf5+13mLIRS3kSTXquwtGGpLsPS3iRrMGjuctKd5lHdJhIdZjB0dyFo7qSrcJylbC+ztq1x9W1yNa/z9thiKYRTHluka1ag6IYUX0dVYAcVIAXUHwXUH3///9eflIhAAAAAWJLR0RFjrOoVwAAAAlwSFlzAAALEwAACxMBAJqcGAAAAAd0SU1FB+oCEg8dKU1Z7X0AAAB/SURBVBjTY2CgCmBkAgJmRhYWFiZmsAArGzs7GwcnFzc7Dy8fkM/ELyAoJCwiKiYuLiEpBRKQlpGVk1dQVFJWUVUDC6hraGpJaevo6mnpQwSkDQyNjIxNTKVYzMAC2uYWllZWVtY2tkx29rYgWxwcQcDJmYHBxRniED4QoMAjAM00DSZ3WDnzAAAAJXRFWHRkYXRlOmNyZWF0ZQAyMDI2LTAyLTE4VDE1OjI4OjU4KzAwOjAwTviF3AAAACV0RVh0ZGF0ZTptb2RpZnkAMjAyNi0wMi0xOFQxNToyODo1OCswMDowMD+lPWAAAAAgdEVYdHNvZnR3YXJlAGh0dHBzOi8vaW1hZ2VtYWdpY2sub3JnvM8dnQAAABh0RVh0VGh1bWI6OkRvY3VtZW50OjpQYWdlcwAxp/+7LwAAABh0RVh0VGh1bWI6OkltYWdlOjpIZWlnaHQAMTkyQF1xVQAAABd0RVh0VGh1bWI6OkltYWdlOjpXaWR0aAAxOTLTrCEIAAAAGXRFWHRUaHVtYjo6TWltZXR5cGUAaW1hZ2UvcG5nP7JWTgAAABd0RVh0VGh1bWI6Ok1UaW1lADE3NzE0Mjg1Mzi+HkkOAAAAD3RFWHRUaHVtYjo6U2l6ZQAwQkKUoj7sAAAAVnRFWHRUaHVtYjo6VVJJAGZpbGU6Ly8vbW50bG9nL2Zhdmljb25zLzIwMjYtMDItMTgvNjMwZTA2NWI3N2I4NDA2ZjcwODU1MjI4YTM3Zjk5NzkuaWNvLnBuZ2eFPr8AAAAASUVORK5CYII=">
	<link rel="icon" type="image/png" sizes="32x32" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAABWVBMVEUUTnsRTHkSTXoRS3kaUn5BcJRLeJpNeZtMeJpLd5owY4sTTXsaU39CcZVOepxPepxPe5xOeZtMeJs0Zo0RTHo9bZLh6O72+Pr4+vv5+vv6/Py0x9UYUX1XgKHr8PT5+/z6+/z5+vz4+fv3+fuovs4XUH1GdJfy9fju8vWzxtSvw9J7m7Vvkq7///+2yNauwtK9zdr8/f61yNZHdJj19/nK1+EeVoESTHoVT3wWUHwWT3wQS3lwk6/Y4ekqX4g/bpPw9PdIdZj09/ng5+15mrRli6kmXIUTTXrX4egoXYbw9Pb+/v51l7HW4OgnXYb09vmovc6dtshxlK8nXIb1+PocVH8OSXgPSngMSHZukq7V4OhHdJfd5uxtka1li6hmi6ljiaeiucvU3+fw8/ZFc5bS3eYkWoQ7bJHv8/aQrMGft8metsmgt8p4mbMZUn4lW4WOqr9vk64TTnuUU3uzAAAAAWJLR0QtzdpBPQAAAAlwSFlzAAALEwAACxMBAJqcGAAAAAd0SU1FB+oCEg8dKU1Z7X0AAADySURBVDjLY2AYBdQEjDDABGQzgwGYDQMsrGzsIMDBycXNw8vHL8DPJyjEKQyXFxEVE5eQBAIJKWkZWTlxeQUFeUUlZRVVuAVq6hqaWmCgzaCjq64HZOgbGBrJwBUYm5iamVtYWlpacVvb6NraAZkM9g56CAWOTs4urm7uIDZQgYcnkMUo6oWsQF1X19uHEarA188d5C4UBf5yAYFqYG9ZB+n6BmMqCDENDQu3BoGISN0oTAXR/jGxcfFgkJCom+QGVpCMpCDFQRcJpKYBFZinZyAUuNtlZmXDQU5uHkgsv6DQAh6URUwoACLI5D7QKWRoAQBw9jKq341VcAAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAyNi0wMi0xOFQxNToyODo1OCswMDowME74hdwAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMjYtMDItMThUMTU6Mjg6NTgrMDA6MDA/pT1gAAAAIHRFWHRzb2Z0d2FyZQBodHRwczovL2ltYWdlbWFnaWNrLm9yZ7zPHZ0AAAAYdEVYdFRodW1iOjpEb2N1bWVudDo6UGFnZXMAMaf/uy8AAAAYdEVYdFRodW1iOjpJbWFnZTo6SGVpZ2h0ADE5MkBdcVUAAAAXdEVYdFRodW1iOjpJbWFnZTo6V2lkdGgAMTky06whCAAAABl0RVh0VGh1bWI6Ok1pbWV0eXBlAGltYWdlL3BuZz+yVk4AAAAXdEVYdFRodW1iOjpNVGltZQAxNzcxNDI4NTM4vh5JDgAAAA90RVh0VGh1bWI6OlNpemUAMEJClKI+7AAAAFZ0RVh0VGh1bWI6OlVSSQBmaWxlOi8vL21udGxvZy9mYXZpY29ucy8yMDI2LTAyLTE4LzYzMGUwNjViNzdiODQwNmY3MDg1NTIyOGEzN2Y5OTc5Lmljby5wbmdnhT6/AAAAAElFTkSuQmCC">
	<link rel="apple-touch-icon" sizes="180x180" href="data:image/png;base64,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">

	<!-- Open Graph pour prévisualisations sociales -->
	{{if .IsImage}}
	<!-- Image only: minimal metadata for direct image display in Embed.ly -->
	<meta property="og:image" content="{{.DirectURL}}">
	{{else}}
	<!-- Non-image: full metadata for rich preview card -->
	<meta property="og:title" content="{{.Filename}}">
	<meta property="og:type" content="website">
	<meta property="og:url" content="{{.PageURL}}">
	<meta property="og:description" content="{{.T.SharedFile}} - {{.FileSizeHuman}}">
	{{end}}

	<!-- Empêcher l'indexation (privacy) -->
	<meta name="robots" content="noindex, nofollow">

	<!-- Security headers -->
	<meta http-equiv="Content-Security-Policy" content="default-src 'self' https://chat.europnet.org http://www.chat-fr.org http://quote.europnet.org; script-src * 'unsafe-inline'; style-src 'unsafe-inline' https://cdnjs.cloudflare.com https://fonts.googleapis.com; media-src 'self'; img-src 'self' data: https:; font-src https://cdnjs.cloudflare.com https://fonts.gstatic.com; connect-src https:; frame-src https:;">

	<!-- Font Awesome for navbar icons -->
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">

	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
			background: #144e7b;
			color: #333;
			line-height: 1.6;
			padding: 0;
			margin: 0;
			min-height: 100vh;
			display: flex;
			flex-direction: column;
		}

		/* EuropNet Navbar */
		#navbar {
			position: fixed;
			top: 0;
			width: 100%;
			margin: 0;
			padding-left: 0;
			padding-right: 0;
			border-width: 0;
			border-radius: 0;
			-webkit-box-shadow: none;
			box-shadow: none;
			height: 50px;
			background: #103d61;
			z-index: 1000;
		}

		#navbar .logo {
			height: 50px;
			position: relative;
			display: block;
			width: 150px;
			float: left;
			margin-left: 10px;
		}

		#navbar .nav ul {
			display: block;
			float: left;
			list-style: outside none none;
			font-family: "Helvetica Neue",Helvetica,Arial,sans-serif;
			color: rgb(255, 255, 255);
			line-height: 1.42857;
			font-size: 14px;
			padding: 0;
			margin: 0;
		}

		#navbar .nav li {
			display: block;
			float: left;
			position: relative;
		}

		#navbar .nav li a {
			text-decoration: none;
			color: #FFF;
			padding: 0.2em 0.5em;
			font-size: 1em;
			margin: 11px 7px;
			display: inline-block;
			background: #428BCA none repeat scroll 0;
			font-weight: normal;
			line-height: 1.42857;
			text-align: center;
			white-space: nowrap;
			vertical-align: middle;
			cursor: pointer;
			-moz-user-select: none;
			border: 1px solid transparent;
			border-radius: 4px;
		}

		#navbar .nav li a i {
			padding-right: 0.5em;
		}

		#navbar .nav li a:hover {
			background-color: #3071a9;
			border-color: #428BCA;
		}

		.container {
			flex: 1;
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			padding: 70px 20px 20px;
		}

		{{template "ads-css" .}}

		main {
			width: 100%;
			max-width: 1200px;
			text-align: center;
		}

		.preview-image {
			max-width: 100%;
			max-height: calc(100vh - 100px);
			height: auto;
			border-radius: 4px;
			box-shadow: 0 4px 12px rgba(0,0,0,0.2);
		}

		.preview-video,
		.preview-audio {
			max-width: 100%;
			border-radius: 4px;
			box-shadow: 0 4px 12px rgba(0,0,0,0.2);
		}

		.preview-video {
			max-height: calc(100vh - 100px);
		}

		.preview-pdf {
			width: 100%;
			height: calc(100vh - 90px);
			border: none;
			border-radius: 4px;
			box-shadow: 0 4px 12px rgba(0,0,0,0.2);
		}

		.preview-text {
			background: #f5f5f5;
			padding: 20px;
			border-radius: 4px;
			overflow-x: auto;
			white-space: pre-wrap;
			word-wrap: break-word;
			font-family: 'Courier New', Courier, monospace;
			font-size: 1.1em;
			line-height: 1.5;
			max-height: calc(100vh - 120px);
			overflow-y: auto;
			text-align: left;
			border: 1px solid #ddd;
			box-shadow: 0 4px 12px rgba(0,0,0,0.2);
		}

		.preview-markdown {
			background: #fff;
			padding: 30px 40px;
			border-radius: 4px;
			max-height: calc(100vh - 120px);
			overflow-y: auto;
			text-align: left;
			border: 1px solid #ddd;
			box-shadow: 0 4px 12px rgba(0,0,0,0.2);
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
			font-size: 15px;
			line-height: 1.7;
			color: #1a1a1a;
		}

		.preview-markdown h1, .preview-markdown h2, .preview-markdown h3,
		.preview-markdown h4, .preview-markdown h5, .preview-markdown h6 {
			margin-top: 1.2em;
			margin-bottom: 0.6em;
			font-weight: 600;
			line-height: 1.3;
		}

		.preview-markdown h1 { font-size: 1.8em; border-bottom: 1px solid #ddd; padding-bottom: 0.3em; }
		.preview-markdown h2 { font-size: 1.5em; border-bottom: 1px solid #eee; padding-bottom: 0.3em; }
		.preview-markdown h3 { font-size: 1.25em; }

		.preview-markdown p { margin-bottom: 1em; }

		.preview-markdown code {
			background: #f0f0f0;
			padding: 0.2em 0.4em;
			border-radius: 3px;
			font-size: 0.9em;
			font-family: 'Courier New', Courier, monospace;
		}

		.preview-markdown pre {
			background: #f5f5f5;
			padding: 16px;
			border-radius: 4px;
			overflow-x: auto;
			border: 1px solid #ddd;
			margin-bottom: 1em;
		}

		.preview-markdown pre code {
			background: none;
			padding: 0;
			font-size: 0.85em;
			line-height: 1.5;
		}

		.preview-markdown blockquote {
			border-left: 4px solid #ddd;
			padding: 0.5em 1em;
			margin: 0 0 1em 0;
			color: #555;
			background: #f9f9f9;
		}

		.preview-markdown ul, .preview-markdown ol {
			margin-bottom: 1em;
			padding-left: 2em;
		}

		.preview-markdown li { margin-bottom: 0.3em; }

		.preview-markdown table {
			border-collapse: collapse;
			margin-bottom: 1em;
			width: 100%;
		}

		.preview-markdown th, .preview-markdown td {
			border: 1px solid #ddd;
			padding: 8px 12px;
			text-align: left;
		}

		.preview-markdown th {
			background: #f5f5f5;
			font-weight: 600;
		}

		.preview-markdown img {
			max-width: 100%;
			height: auto;
		}

		.preview-markdown a {
			color: #0366d6;
			text-decoration: none;
		}

		.preview-markdown a:hover {
			text-decoration: underline;
		}

		.preview-markdown hr {
			border: none;
			border-top: 1px solid #ddd;
			margin: 1.5em 0;
		}

		/* Dark mode support */
		@media (prefers-color-scheme: dark) {
			body {
				background: #0d2d47;
				color: #e0e0e0;
			}

			.preview-text {
				background: #1a1a1a;
				color: #e0e0e0;
				border-color: #444;
			}

			.preview-markdown {
				background: #1e1e1e;
				color: #d4d4d4;
				border-color: #444;
			}

			.preview-markdown h1, .preview-markdown h2 {
				border-bottom-color: #444;
			}

			.preview-markdown code {
				background: #2d2d2d;
			}

			.preview-markdown pre {
				background: #2d2d2d;
				border-color: #444;
			}

			.preview-markdown blockquote {
				border-left-color: #555;
				background: #252525;
				color: #aaa;
			}

			.preview-markdown th {
				background: #2d2d2d;
			}

			.preview-markdown th, .preview-markdown td {
				border-color: #444;
			}

			.preview-markdown a {
				color: #58a6ff;
			}

			.preview-markdown hr {
				border-top-color: #444;
			}

			.preview-image,
			.preview-video,
			.preview-audio,
			.preview-pdf {
				box-shadow: 0 4px 12px rgba(0,0,0,0.5);
			}
		}

		/* Mobile responsive */
		@media (max-width: 768px) {
			.container {
				padding: 60px 10px 10px;
			}

			{{template "ads-css-mobile" .}}

			#navbar .nav {
				display: none;
			}

			.preview-image {
				max-height: calc(100vh - 80px);
			}

			.preview-video {
				max-height: calc(100vh - 80px);
			}

			.preview-pdf {
				height: calc(100vh - 70px);
			}

			.preview-text {
				max-height: calc(100vh - 90px);
			}

			.preview-markdown {
				padding: 16px 20px;
				max-height: calc(100vh - 90px);
			}
		}
	</style>

	{{template "ads-head" .}}
</head>
<body>
	<!-- EuropNet Navbar -->
	<div class="navbar" id="navbar">
		<div class="logo">
			<img src="https://chat.europnet.org/static/logo-en-white.png" height="50" alt="EuropNet" title="EuropNet">
		</div>
		<div class="nav">
			<ul>
				<li>
					<a href="http://www.chat-fr.org/photos" target="_blank" title="Photos chatteurs">
						<i class="fa fa-camera" title="Photos chatteurs"></i>Photos chatteurs
					</a>
				</li>
				<li>
					<a href="http://www.chat-fr.org/rencontre" target="_blank" title="Annuaire chatteurs">
						<i class="fa fa-book" title="Annuaire chatteurs"></i>Annuaire chatteurs
					</a>
				</li>
				<li>
					<a href="http://www.chat-fr.org/register" target="_blank" title="Créez votre fiche">
						<i class="fa fa-user" title="Crééz votre fiche"></i>Créez votre fiche
					</a>
				</li>
				<li>
					<a href="http://quote.europnet.org" target="_blank" title="Quotes">
						<i class="fa fa-quote-right" title="Quotes"></i>Quotes
					</a>
				</li>
			</ul>
		</div>
	</div>

	<div class="container">
		{{if .ShowAds}}
		<div class="content-row">
			{{if eq .AdProvider "ezoic"}}<div class="ad-sidebar ad-sidebar-left"><div id="ezoic-pub-ad-placeholder-118"></div></div>{{end}}
			<main>{{template "file-content" .}}</main>
			{{if eq .AdProvider "ezoic"}}<div class="ad-sidebar ad-sidebar-right"><div id="ezoic-pub-ad-placeholder-121"></div></div>{{end}}
		</div>
		{{if eq .AdProvider "ezoic"}}
		<div class="ad-below-content"><div id="ezoic-pub-ad-placeholder-119"></div></div>
		<div class="ad-below-content"><div id="ezoic-pub-ad-placeholder-120"></div></div>
		{{end}}
		{{else}}
		<main>{{template "file-content" .}}</main>
		{{end}}
	</div>

	<!-- Google tag (gtag.js) -->
	<script async src="https://www.googletagmanager.com/gtag/js?id=G-69ZMVPJMVF"></script>
	<script>
		window.dataLayer = window.dataLayer || [];
		function gtag(){dataLayer.push(arguments);}
		gtag('js', new Date());
		gtag('config', 'G-69ZMVPJMVF');
		{{if .ShowAds}}gtag('event', 'ads', {eligible: true});{{else}}gtag('event', 'ads', {eligible: false});{{end}}
	</script>

	{{template "ads-scripts" .}}
</body>
</html>`

// ParseFileViewTemplate parses and returns the file view template,
// composed with the ad sub-templates defined in ads_template.go.
func ParseFileViewTemplate() (*template.Template, error) {
	t, err := template.New("fileview").Parse(fileViewTemplateHTML)
	if err != nil {
		return nil, err
	}
	return t.Parse(adsSubTemplates)
}

// NotFoundView contains data for rendering the 404 error page
type NotFoundView struct {
	MaxAge             string // e.g. "24 heures" / "24 hours"
	IdentifiedMaxAge   string // e.g. "7 jours" / "7 days"
	ShowAds            bool
	AdProvider         string // "ezoic" | "google"
	GooglePublisherId  string // Google publisher ID (Google Ads only)
	GoogleConsentNonce string // optional CSP nonce (Google Ads only)

	// Localisation
	T *Translations
}

// notFoundTemplateHTML is the HTML template for the 404 error page
var notFoundTemplateHTML = `<!DOCTYPE html>
<html lang="{{.T.Lang}}">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.T.PageTitleNotFound}}</title>

	<!-- Favicon -->
	<link rel="shortcut icon" href="data:image/x-icon;base64,AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAABMLAAATCwAAAAAAAAAAAAB7ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP99URj/gFQc/4BVHf+AVR3/gFUd/4BVHf9/VBz/fFAW/3tOFP97ThT/fVAX/3xQFv97ThT/e04U/3tOFP96TRL/sJRx/9rOvv/WyLb/1si2/9bItv/WyLb/28+//6aIYf95TBH/ekwS/62Rbv+ig1r/eUwR/3tOFP97ThT/ekwS/8WxmP/MuqT/mnhL/5h1SP+YdUj/mXZK/93Rw/+7o4X/eUsQ/3lMEf/CrZL/sph2/3lLEP97ThT/e04U/3pMEv/FsZf/7efg/7OZd/95SxH/eUsQ/3pNEv/Sw6//u6SG/3hLEP95SxD/wayR/7KYdv95SxD/e04U/3tOFP96TBL/xbGY/8+/qf+UcEH/g1kj/4NZIv9+Uhr/0sOv/8Crj/+BVx//glcg/8aymf+ymHb/eUsQ/3tOFP97ThT/ekwS/72mif/l3dH/18q4/9nMu//Yy7n/n39U/8WymP/o4Nb/2cy7/9nLu//n39X/q45p/3lLEP97ThT/e04U/3tOFP+GXCf/kmw8/5JtPf+SbT3/kmw8/4NZIv+GXSj/km09/5NuPv+SbT7/kWs7/4JYIf97ThP/e04U/3tOFP97ThT/ek0T/3pMEv96TBL/ekwS/3pMEv97TRP/ek0T/3pMEf96TBL/ekwS/3pMEv97TRP/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/e04U/3tOFP97ThT/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==" type="image/x-icon">
	<link rel="icon" type="image/png" sizes="16x16" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAMAAAAoLQ9TAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAA0lBMVEUUTnsTTXoSTHoTTXsRTHonXIY9bZI+bpMiWYMoXYY+bZI/bpM8bJEiWIMTTnuJpr3S3ea6y9i8zdq6y9lUfp+XscXW4Oi8zdnV4OhpjqsQS3mYssWpvs9BcJQjWoQjWYMaUn6uwtKPq8AgV4IhWIKZssZ1l7GYscXf5+13mLIRS3kSTXquwtGGpLsPS3iRrMGjuctKd5lHdJhIdZjB0dyFo7qSrcJylbC+ztq1x9W1yNa/z9thiKYRTHluka1ag6IYUX0dVYAcVIAXUHwXUH3///9eflIhAAAAAWJLR0RFjrOoVwAAAAlwSFlzAAALEwAACxMBAJqcGAAAAAd0SU1FB+oCEg8dKU1Z7X0AAAB/SURBVBjTY2CgCmBkAgJmRhYWFiZmsAArGzs7GwcnFzc7Dy8fkM/ELyAoJCwiKiYuLiEpBRKQlpGVk1dQVFJWUVUDC6hraGpJaevo6mnpQwSkDQyNjIxNTKVYzMAC2uYWllZWVtY2tkx29rYgWxwcQcDJmYHBxRniED4QoMAjAM00DSZ3WDnzAAAAJXRFWHRkYXRlOmNyZWF0ZQAyMDI2LTAyLTE4VDE1OjI4OjU4KzAwOjAwTviF3AAAACV0RVh0ZGF0ZTptb2RpZnkAMjAyNi0wMi0xOFQxNToyODo1OCswMDowMD+lPWAAAAAgdEVYdHNvZnR3YXJlAGh0dHBzOi8vaW1hZ2VtYWdpY2sub3JnvM8dnQAAABh0RVh0VGh1bWI6OkRvY3VtZW50OjpQYWdlcwAxp/+7LwAAABh0RVh0VGh1bWI6OkltYWdlOjpIZWlnaHQAMTkyQF1xVQAAABd0RVh0VGh1bWI6OkltYWdlOjpXaWR0aAAxOTLTrCEIAAAAGXRFWHRUaHVtYjo6TWltZXR5cGUAaW1hZ2UvcG5nP7JWTgAAABd0RVh0VGh1bWI6Ok1UaW1lADE3NzE0Mjg1Mzi+HkkOAAAAD3RFWHRUaHVtYjo6U2l6ZQAwQkKUoj7sAAAAVnRFWHRUaHVtYjo6VVJJAGZpbGU6Ly8vbW50bG9nL2Zhdmljb25zLzIwMjYtMDItMTgvNjMwZTA2NWI3N2I4NDA2ZjcwODU1MjI4YTM3Zjk5NzkuaWNvLnBuZ2eFPr8AAAAASUVORK5CYII=">
	<link rel="icon" type="image/png" sizes="32x32" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAABWVBMVEUUTnsRTHkSTXoRS3kaUn5BcJRLeJpNeZtMeJpLd5owY4sTTXsaU39CcZVOepxPepxPe5xOeZtMeJs0Zo0RTHo9bZLh6O72+Pr4+vv5+vv6/Py0x9UYUX1XgKHr8PT5+/z6+/z5+vz4+fv3+fuovs4XUH1GdJfy9fju8vWzxtSvw9J7m7Vvkq7///+2yNauwtK9zdr8/f61yNZHdJj19/nK1+EeVoESTHoVT3wWUHwWT3wQS3lwk6/Y4ekqX4g/bpPw9PdIdZj09/ng5+15mrRli6kmXIUTTXrX4egoXYbw9Pb+/v51l7HW4OgnXYb09vmovc6dtshxlK8nXIb1+PocVH8OSXgPSngMSHZukq7V4OhHdJfd5uxtka1li6hmi6ljiaeiucvU3+fw8/ZFc5bS3eYkWoQ7bJHv8/aQrMGft8metsmgt8p4mbMZUn4lW4WOqr9vk64TTnuUU3uzAAAAAWJLR0QtzdpBPQAAAAlwSFlzAAALEwAACxMBAJqcGAAAAAd0SU1FB+oCEg8dKU1Z7X0AAADySURBVDjLY2AYBdQEjDDABGQzgwGYDQMsrGzsIMDBycXNw8vHL8DPJyjEKQyXFxEVE5eQBAIJKWkZWTlxeQUFeUUlZRVVuAVq6hqaWmCgzaCjq64HZOgbGBrJwBUYm5iamVtYWlpacVvb6NraAZkM9g56CAWOTs4urm7uIDZQgYcnkMUo6oWsQF1X19uHEarA188d5C4UBf5yAYFqYG9ZB+n6BmMqCDENDQu3BoGISN0oTAXR/jGxcfFgkJCom+QGVpCMpCDFQRcJpKYBFZinZyAUuNtlZmXDQU5uHkgsv6DQAh6URUwoACLI5D7QKWRoAQBw9jKq341VcAAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAyNi0wMi0xOFQxNToyODo1OCswMDowME74hdwAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMjYtMDItMThUMTU6Mjg6NTgrMDA6MDA/pT1gAAAAIHRFWHRzb2Z0d2FyZQBodHRwczovL2ltYWdlbWFnaWNrLm9yZ7zPHZ0AAAAYdEVYdFRodW1iOjpEb2N1bWVudDo6UGFnZXMAMaf/uy8AAAAYdEVYdFRodW1iOjpJbWFnZTo6SGVpZ2h0ADE5MkBdcVUAAAAXdEVYdFRodW1iOjpJbWFnZTo6V2lkdGgAMTky06whCAAAABl0RVh0VGh1bWI6Ok1pbWV0eXBlAGltYWdlL3BuZz+yVk4AAAAXdEVYdFRodW1iOjpNVGltZQAxNzcxNDI4NTM4vh5JDgAAAA90RVh0VGh1bWI6OlNpemUAMEJClKI+7AAAAFZ0RVh0VGh1bWI6OlVSSQBmaWxlOi8vL21udGxvZy9mYXZpY29ucy8yMDI2LTAyLTE4LzYzMGUwNjViNzdiODQwNmY3MDg1NTIyOGEzN2Y5OTc5Lmljby5wbmdnhT6/AAAAAElFTkSuQmCC">
	<link rel="apple-touch-icon" sizes="180x180" href="data:image/png;base64,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">

	<!-- Empêcher l'indexation -->
	<meta name="robots" content="noindex, nofollow">

	<!-- Security headers -->
	<meta http-equiv="Content-Security-Policy" content="default-src 'self' https://chat.europnet.org http://www.chat-fr.org http://quote.europnet.org; script-src 'unsafe-inline' https://www.googletagmanager.com; style-src 'unsafe-inline' https://cdnjs.cloudflare.com https://fonts.googleapis.com; media-src 'self'; img-src 'self' data: https://chat.europnet.org; font-src https://cdnjs.cloudflare.com https://fonts.gstatic.com; connect-src https://*.google-analytics.com https://www.googletagmanager.com;">

	<!-- Font Awesome for navbar icons -->
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">

	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }

		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
			background: #144e7b;
			color: #333;
			line-height: 1.6;
			min-height: 100vh;
			display: flex;
			flex-direction: column;
		}

		/* EuropNet Navbar */
		#navbar {
			position: fixed;
			top: 0;
			width: 100%;
			margin: 0;
			padding-left: 0;
			padding-right: 0;
			border-width: 0;
			border-radius: 0;
			-webkit-box-shadow: none;
			box-shadow: none;
			height: 50px;
			background: #103d61;
			z-index: 1000;
		}

		#navbar .logo {
			height: 50px;
			position: relative;
			display: block;
			width: 150px;
			float: left;
			margin-left: 10px;
		}

		#navbar .nav ul {
			display: block;
			float: left;
			list-style: outside none none;
			font-family: "Helvetica Neue",Helvetica,Arial,sans-serif;
			color: rgb(255, 255, 255);
			line-height: 1.42857;
			font-size: 14px;
			padding: 0;
			margin: 0;
		}

		#navbar .nav li {
			display: block;
			float: left;
			position: relative;
		}

		#navbar .nav li a {
			text-decoration: none;
			color: #FFF;
			padding: 0.2em 0.5em;
			font-size: 1em;
			margin: 11px 7px;
			display: inline-block;
			background: #428BCA none repeat scroll 0;
			font-weight: normal;
			line-height: 1.42857;
			text-align: center;
			white-space: nowrap;
			vertical-align: middle;
			cursor: pointer;
			-moz-user-select: none;
			border: 1px solid transparent;
			border-radius: 4px;
		}

		#navbar .nav li a i { padding-right: 0.5em; }

		#navbar .nav li a:hover {
			background-color: #3071a9;
			border-color: #428BCA;
		}

		/* 404 content */
		.error-container {
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			padding: 70px 20px 20px;
		}

		.error-card {
			background: rgba(255, 255, 255, 0.95);
			border-radius: 12px;
			padding: 48px 40px;
			text-align: center;
			max-width: 460px;
			width: 100%;
			box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
		}

		.error-icon {
			color: #144e7b;
			font-size: 3.5rem;
			opacity: 0.45;
			margin-bottom: 2px;
		}

		.error-code {
			font-size: 5rem;
			font-weight: 700;
			color: #144e7b;
			line-height: 1;
			margin: 0 0 16px;
			opacity: 0.6;
		}

		.error-title {
			font-size: 1.25rem;
			font-weight: 600;
			color: #2c3e50;
			margin: 0 0 14px;
		}

		.error-message {
			color: #666;
			font-size: 0.9rem;
			line-height: 1.75;
			margin: 0 0 28px;
		}

		.btn-back {
			display: inline-block;
			background: #428BCA;
			color: white;
			text-decoration: none;
			padding: 10px 22px;
			border-radius: 4px;
			font-size: 0.9rem;
		}

		.btn-back:hover { background: #3071a9; }

		@media (prefers-color-scheme: dark) {
			body { background: #0d2d47; }
			.error-card { background: rgba(16, 61, 97, 0.92); }
			.error-icon { color: #5ba3d9; }
			.error-code { color: #5ba3d9; }
			.error-title { color: #e0e0e0; }
			.error-message { color: #8aacbf; }
		}

		@media (max-width: 768px) {
			.error-container { padding: 60px 16px 16px; }
			#navbar .nav { display: none; }
			.error-card { padding: 32px 24px; }
			.error-code { font-size: 4rem; }
			{{template "ads-css-mobile" .}}
		}
		{{template "ads-css" .}}
	</style>
	{{template "ads-head" .}}
</head>
<body>
	<!-- EuropNet Navbar -->
	<div class="navbar" id="navbar">
		<div class="logo">
			<img src="https://chat.europnet.org/static/logo-en-white.png" height="50" alt="EuropNet" title="EuropNet">
		</div>
		<div class="nav">
			<ul>
				<li>
					<a href="http://www.chat-fr.org/photos" target="_blank" title="Photos chatteurs">
						<i class="fa fa-camera" title="Photos chatteurs"></i>Photos chatteurs
					</a>
				</li>
				<li>
					<a href="http://www.chat-fr.org/rencontre" target="_blank" title="Annuaire chatteurs">
						<i class="fa fa-book" title="Annuaire chatteurs"></i>Annuaire chatteurs
					</a>
				</li>
				<li>
					<a href="http://www.chat-fr.org/register" target="_blank" title="Créez votre fiche">
						<i class="fa fa-user" title="Crééz votre fiche"></i>Créez votre fiche
					</a>
				</li>
				<li>
					<a href="http://quote.europnet.org" target="_blank" title="Quotes">
						<i class="fa fa-quote-right" title="Quotes"></i>Quotes
					</a>
				</li>
			</ul>
		</div>
	</div>

	<div class="error-container">
		<div class="error-card">
			<div class="error-icon"><i class="fa fa-hourglass-end"></i></div>
			<div class="error-code">404</div>
			<h1 class="error-title">{{.T.NotFoundHeading}}</h1>
			<p class="error-message">
				{{.T.FilesKeptFor}} <strong>{{.MaxAge}}</strong><br>
				({{.IdentifiedMaxAge}} {{.T.ForConnected}}).<br>
				{{.T.LinkExpired}}
			</p>
			<a href="https://chat.europnet.org" class="btn-back">
				<i class="fa fa-comments"></i>&nbsp; {{.T.ReturnToChat}}
			</a>
		</div>
	</div>

	{{if and .ShowAds (eq .AdProvider "ezoic")}}
	<div class="ad-below-content"><div id="ezoic-pub-ad-placeholder-119"></div></div>
	<div class="ad-below-content"><div id="ezoic-pub-ad-placeholder-120"></div></div>
	{{end}}

	<!-- Google tag (gtag.js) -->
	<script async src="https://www.googletagmanager.com/gtag/js?id=G-69ZMVPJMVF"></script>
	<script>
		window.dataLayer = window.dataLayer || [];
		function gtag(){dataLayer.push(arguments);}
		gtag('js', new Date());
		gtag('config', 'G-69ZMVPJMVF');
		{{if .ShowAds}}gtag('event', 'ads', {eligible: true});{{else}}gtag('event', 'ads', {eligible: false});{{end}}
	</script>
	{{template "ads-404-scripts" .}}
</body>
</html>`

// ParseNotFoundTemplate parses and returns the 404 error page template
func ParseNotFoundTemplate() (*template.Template, error) {
	t, err := template.New("notfound").Parse(notFoundTemplateHTML)
	if err != nil {
		return nil, err
	}
	return t.Parse(adsSubTemplates)
}
