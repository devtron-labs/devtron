package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initializeRouter() {
	r := mux.NewRouter()

	//r.HandleFunc("/users/tools", GetTools).Methods("GET")
	//r.HandleFunc("/users", GetLinks).Methods("GET")
	r.HandleFunc("/users", CreateLink).Methods("POST")
	//r.HandleFunc("/users", UpdateLink).Methods("PUT")
	//r.HandleFunc("/users", DeleteLink).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":4000", r))
}

func main() {
	InitialMigration()
	initializeRouter()
}

var DB *gorm.DB
var err error

type Linkout struct {
	gorm.Model
	Id                 int    `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Name               string `gorm:"type:varchar(100);not null" json:"name"`
	Url                string `gorm:"type:varchar(100);not null" json:"url"`
	Monitoring_tool_id int    `gorm:"not null" json:"monitoring_tool_id"`
	Is_active          bool   `gorm:"not null;default:true" json:"is_active"`
}

type Clusters struct {
	gorm.Model
	Id         int    `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	Linkout_id int    `json:"linkout_id"`
	Cluster_id string `json:"cluster_id"`
}

const DNS = "root:@aRyaN777@tcp(127.0.0.1:3306)/golang?charset=utf8mb4&parseTime=True&loc=Local"

func InitialMigration() {
	DB, err = gorm.Open(mysql.Open(DNS), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		panic("Cannot connect to DB")
	}
	DB.AutoMigrate(&Clusters{})
	DB.AutoMigrate(&Linkout{})
}

type Linkout2 struct {
	Id                 int      `json:"id"`
	Name               string   `json:"name"`
	Url                string   `json:"url"`
	Monitoring_tool_id int      `json:"monitoring_tool_id"`
	Clusters           []string `json:"clusters_ids"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Creating New Link....")
	w.Header().Set("Content-Type", "application/json")
	var user Linkout2
	json.NewDecoder(r.Body).Decode(&user)
	DB.Table("linkouts").Omit("Clusters").Create(&user)

	for _, v := range user.Clusters {
		var temp Clusters
		temp.Linkout_id = user.Id
		temp.Cluster_id = v
		DB.Table("clusters").Create(&temp)
	}

	json.NewEncoder(w).Encode(user)

	fmt.Println("Last inserted id is.: ", user.Id)
}
