package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGetKafkaUserRequiresExactNameAcrossPages(t *testing.T) {
	firstPage := make([]string, 10)
	for i := range firstPage {
		firstPage[i] = "onboarding-forms-" + strconv.Itoa(i)
	}
	tests := []struct {
		name  string
		pages [][]string
		found bool
	}{
		{name: "exact follows partial match", pages: [][]string{{"onboarding-forms", "onboarding"}}, found: true},
		{name: "only partial match", pages: [][]string{{"onboarding-forms"}}, found: false},
		{name: "exact on next page", pages: [][]string{firstPage, {"onboarding"}}, found: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("userNames"); got != "onboarding" {
					t.Errorf("userNames = %q", got)
				}
				page := 1
				if value := r.URL.Query().Get("page"); value != "" {
					var err error
					page, err = strconv.Atoi(value)
					if err != nil {
						t.Errorf("page = %q: %v", value, err)
						return
					}
				}
				if page < 1 || page > len(tt.pages) {
					t.Errorf("unexpected page %d", page)
					return
				}
				users := make([]KafkaUserVO, 0, len(tt.pages[page-1]))
				for _, name := range tt.pages[page-1] {
					users = append(users, KafkaUserVO{Name: name})
				}
				totalPage := int64(len(tt.pages))
				if err := json.NewEncoder(w).Encode(PageNumResultKafkaUserVO{List: users, TotalPage: &totalPage}); err != nil {
					t.Errorf("encode response: %v", err)
				}
			}))
			defer server.Close()

			c, err := NewClient(context.Background(), server.URL, AuthCredentials{})
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(context.Background(), EnvIdKey, "env-test")
			user, err := c.GetKafkaUser(ctx, "kf-test", "onboarding")
			if tt.found {
				if err != nil || user == nil || user.Name != "onboarding" {
					t.Fatalf("GetKafkaUser = %v, %v; want onboarding", user, err)
				}
			} else if response, ok := err.(*ErrorResponse); !ok || response.Code != http.StatusNotFound || user != nil {
				t.Fatalf("GetKafkaUser = %v, %v; want 404", user, err)
			}
		})
	}
}
