package models

import "testing"

func TestGetAllArticles(t *testing.T) {
	aList := GetAllArticles()

	if len(aList) != len(ArticleList) {
		t.Fail()
	}

	for i, a := range aList {
		if a.Content != ArticleList[i].Content ||
			a.ID != ArticleList[i].ID ||
			a.Title != ArticleList[i].Title {
			t.Fail()
			break
		}
	}
}
