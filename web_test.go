package main

import "testing"

func createTestRankedReviews() []rankedReview {
	revs := []rankedReview{}
	for i := 0; i < 10; i++ {
		curr := MovieReview{ReviewID: 0, MovieName: "test movie 1", Rating: 5}
		currRR := rankedReview{Review: curr, Rank: i}
		revs = append(revs, currRR)
	}
	return revs
}

var testReviews = createTestRankedReviews()

func Test_labelRankedReviews(t *testing.T) {
	tests := []struct {
		name    string
		reviews []rankedReview
		want    []rankedReview
	}{
		{
			name:    "1 review",
			reviews: []rankedReview{testReviews[0]},
			want: []rankedReview{
				{Review: testReviews[0].Review, Rank: testReviews[0].Rank, Label: ""},
			},
		},
		{
			name:    "3 review",
			reviews: testReviews[0:3],
			want: []rankedReview{
				{Review: testReviews[0].Review, Rank: testReviews[0].Rank, Label: "High"},
				{Review: testReviews[1].Review, Rank: testReviews[1].Rank, Label: "Mid"},
				{Review: testReviews[2].Review, Rank: testReviews[2].Rank, Label: "Low"},
			},
		},
		{
			name:    "5 review",
			reviews: testReviews[0:5],
			want: []rankedReview{
				{Review: testReviews[0].Review, Rank: testReviews[0].Rank, Label: "High"},
				{Review: testReviews[1].Review, Rank: testReviews[1].Rank, Label: "Mid"},
				{Review: testReviews[2].Review, Rank: testReviews[2].Rank, Label: "Mid"},
				{Review: testReviews[3].Review, Rank: testReviews[3].Rank, Label: "Mid"},
				{Review: testReviews[4].Review, Rank: testReviews[4].Rank, Label: "Low"},
			},
		},
		{
			name:    "6 review",
			reviews: testReviews[0:6],
			want: []rankedReview{
				{Review: testReviews[0].Review, Rank: testReviews[0].Rank, Label: "High"},
				{Review: testReviews[1].Review, Rank: testReviews[1].Rank, Label: "Mid"},
				{Review: testReviews[2].Review, Rank: testReviews[2].Rank, Label: "Mid"},
				{Review: testReviews[3].Review, Rank: testReviews[3].Rank, Label: "Mid"},
				{Review: testReviews[4].Review, Rank: testReviews[4].Rank, Label: "Mid"},
				{Review: testReviews[5].Review, Rank: testReviews[5].Rank, Label: "Low"},
			},
		},
		{
			name:    "10 review",
			reviews: testReviews,
			want: []rankedReview{
				{Review: testReviews[0].Review, Rank: testReviews[0].Rank, Label: "High"},
				{Review: testReviews[0].Review, Rank: testReviews[0].Rank, Label: "High"},
				{Review: testReviews[1].Review, Rank: testReviews[1].Rank, Label: "Mid"},
				{Review: testReviews[1].Review, Rank: testReviews[1].Rank, Label: "Mid"},
				{Review: testReviews[2].Review, Rank: testReviews[2].Rank, Label: "Mid"},
				{Review: testReviews[2].Review, Rank: testReviews[2].Rank, Label: "Mid"},
				{Review: testReviews[3].Review, Rank: testReviews[3].Rank, Label: "Mid"},
				{Review: testReviews[3].Review, Rank: testReviews[3].Rank, Label: "Mid"},
				{Review: testReviews[4].Review, Rank: testReviews[4].Rank, Label: "Low"},
				{Review: testReviews[4].Review, Rank: testReviews[4].Rank, Label: "Low"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labelRankedReviews(tt.reviews)
			for i, v := range got {
				if v.Label != tt.want[i].Label {
					t.Errorf("labelRankedReviews() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
