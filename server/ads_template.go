package server

// adsSubTemplates defines named sub-templates for Ezoic ad integration.
// They are parsed together with fileViewTemplateHTML in ParseFileViewTemplate().
// All blocks are guarded by {{if .ShowAds}} so they produce no output when ads
// are disabled, regardless of which template calls them.
var adsSubTemplates = `

{{define "ads-head"}}{{if .ShowAds}}
{{if eq .AdProvider "google"}}
	<!-- Google Auto Ads -->
	<script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client={{.GooglePublisherId}}" crossorigin="anonymous"></script>
	{{if .GoogleConsentNonce}}
	<script src="https://fundingchoicesmessages.google.com/i/{{.GooglePublisherId}}?ers=1" nonce="{{.GoogleConsentNonce}}"></script>
	{{end}}
{{else}}
	<!-- Ezoic Consent + Standalone -->
	<script src="https://cmp.gatekeeperconsent.com/min.js" data-cfasync="false"></script>
	<script src="https://the.gatekeeperconsent.com/cmp.min.js" data-cfasync="false"></script>
	<script>
		window.ezstandalone = window.ezstandalone || {};
		ezstandalone.cmd = ezstandalone.cmd || [];
	</script>
	<script async src="https://www.ezojs.com/ezoic/sa.min.js"></script>
{{end}}
{{end}}{{end}}

{{define "ads-css"}}{{if .ShowAds}}
		.content-row {
			display: flex;
			flex-direction: row;
			align-items: flex-start;
			justify-content: center;
			width: 100%;
			max-width: 1600px;
			overflow: hidden;
		}

		.ad-sidebar {
			flex-shrink: 0;
			width: 0;
			overflow: hidden;
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: flex-start;
			/* transform creates a new containing block for position:fixed children,
			   preventing Ezoic's fixed-position ads from escaping the sidebar bounds */
			transform: translateZ(0);
		}

		.ad-sidebar.visible { width: 160px; max-height: 620px; }
		.ad-sidebar.visible.wide { width: 300px; max-height: 620px; }

		.ad-below-content {
			display: none;
			width: 100%;
			max-width: 970px;
			margin-top: 16px;
			text-align: center;
		}

		.ad-below-content.visible { display: block; }
{{end}}{{end}}

{{define "ads-css-mobile"}}{{if .ShowAds}}
			.ad-sidebar {
				display: none !important;
			}
{{end}}{{end}}

{{define "file-content"}}
			{{if .IsImage}}
			<img src="{{.DirectURL}}" alt="{{.Filename}}" class="preview-image">
			{{else if .IsVideo}}
			<video controls preload="metadata" class="preview-video">
				<source src="{{.DirectURL}}" type="{{.MimeType}}">
				{{.T.VideoNotSupported}}
			</video>
			{{else if .IsAudio}}
			<audio controls preload="metadata" class="preview-audio">
				<source src="{{.DirectURL}}" type="{{.MimeType}}">
				{{.T.AudioNotSupported}}
			</audio>
			{{else if .IsPDF}}
			<iframe src="{{.DirectURL}}" class="preview-pdf"></iframe>
			{{else if .IsMarkdown}}
			<div id="markdown-body" class="preview-markdown"></div>
			<textarea id="markdown-source" style="display:none">{{.TextContent}}</textarea>
			<script src="https://cdn.jsdelivr.net/npm/marked@15/marked.min.js"></script>
			<script>
				(function() {
					var src = document.getElementById('markdown-source').value;
					marked.setOptions({breaks: true});
					document.getElementById('markdown-body').innerHTML = marked.parse(src);
				})();
			</script>
			{{else if .IsText}}
			<pre class="preview-text">{{.TextContent}}</pre>
			{{end}}
{{end}}

{{define "ads-scripts"}}{{if .ShowAds}}
{{if eq .AdProvider "google"}}
	<!-- Google Auto Ads: placement handled by Google -->
	<script>
	(function() {
		// Signal Google Funding Choices consent frame
		function signalGooglefcPresent() {
			if (!window.frames['googlefcPresent']) {
				if (document.body) {
					var iframe = document.createElement('iframe');
					iframe.style.cssText = 'width:0;height:0;border:none;z-index:-1000;left:-1000px;top:-1000px;display:none';
					iframe.name = 'googlefcPresent';
					document.body.appendChild(iframe);
				} else {
					setTimeout(signalGooglefcPresent, 0);
				}
			}
		}
		signalGooglefcPresent();
		if (typeof gtag !== 'undefined') {
			gtag('event', 'ads', {placement: 'google_auto'});
		}
	})();
	</script>
{{else}}
	<!-- Ezoic ad placement decision engine -->
	<script>
	(function() {
		'use strict';

		var MIN_SIDE_WIDTH    = 160;   // px each side minimum for skyscraper
		var WIDE_SIDE_WIDTH   = 300;   // px each side for half-page
		var MIN_SIDE_HEIGHT   = 400;   // px content height to justify side ads
		var MIN_BELOW_WIDTH   = 728;   // px viewport width for leaderboard
		var MIN_BELOW_HEIGHT  = 90;    // px vertical space needed below content
		var MOBILE_BREAKPOINT = 768;

		var sidebarLeft  = document.getElementById('ezoic-pub-ad-placeholder-118');
		var sidebarRight = document.getElementById('ezoic-pub-ad-placeholder-121');
		var belowDesktop = document.getElementById('ezoic-pub-ad-placeholder-119');
		var belowMobile  = document.getElementById('ezoic-pub-ad-placeholder-120');
		var mainEl       = document.querySelector('main');

		if (!mainEl) return;

		function getAvailableSpaceEachSide() {
			var mainRect = mainEl.getBoundingClientRect();
			var usableWidth = window.innerWidth - 40; // 20px padding each side
			return (usableWidth - mainRect.width) / 2;
		}

		function getAvailableSpaceBelow() {
			var mainRect = mainEl.getBoundingClientRect();
			return window.innerHeight - mainRect.bottom - 20;
		}

		function trackAd(placement) {
			if (typeof gtag !== 'undefined') {
				gtag('event', 'ads', {placement: placement});
			}
		}

		function showSideAds() {
			var availSide = getAvailableSpaceEachSide();
			var useWide = availSide >= WIDE_SIDE_WIDTH;
			if (sidebarLeft)  { sidebarLeft.classList.add('visible');  if (useWide) sidebarLeft.classList.add('wide');  }
			if (sidebarRight) { sidebarRight.classList.add('visible'); if (useWide) sidebarRight.classList.add('wide'); }
			ezstandalone.cmd.push(function() { ezstandalone.showAds(118, 121); });
			trackAd('side');
		}

		function showBelowDesktopAd() {
			if (belowDesktop) belowDesktop.classList.add('visible');
			ezstandalone.cmd.push(function() { ezstandalone.showAds(119); });
			trackAd('below');
		}

		function showMobileAd() {
			if (belowMobile) belowMobile.classList.add('visible');
			ezstandalone.cmd.push(function() { ezstandalone.showAds(120); });
			trackAd('mobile');
		}

		function tryBelowDesktopAd() {
			if (window.innerWidth >= MIN_BELOW_WIDTH && getAvailableSpaceBelow() >= MIN_BELOW_HEIGHT) {
				showBelowDesktopAd();
			} else {
				trackAd('none');
			}
		}

		function decide(orientation) {
			if (getAvailableSpaceBelow() < MIN_BELOW_HEIGHT) {
				ezstandalone.cmd.push(function() { ezstandalone.setEzoicAnchorAd(false); });
			}

			var isMobile = window.innerWidth <= MOBILE_BREAKPOINT;

			if (isMobile) {
				showMobileAd();
				return;
			}

			if (orientation === 'portrait') {
				var availSide = getAvailableSpaceEachSide();
				var mainRect = mainEl.getBoundingClientRect();
				if (availSide >= MIN_SIDE_WIDTH && mainRect.height >= MIN_SIDE_HEIGHT) {
					showSideAds();
				} else {
					tryBelowDesktopAd();
				}
			} else {
				// landscape or fullwidth
				tryBelowDesktopAd();
			}
		}

		{{if .IsImage}}
		(function() {
			var img = document.querySelector('.preview-image');
			if (!img) return;
			function onLoad() {
				if (!img.naturalWidth) return;
				mainEl.style.width = img.getBoundingClientRect().width + 'px';
				decide(img.naturalHeight > img.naturalWidth ? 'portrait' : 'landscape');
			}
			if (img.complete && img.naturalWidth > 0) { onLoad(); }
			else {
				img.addEventListener('load', onLoad);
				img.addEventListener('error', function() { tryBelowDesktopAd(); });
			}
		})();
		{{else if .IsVideo}}
		(function() {
			var vid = document.querySelector('.preview-video');
			if (!vid) return;
			function onMeta() {
				mainEl.style.width = vid.getBoundingClientRect().width + 'px';
				decide(vid.videoHeight > vid.videoWidth ? 'portrait' : 'landscape');
			}
			if (vid.readyState >= 1) { onMeta(); }
			else {
				vid.addEventListener('loadedmetadata', onMeta);
				vid.addEventListener('error', function() { tryBelowDesktopAd(); });
			}
		})();
		{{else}}
		// Audio, PDF, text, markdown: measure layout after paint
		function runAfterLayout() {
			requestAnimationFrame(function() {
				requestAnimationFrame(function() {
					decide('{{if .IsAudio}}portrait{{else}}fullwidth{{end}}');
				});
			});
		}
		if (document.readyState === 'loading') {
			document.addEventListener('DOMContentLoaded', runAfterLayout);
		} else {
			runAfterLayout();
		}
		{{end}}

	})();
	</script>
{{end}}
{{end}}{{end}}

{{define "ads-404-scripts"}}{{if .ShowAds}}
{{if eq .AdProvider "google"}}
	<!-- Google Auto Ads: placement handled by Google -->
	<script>
	if (typeof gtag !== 'undefined') {
		gtag('event', 'ads', {placement: '404_google_auto'});
	}
	</script>
{{else}}
	<script>
	ezstandalone.cmd.push(function() {
		var isMobile = window.innerWidth <= 768;
		var zoneId = isMobile ? 120 : 119;
		var el = document.getElementById('ezoic-pub-ad-placeholder-' + zoneId);
		if (el) el.closest('.ad-below-content').style.display = 'block';
		ezstandalone.showAds(zoneId);
		if (typeof gtag !== 'undefined') {
			gtag('event', 'ads', {placement: isMobile ? 'mobile_404' : 'below_404'});
		}
	});
	</script>
{{end}}
{{end}}{{end}}
`
