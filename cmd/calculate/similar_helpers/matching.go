package similar_helpers

import (
	"github.com/similar-manga/similar/internal"
	"strings"
)

var oneWayTags = []string{
	"b11fda93-8f1d-4bef-b2ed-8803d3733170", // 4-Koma
	"b13b2a48-c720-44a9-9c77-39c9979373fb", //Doujinshi
	"b29d6a3d-1569-4e7a-8caf-7557bc92cd5d", //Gore
	"97893a4c-12af-4dac-b6be-0dffb353568e", //Sexual Violence
	"5920b825-4181-4a17-beeb-9918b0ff7a30", //Boys' Love
	"a3c67850-4684-404e-9b7f-c69850ee5da6", //"Girls' Love
	"acc803a4-c95a-4c22-86fc-eb6b582d82a2", //Wuxia
	"2d1f5d56-a1e5-4d0d-a961-2193588b08ec", //Loli
	"ddefd648-5140-4e5f-ba18-4eca4071d19b", //Shota
	"5bd0e105-4481-44ca-b6e7-7544da56b1a3", //Incest
}

func NotValidMatch(manga internal.Manga, mangaOther internal.Manga) bool {

	// Enforce that the two do not have another as a *related* manga
	for _, relatedId := range manga.RelatedIds {
		if relatedId == mangaOther.Id {
			return true
		}
	}

	for _, relatedId := range mangaOther.RelatedIds {
		if relatedId == manga.Id {
			return true
		}
	}

	// Enforce that our two demographics are the same
	if manga.ContentRating != "" &&
		manga.ContentRating != mangaOther.ContentRating {
		return true
	}

	// Small check for "promo" titles, don't match to promotional titles
	title := strings.ToLower((*manga.Title)["en"])
	titleOther := strings.ToLower((*mangaOther.Title)["en"])
	if !strings.Contains(title, "(promo)") && strings.Contains(titleOther, "(promo)") {
		return true
	}

	// Enforce that our two demographics are the same
	if manga.PublicationDemographic != "" &&
		manga.PublicationDemographic != mangaOther.PublicationDemographic {
		return true
	}

	// No need to check tags for our top level content ratings
	// They will be a valid match no matter the tags (not that many options thus can't limit)
	if manga.ContentRating == "erotica" || manga.ContentRating == "pornographic" {
		return false
	}

	// Next we should enforce the following tags
	for _, tagId := range oneWayTags {

		// Check to see if this tag is in our first manga
		hasTag := false
		for _, currentMangaTag := range manga.Tags {
			if currentMangaTag.Id == tagId {
				hasTag = true
				break
			}
		}

		// If we have the tag, then no need to check the other manga
		// If we don't have it, then the other manga shouldn't have it..
		if hasTag {
			continue
		}

		// Check if other does not have the tag
		for _, otherMangaTag := range mangaOther.Tags {
			if otherMangaTag.Id == tagId {
				return true
			}
		}

	}

	// Else this is a valid match we can use!
	return false

}
