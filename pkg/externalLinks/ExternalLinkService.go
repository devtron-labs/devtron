package externalLinks

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

type Linkout2 struct {
	Id                 int      `json:"id"`
	Name               string   `json:"name"`
	Url                string   `json:"url"`
	Monitoring_tool_id int      `json:"monitoring_tool_id"`
	Clusters           []string `json:"clusters_ids"`
}

type Clusters struct {
	gorm.Model
	Id         int    `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Linkout_id int    `json:"linkout_id"`
	Cluster_id string `json:"cluster_id"`
}

func CreateLink(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Creating New Link....")
	w.Header().Set("Content-Type", "application/json")
	var link Linkout2
	json.NewDecoder(r.Body).Decode(&link)
	DB.Table("linkouts").Omit("Clusters").Create(&link)

	for _, v := range link.Clusters {
		var temp Clusters
		temp.Linkout_id = link.Id
		temp.Cluster_id = v
		DB.Table("clusters").Create(&temp)
	}

	json.NewEncoder(w).Encode(link)
}
