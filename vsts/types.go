package vsts

type author struct {
	ID string `json:"id"`
}

type comment struct {
	ID              int    `json:"id"`
	ParentCommentID int    `json:"parentCommentId"`
	Author          author `json:"author"`
	Content         string `json:"content"`
	CommentType     string `json:"commentType"`
	IsDeleted       bool   `json:"isDeleted"`
}

type filePosition struct {
	Line   int `json:"line"`
	Offset int `json:"offset"`
}

type threadContext struct {
	FilePath       string       `json:"filePath"`
	RightFileStart filePosition `json:"rightFileStart"`
	RightFileEnd   filePosition `json:"rightFileEnd"`
}

type commentThread struct {
	ID            int           `json:"id"`
	Comments      []comment     `json:"comments"`
	Status        string        `json:"status"`
	ThreadContext threadContext `json:"threadContext"`
	IsDeleted     bool          `json:"isDeleted"`
}

type commentThreads struct {
	Value []commentThread `json:"value"`
	Count int             `json:"count"`
}

type postComment struct {
	ParentCommentID int    `json:"parentCommentId"`
	Content         string `json:"content"`
	CommentType     int    `json:"commentType"`
}

type supportsMarkDown struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type threadProperty struct {
	MicrosoftTeamFoundationDiscussionSupportsMarkdown supportsMarkDown `json:"Microsoft.TeamFoundation.Discussion.SupportsMarkdown"`
}

type postThread struct {
	Comments      []postComment  `json:"comments"`
	Properties    threadProperty `json:"properties"`
	Status        int            `json:"status"`
	ThreadContext threadContext  `json:"threadContext"`
}

type patchThread struct {
	Status int `json:"status"`
}

type putVote struct {
	Vote int `json:"vote"`
}

type diffs struct {
	AllChangesIncluded bool `json:"allChangesIncluded"`
	ChangeCounts       struct {
		Edit int `json:"Edit"`
	} `json:"changeCounts"`
	Changes []struct {
		Item struct {
			ObjectID         string `json:"objectId"`
			OriginalObjectID string `json:"originalObjectId"`
			GitObjectType    string `json:"gitObjectType"`
			CommitID         string `json:"commitId"`
			Path             string `json:"path"`
			IsFolder         bool   `json:"isFolder"`
			URL              string `json:"url"`
		} `json:"item"`
		ChangeType string `json:"changeType"`
	} `json:"changes"`
	CommonCommit string `json:"commonCommit"`
	BaseCommit   string `json:"baseCommit"`
	TargetCommit string `json:"targetCommit"`
	AheadCount   int    `json:"aheadCount"`
	BehindCount  int    `json:"behindCount"`
}

type imageList struct {
	Version string   `json:"version"`
	Common  []string `json:"common"`
}
