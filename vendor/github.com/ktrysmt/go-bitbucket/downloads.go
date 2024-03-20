package bitbucket

type Downloads struct {
	c *Client
}

func (dl *Downloads) Create(do *DownloadsOptions) (interface{}, error) {
	urlStr := dl.c.requestUrl("/repositories/%s/%s/downloads", do.Owner, do.RepoSlug)
	return dl.c.executeFileUpload("POST", urlStr, do.FilePath, do.FileName, "files", make(map[string]string))
}

func (dl *Downloads) List(do *DownloadsOptions) (interface{}, error) {
	urlStr := dl.c.requestUrl("/repositories/%s/%s/downloads", do.Owner, do.RepoSlug)
	return dl.c.executePaginated("GET", urlStr, "", nil)
}
