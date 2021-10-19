package bitbucket

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

type Diff struct {
	c *Client
}

type DiffStatRes struct {
	Page      int
	Pagelen   int
	MaxDepth  int
	Size      int
	Next      string
	DiffStats []DiffStat
}

type DiffStat struct {
	Type         string
	Status       string
	LinesRemoved int
	LinedAdded   int
	Old          map[string]interface{}
	New          map[string]interface{}
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

	var diffStatResponseMap map[string]interface{}
	err := json.Unmarshal([]byte(diffStatResponseStr), &diffStatResponseMap)
	if err != nil {
		return nil, err
	}

	diffStatArray := diffStatResponseMap["values"].([]interface{})
	var diffStatsSlice []DiffStat
	for _, diffStatEntry := range diffStatArray {
		var diffStat DiffStat
		err = mapstructure.Decode(diffStatEntry, &diffStat)
		if err == nil {
			diffStatsSlice = append(diffStatsSlice, diffStat)
		}
	}

	page, ok := diffStatResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := diffStatResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}

	max_depth, ok := diffStatResponseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}

	size, ok := diffStatResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := diffStatResponseMap["next"].(string)
	if !ok {
		next = ""
	}

	diffStats := DiffStatRes{
		Page:      int(page),
		Pagelen:   int(pagelen),
		MaxDepth:  int(max_depth),
		Size:      int(size),
		Next:      next,
		DiffStats: diffStatsSlice,
	}
	return &diffStats, nil
}
