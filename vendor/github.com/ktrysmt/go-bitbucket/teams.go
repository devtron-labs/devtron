package bitbucket

type Teams struct {
	c *Client
}

func (t *Teams) List(role string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/?role=%s", role)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Profile(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Members(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/members", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Followers(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/followers", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Following(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/following", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Repositories(teamname string) (interface{}, error) {
	urlStr := t.c.requestUrl("/teams/%s/repositories", teamname)
	return t.c.execute("GET", urlStr, "")
}

func (t *Teams) Projects(teamname string) (interface{}, error) {
        urlStr := t.c.requestUrl("/teams/%s/projects/", teamname)
        return t.c.execute("GET", urlStr, "")
}

