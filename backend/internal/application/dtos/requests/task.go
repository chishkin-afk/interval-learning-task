package requests

type CreateTask struct {
	Title       string `json:"title"`
	LeetcodeURL string `json:"leetcode_url"`
}

type UpdateTask struct {
	Title       *string `json:"title"`
	LeetcodeURL *string `json:"leetcode_url"`
}
