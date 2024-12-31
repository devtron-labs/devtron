package bitbucket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
)

type Diff struct {
	c *Client
}

type DiffStatRes struct {
	Page      int         `json:"page,omitempty"`
	Pagelen   int         `json:"pagelen,omitempty"`
	Size      int         `json:"size,omitempty"`
	Next      string      `json:"next,omitempty"`
	Previous  string      `json:"previous,omitempty"`
	DiffStats []*DiffStat `json:"values,omitempty"`
}

type DiffStat struct {
	Type         string                 `json:"type,omitempty"`
	Status       string                 `json:"status,omitempty"`
	LinesRemoved int                    `json:"lines_removed,omitempty"`
	LinedAdded   int                    `json:"lines_added,omitempty"`
	Old          map[string]interface{} `json:"old,omitempty"`
	New          map[string]interface{} `json:"new,omitempty"`
}

func (d *Diff) GetDiff(do *DiffOptions) (interface{}, error) {
	urlStr := d.c.requestUrl("/repositories/%s/%s/diff/%s", do.Owner, do.RepoSlug, do.Spec)
	return d.c.executeRaw("GET", urlStr, "diff")
}

func (d *Diff) GetPatch(do *DiffOptions) (interface{}, error) {
	urlStr := d.c.requestUrl("/repositories/%s/%s/patch/%s", do.Owner, do.RepoSlug, do.Spec)
	return d.c.executeRaw("GET", urlStr, "")
}

func (d *Diff) GetDiffStat(dso *DiffStatOptions) (*DiffStatRes, error) {

	params := url.Values{}
	if dso.Whitespace == true {
		params.Add("ignore_whitespace", strconv.FormatBool(dso.Whitespace))
	}

	if dso.Merge == false {
		params.Add("merge", strconv.FormatBool(dso.Merge))
	}

	if dso.Path != "" {
		params.Add("path", dso.Path)
	}

	if dso.Renames == false {
		params.Add("renames", strconv.FormatBool(dso.Renames))
	}

	if dso.PageNum > 0 {
		params.Add("page", strconv.Itoa(dso.PageNum))
	}

	if dso.Pagelen > 0 {
		params.Add("pagelen", strconv.Itoa(dso.Pagelen))
	}

	if dso.MaxDepth > 0 {
		params.Add("max_depth", strconv.Itoa(dso.MaxDepth))
	}

	if len(dso.Fields) > 0 {
		params.Add("fields", cleanFields(dso.Fields))
	}

	urlStr := d.c.requestUrl("/repositories/%s/%s/diffstat/%s?%s", dso.Owner, dso.RepoSlug,
		dso.Spec,
		params.Encode())
	response, err := d.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}
	bodyString := string(bodyBytes)
	return decodeDiffStat(bodyString)
}

func decodeDiffStat(diffStatResponseStr string) (*DiffStatRes, error) {

	var diffStatRes DiffStatRes

	err := json.Unmarshal([]byte(diffStatResponseStr), &diffStatRes)
	if err != nil {
		return nil, fmt.Errorf("DiffStat decode error: %w", err)
	}

	return &diffStatRes, nil
}

// cleanFields combines all query params in the slice of field strigs into a sigle string
// and removes any whitespace before returing the string.
func cleanFields(fields []string) string {
	interS := strings.Join(fields, ",")
	s := strings.ReplaceAll(interS, " ", "")
	return s
}
