package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	_ "io/ioutil"
	"net/http"
)

const (
	host     = "192.168.163.196"
	port     = 5432
	user     = "gitlab"
	password = "Aa111111"
	dbname   = "gitlabhq_production"
)

var Db *sql.DB

type hook struct {
	ObjectKind       string `json:"object_kind"`
	ObjectAttributes struct {
		Id              int    `json:"id"`
		TargetBranch    string `json:"target_branch"`
		SourceBranch    string `json:"source_branch"`
		SourceProjectId int    `json:"source_project_id"`
		TargetProjectId int    `json:"target_project_id"`
		AuthorId        int    `json:"author_id"`
		State           string `json:"state"`
		MergeStatus     string `json:"merge_status"`
		Iid             int    `json:"iid"`
	} `json:"object_attributes"`
	MergeRequest struct {
		Iid int `json:"iid"`
	} `json:"merge_request"`
}

func main() {
	InitDb()
	router := httprouter.New()
	router.POST("/", Handle)
	err := http.ListenAndServe(":3000", router)
	if err != nil {
		panic("listen and serve:")
	}
}

func Handle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	/*	read, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(read))
	*/
	var h hook
	err := json.NewDecoder(r.Body).Decode(&h)

	if err != nil {
		panic(err)
	}

	fmt.Println(h.ObjectKind)
	fmt.Println(h.ObjectAttributes.Id)
	fmt.Println(h)
	id, err := CheckStatus(h)
	if err != nil && err.Error() == "sql: no rows in result set" {
		fmt.Println("No ROws!!")
	} else if err != nil {
		panic(err)
	}
	fmt.Println("working id:", id)

	/*
		read, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(read))
	*/
}

func InitDb() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	Db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	//defer Db.Close()
	err = Db.Ping()
	if err != nil {
		fmt.Println("Connection to database Failed:", err)
	} else {
		fmt.Println("Connection to database Successful")
	}

}

//Check the status of Merge Request, Im Processing Open and Not Merged only.
func CheckStatus(h hook) (int, error) {
	fmt.Println("beforequery:", h.ObjectAttributes.TargetProjectId, h.ObjectAttributes.Iid)
	row := Db.QueryRow(`SELECT id FROM merge_requests WHERE target_project_id = $1 AND iid = $2 AND state != 'closed' AND state != 'merged'`, h.ObjectAttributes.TargetProjectId, h.ObjectAttributes.Iid)
	var id int
	err := row.Scan(&id)
	if err != nil {
		fmt.Println("Error in select:", err)
		return 0, err
	} else {
		return id, nil
	}
}
