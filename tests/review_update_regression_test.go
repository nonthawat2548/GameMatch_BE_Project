package tests

import (
	"context"
	"gamematch/store"
	"strconv"
	"testing"
)

func TestReviewUpdatePreservesHelpfulCountAndOwnership(t *testing.T) {
	a, server := databaseAPI(t)
	ownerID := registerMember(t, server, "edit-owner@example.test")
	registerMember(t, server, "edit-reader@example.test")
	owner := loginMember(t, server, "edit-owner@example.test", "Password123!")
	reader := loginMember(t, server, "edit-reader@example.test", "Password123!")
	gameID, err := a.S.Save(context.Background(), store.Game{Title: "Review update fixture", Published: true})
	if err != nil {
		t.Fatal(err)
	}
	base := "/api/games/" + strconv.FormatInt(gameID, 10) + "/reviews"
	var created store.Review
	requestJSON(t, server, "POST", base, owner, map[string]any{"rating": 5, "comment": "Original review before edits"}, 201, &created)
	path := base + "/" + strconv.FormatInt(created.ID, 10)
	edit := map[string]any{"rating": 4, "comment": "Owner edit before helpful votes"}
	var updated store.Review
	requestJSON(t, server, "PUT", path, owner, edit, 200, &updated)
	if updated.HelpfulCount != 0 || updated.ID != created.ID || updated.MemberID != ownerID {
		t.Fatal("editing an unvoted review returned an incorrect count or owner")
	}
	requestJSON(t, server, "PUT", path+"/helpful", reader, nil, 200, nil)
	edit = map[string]any{"rating": 3, "comment": "Owner edit after helpful votes"}
	requestJSON(t, server, "PUT", path, owner, edit, 200, &updated)
	if updated.HelpfulCount != 1 || updated.Rating != 3 || updated.Comment != edit["comment"] || updated.MemberID != ownerID {
		t.Fatal("editing a voted review returned an incorrect count, content, or owner")
	}
	requestJSON(t, server, "PUT", path, reader, map[string]any{"rating": 1, "comment": "Another member must not replace this review"}, 404, nil)
	var reviews []store.Review
	requestJSON(t, server, "GET", base, "", nil, 200, &reviews)
	if len(reviews) != 1 || reviews[0].HelpfulCount != updated.HelpfulCount || reviews[0].Rating != updated.Rating || reviews[0].Comment != updated.Comment || reviews[0].MemberID != ownerID {
		t.Fatal("GET differs from the update response or a different member changed the review")
	}
}
