package connector

import (
	"fmt"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {
	t.Run("isLastPage", func(t *testing.T) {
		testCases := []struct {
			count    int
			pageSize int
			expected bool
		}{
			{1, 1, false},
			{0, 1, true},
			{2, 1, false},
		}
		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("%d %d", testCase.count, testCase.pageSize), func(t *testing.T) {
				assert.Equal(t, testCase.expected, isLastPage(testCase.count, testCase.pageSize))
			})
		}
	})

	t.Run("getPageFromPageToken", func(t *testing.T) {
		testCases := []struct {
			pageToken string
			expected  int
		}{
			{"", 0},
			{"0", 0},
			{"1", 1},
		}
		for _, testCase := range testCases {
			t.Run(testCase.pageToken, func(t *testing.T) {
				page, err := getPageFromPageToken(testCase.pageToken)
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, page)
			})
		}
	})

	t.Run("parsePageToken", func(t *testing.T) {
		testCases := []struct {
			token     string
			resources []*v2.ResourceId
			expected  int
		}{
			{"", nil, 0},
			{`{"current_state": {"token": "2"}}`, nil, 2},
			{
				`{"current_state": {"token": "2"}}`,
				[]*v2.ResourceId{
					{
						ResourceType: "",
						Resource:     "",
					},
				},
				2,
			},
			{
				"",
				[]*v2.ResourceId{
					{
						ResourceType: "",
						Resource:     "",
					},
				},
				0,
			},
		}
		for _, testCase := range testCases {
			t.Run(testCase.token, func(t *testing.T) {
				_, page, err := parsePageToken(testCase.token, testCase.resources...)
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, page)
			})
		}
	})

	t.Run("getPageTokenFromPage", func(t *testing.T) {
		testCases := []struct {
			nextPage int
			expected string
		}{
			{
				0,
				`{"states":null,"current_state":{"token":"0"}}`,
			},
			{
				10,
				`{"states":null,"current_state":{"token":"10"}}`,
			},
		}
		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("%d", testCase.nextPage), func(t *testing.T) {
				bag := &pagination.Bag{}
				bag.Push(pagination.PageState{})
				token, err := getPageTokenFromPage(bag, testCase.nextPage)
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, token)
				assert.Equal(t, fmt.Sprintf("%d", testCase.nextPage), bag.Current().Token)
			})
		}

		t.Run("should error on empty bag", func(t *testing.T) {
			pageToken, err := getPageTokenFromPage(&pagination.Bag{}, 0)

			assert.NotNil(t, err)
			assert.Equal(t, "", pageToken)
		})
	})
}
