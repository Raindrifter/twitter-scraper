package twitterscraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Global cache for user IDs
var cacheIDs sync.Map

// Profile of twitter user.
type Profile struct {
	Avatar         string
	Banner         string
	Biography      string
	Birthday       string
	FollowersCount int
	FollowingCount int
	FriendsCount   int
	IsPrivate      bool
	IsVerified     bool
	Joined         *time.Time
	LikesCount     int
	ListedCount    int
	Location       string
	Name           string
	PinnedTweetIDs []string
	TweetsCount    int
	URL            string
	UserID         string
	Username       string
	Website        string
	Sensitive      bool
}

type user struct {
	Data struct {
		User struct {
			RestID string     `json:"rest_id"`
			Legacy legacyUser `json:"legacy"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// GetProfile return parsed user profile.
func (s *Scraper) GetProfile(username string) (Profile, error) {
	var jsn user
	req, err := http.NewRequest("GET", "https://api.twitter.com/graphql/4S2ihIKfF3xhp-ENxvUAfQ/UserByScreenName", nil)
	if err != nil {
		return Profile{}, err
	}

	variables := map[string]interface{}{
		"screen_name":          username,
		"withHighlightedLabel": true,
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	req.URL.RawQuery = query.Encode()

	err = s.RequestAPI(req, &jsn)
	if err != nil {
		return Profile{}, err
	}

	if len(jsn.Errors) > 0 {
		return Profile{}, fmt.Errorf("%s", jsn.Errors[0].Message)
	}

	if jsn.Data.User.RestID == "" {
		return Profile{}, fmt.Errorf("rest_id not found")
	}
	jsn.Data.User.Legacy.IDStr = jsn.Data.User.RestID

	if jsn.Data.User.Legacy.ScreenName == "" {
		return Profile{}, fmt.Errorf("either @%s does not exist or is private", username)
	}

	return parseProfile(jsn.Data.User.Legacy), nil
}

// GetFansByUserID gets fans for a given userID, via the Twitter frontend GraphQL API.
func (s *Scraper) GetFansByUserID(userId string, maxTweetsNbr int, cursor string) ([]*Profile, string, error) {
	if maxTweetsNbr > 200 {
		maxTweetsNbr = 200
	}

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/3_7xfjmh897x8h_n6QBqTA/Followers")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"userId":                                 userId,
		"count":                                  maxTweetsNbr,
		"includePromotedContent":                 false,
		"withQuickPromoteEligibilityTweetFields": false,
		// "withV2Timeline":                         false,
	}
	features := map[string]interface{}{
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                false,
		"tweet_awards_web_tipping_enabled":                                        false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"rweb_video_timestamps_enabled":                                           true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_media_download_video_enabled":                             false,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	var timeline timelineV2
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}
	// var result map[string]interface{}
	// jsoniter.Get(timeline, "data", "user").ToVal(&result)

	var nextCursor = ""
	users, nextCursor := timeline.parseUsers()
	return users, nextCursor, nil
}

// GetFansByUserID gets fans for a given userID, via the Twitter frontend GraphQL API.
func (s *Scraper) GetFollowingByUserID(userId string, maxTweetsNbr int, cursor string) ([]*Profile, string, error) {
	if maxTweetsNbr > 200 {
		maxTweetsNbr = 200
	}

	req, err := s.newRequest("GET", "https://twitter.com/i/api/graphql/g5P4cbXR4ta4oCeE7y2vLQ/Following")
	if err != nil {
		return nil, "", err
	}

	variables := map[string]interface{}{
		"userId":                 userId,
		"count":                  maxTweetsNbr,
		"includePromotedContent": false,
		"withV2Timeline":         true,
		// "withQuickPromoteEligibilityTweetFields": false,
	}
	features := map[string]interface{}{
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                false,
		"tweet_awards_web_tipping_enabled":                                        false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"rweb_video_timestamps_enabled":                                           true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_media_download_video_enabled":                             false,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	fmt.Println(cursor)
	cnumber, _ := strconv.ParseInt(cursor, 10, 64)
	if cursor != "" {
		variables["cursor"] = cnumber
	}

	query := url.Values{}
	query.Set("variables", mapToJSONString(variables))
	query.Set("features", mapToJSONString(features))
	req.URL.RawQuery = query.Encode()

	// var result []byte
	var timeline timelineV2
	err = s.RequestAPI(req, &timeline)
	if err != nil {
		return nil, "", err
	}
	// fmt.Println(string(result))
	// jsoniter.Unmarshal(result, &timeline)

	var nextCursor = ""
	users, nextCursor := timeline.parseUsers()
	return users, nextCursor, nil
}

// GetUserIDByScreenName from API
func (s *Scraper) GetUserIDByScreenName(screenName string) (string, error) {
	id, ok := cacheIDs.Load(screenName)
	if ok {
		return id.(string), nil
	}

	profile, err := s.GetProfile(screenName)
	if err != nil {
		return "", err
	}

	cacheIDs.Store(screenName, profile.UserID)

	return profile.UserID, nil
}
