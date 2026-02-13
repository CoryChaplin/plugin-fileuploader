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
	DirectURL     string    // URL for direct download

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
	<meta property="og:title" content="{{.Filename}}">
	<meta property="og:type" content="website">
	<meta property="og:url" content="{{.DirectURL}}">
	{{if .IsImage}}
	<meta property="og:image" content="{{.DirectURL}}">
	{{end}}
	<meta property="og:description" content="Fichier partagé - {{.FileSizeHuman}}">

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
		}

		.container {
			max-width: 1200px;
			margin: 70px auto 20px;
			background: white;
			border-radius: 8px;
			box-shadow: 0 2px 10px rgba(0,0,0,0.1);
			overflow: hidden;
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

		header {
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			color: white;
			padding: 30px;
		}

		h1 {
			font-size: 24px;
			margin-bottom: 15px;
			word-break: break-word;
		}

		.metadata {
			display: flex;
			flex-wrap: wrap;
			gap: 20px;
			font-size: 14px;
			opacity: 0.9;
		}

		.metadata span {
			display: inline-flex;
			align-items: center;
		}

		.metadata span::before {
			content: "•";
			margin-right: 8px;
		}

		.metadata span:first-child::before {
			content: "";
			margin-right: 0;
		}

		main {
			padding: 30px;
			text-align: center;
		}

		.preview-image {
			max-width: 100%;
			height: auto;
			border-radius: 4px;
			box-shadow: 0 2px 8px rgba(0,0,0,0.1);
		}

		.preview-video,
		.preview-audio {
			max-width: 100%;
			border-radius: 4px;
			box-shadow: 0 2px 8px rgba(0,0,0,0.1);
		}

		.preview-video {
			max-height: 70vh;
		}

		.preview-pdf {
			width: 100%;
			height: 80vh;
			border: 1px solid #ddd;
			border-radius: 4px;
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
			max-height: 80vh;
			overflow-y: auto;
			text-align: left;
			border: 1px solid #ddd;
		}

		footer {
			padding: 30px;
			background: #f9f9f9;
			border-top: 1px solid #eee;
			text-align: center;
		}

		.download-btn {
			display: inline-block;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			color: white;
			padding: 12px 30px;
			border-radius: 4px;
			text-decoration: none;
			font-weight: 500;
			transition: transform 0.2s, box-shadow 0.2s;
		}

		.download-btn:hover {
			transform: translateY(-2px);
			box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
		}

		/* Dark mode support */
		@media (prefers-color-scheme: dark) {
			body {
				background: #0d2d47;
				color: #e0e0e0;
			}

			.container {
				background: #2d2d2d;
				box-shadow: 0 2px 10px rgba(0,0,0,0.3);
			}

			.preview-text {
				background: #1a1a1a;
				color: #e0e0e0;
				border-color: #444;
			}

			footer {
				background: #252525;
				border-top-color: #444;
			}

			.preview-pdf {
				border-color: #444;
			}
		}

		/* Mobile responsive */
		@media (max-width: 768px) {
			.container {
				margin: 60px 10px 10px;
			}

			header {
				padding: 20px;
			}

			h1 {
				font-size: 20px;
			}

			main {
				padding: 20px;
			}

			footer {
				padding: 20px;
			}

			.metadata {
				font-size: 13px;
				gap: 15px;
			}

			#navbar .nav {
				display: none;
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
		<header>
			<h1>{{.Filename}}</h1>
			<div class="metadata">
				<span>Taille : {{.FileSizeHuman}}</span>
				<span>Type : {{.MimeType}}</span>
				{{if not .ExpiresAt.IsZero}}
				<span>Expire : {{.ExpiresAt.Format "02/01/2006 15:04"}}</span>
				{{end}}
			</div>
		</header>

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

		<footer>
			<a href="{{.DirectURL}}" download="{{.Filename}}" class="download-btn">
				⬇️ Télécharger le fichier
			</a>
		</footer>
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
