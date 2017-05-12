package webhook

// PushHook defines a hook body for a push request.
type PushHook struct {
	Ref        string                 `json:"ref"`
	Before     string                 `json:"before"`
	After      string                 `json:"after"`
	HeadCommit Commit                 `json:"head_commit"`
	Commits    []Commit               `json:"commits"`
	Repository Repository             `json:"repository"`
	Pusher     User                   `json:"pusher"`
	Sender     map[string]interface{} `json:"sender"`
}

// Commit is a record of a commit
type Commit struct {
	Id        string `json:"id"`
	TreeId    string `json:"tree_id"`
	Distinct  bool   `json:"distinct"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`

	// SHA is documented, but not shown in examples
	SHA string `json:"sha"`

	Author   User `json:"author"`
	Commiter User `json:"committer"`

	Added    []string `json:"added"`
	Removed  []string `json:"remove"`
	Modified []string `json:"modified"`
}

// User is a GitHub user record.
type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Repository struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Owner       User   `json:"owner"`
	Description string `json:"description"`

	Private bool `json:"private"`
	Fork    bool `json:"fork"`

	// URLs:

	URL      string `json:"url"`
	HTMLURL  string `json:"html_url"`
	GitURL   string `json:"git_url"`
	SSHURL   string `json:"ssh_url"`
	CloneURL string `json:"clone_url"`
	Homepage string `json:"homepage"`

	// There are about a dozen other URLs that we could capture

	// Also, there's info on stargazers, watchers, issues, etc.

}
