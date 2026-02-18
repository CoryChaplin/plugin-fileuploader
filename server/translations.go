package server

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Translations holds all user-visible strings for a given language
type Translations struct {
	Lang string // HTML lang attribute: "fr", "en"

	// Duration unit words used by humanizeDuration
	HourSingular string
	HourPlural   string
	DaySingular  string
	DayPlural    string

	// File view page
	SharedFile        string // og:description prefix
	VideoNotSupported string // <video> fallback text
	AudioNotSupported string // <audio> fallback text

	// 404 page
	PageTitleNotFound string
	NotFoundHeading   string
	FilesKeptFor      string // "Les fichiers sont conservés" / "Files are kept for"
	ForConnected      string // "pour les utilisateurs connectés" / "for connected users"
	LinkExpired       string
	ReturnToChat      string
}

var translations = map[string]*Translations{
	"fr": {
		Lang:              "fr",
		HourSingular:      "heure",
		HourPlural:        "heures",
		DaySingular:       "jour",
		DayPlural:         "jours",
		SharedFile:        "Fichier partagé",
		VideoNotSupported: "Votre navigateur ne supporte pas la lecture de vidéos.",
		AudioNotSupported: "Votre navigateur ne supporte pas la lecture audio.",
		PageTitleNotFound: "Fichier introuvable",
		NotFoundHeading:   "Ce fichier n'est plus disponible",
		FilesKeptFor:      "Les fichiers sont conservés",
		ForConnected:      "pour les utilisateurs connectés",
		LinkExpired:       "Ce lien a probablement expiré.",
		ReturnToChat:      "Retourner sur le chat",
	},
	"en": {
		Lang:              "en",
		HourSingular:      "hour",
		HourPlural:        "hours",
		DaySingular:       "day",
		DayPlural:         "days",
		SharedFile:        "Shared file",
		VideoNotSupported: "Your browser does not support video playback.",
		AudioNotSupported: "Your browser does not support audio playback.",
		PageTitleNotFound: "File not found",
		NotFoundHeading:   "This file is no longer available",
		FilesKeptFor:      "Files are kept for",
		ForConnected:      "for connected users",
		LinkExpired:       "This link has probably expired.",
		ReturnToChat:      "Return to chat",
	},
}

var defaultTranslations = translations["fr"]

// detectLanguage parses an Accept-Language header and returns the best
// matching Translations, falling back to French if no match is found.
func detectLanguage(acceptLanguage string) *Translations {
	type langPref struct {
		lang string
		q    float64
	}

	var prefs []langPref
	for _, part := range strings.Split(acceptLanguage, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		q := 1.0
		if idx := strings.Index(part, ";q="); idx >= 0 {
			if v, err := strconv.ParseFloat(strings.TrimSpace(part[idx+3:]), 64); err == nil {
				q = v
			}
			part = part[:idx]
		}
		// Take primary subtag only (e.g. "fr" from "fr-FR")
		primary := strings.ToLower(strings.TrimSpace(strings.SplitN(part, "-", 2)[0]))
		if primary != "" && primary != "*" {
			prefs = append(prefs, langPref{primary, q})
		}
	}

	sort.SliceStable(prefs, func(i, j int) bool { return prefs[i].q > prefs[j].q })

	for _, p := range prefs {
		if t, ok := translations[p.lang]; ok {
			return t
		}
	}
	return defaultTranslations
}

// humanizeDuration converts a duration to a human-readable string in the
// given language (e.g. "24 heures", "7 jours", "24 hours", "7 days").
func humanizeDuration(d time.Duration, t *Translations) string {
	hours := int(d.Hours())
	if hours >= 24 && hours%24 == 0 {
		days := hours / 24
		if days == 1 {
			return fmt.Sprintf("1 %s", t.DaySingular)
		}
		return fmt.Sprintf("%d %s", days, t.DayPlural)
	}
	if hours == 1 {
		return fmt.Sprintf("1 %s", t.HourSingular)
	}
	return fmt.Sprintf("%d %s", hours, t.HourPlural)
}
