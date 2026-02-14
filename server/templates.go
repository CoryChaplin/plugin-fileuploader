package server

import (
	"fmt"
	"html/template"
	"time"
)

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
	IsImage bool
	IsVideo bool
	IsAudio bool
	IsPDF   bool
	IsText  bool

	// Content for text files
	TextContent string // Text file content (if IsText == true)
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
<html lang="fr">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Filename}}</title>

	<!-- Open Graph pour prévisualisations sociales -->
	{{if .IsImage}}
	<!-- Image only: minimal metadata for direct image display in Embed.ly -->
	<meta property="og:image" content="{{.DirectURL}}">
	{{else}}
	<!-- Non-image: full metadata for rich preview card -->
	<meta property="og:title" content="{{.Filename}}">
	<meta property="og:type" content="website">
	<meta property="og:url" content="{{.PageURL}}">
	<meta property="og:description" content="Fichier partagé - {{.FileSizeHuman}}">
	{{end}}

	<!-- Empêcher l'indexation (privacy) -->
	<meta name="robots" content="noindex, nofollow">

	<!-- Security headers -->
	<meta http-equiv="Content-Security-Policy" content="default-src 'self' https://chat.europnet.org http://www.chat-fr.org http://quote.europnet.org; script-src 'unsafe-inline' https://www.googletagmanager.com; style-src 'unsafe-inline'; media-src 'self'; img-src 'self' https://chat.europnet.org; font-src https://cdnjs.cloudflare.com; connect-src https://www.google-analytics.com https://www.googletagmanager.com;">

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
			align-items: center;
			justify-content: center;
			padding: 70px 20px 20px;
		}

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
			font-size: 14px;
			line-height: 1.5;
			max-height: calc(100vh - 120px);
			overflow-y: auto;
			text-align: left;
			border: 1px solid #ddd;
			box-shadow: 0 4px 12px rgba(0,0,0,0.2);
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
		}
	</style>
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
		<main>
			{{if .IsImage}}
			<img src="{{.DirectURL}}" alt="{{.Filename}}" class="preview-image">
			{{else if .IsVideo}}
			<video controls preload="metadata" class="preview-video">
				<source src="{{.DirectURL}}" type="{{.MimeType}}">
				Votre navigateur ne supporte pas la lecture de vidéos.
			</video>
			{{else if .IsAudio}}
			<audio controls preload="metadata" class="preview-audio">
				<source src="{{.DirectURL}}" type="{{.MimeType}}">
				Votre navigateur ne supporte pas la lecture audio.
			</audio>
			{{else if .IsPDF}}
			<iframe src="{{.DirectURL}}" class="preview-pdf"></iframe>
			{{else if .IsText}}
			<pre class="preview-text">{{.TextContent}}</pre>
			{{end}}
		</main>
	</div>

	<!-- Google tag (gtag.js) -->
	<script async src="https://www.googletagmanager.com/gtag/js?id=G-69ZMVPJMVF"></script>
	<script>
		window.dataLayer = window.dataLayer || [];
		function gtag(){dataLayer.push(arguments);}
		gtag('js', new Date());
		gtag('config', 'G-69ZMVPJMVF');
	</script>
</body>
</html>`

// ParseFileViewTemplate parses and returns the file view template
func ParseFileViewTemplate() (*template.Template, error) {
	return template.New("fileview").Parse(fileViewTemplateHTML)
}
