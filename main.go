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

	var h hook
	err := json.NewDecoder(r.Body).Decode(&h)

	if err != nil {
		panic(err)
	}

	fmt.Println(h.ObjectKind)
	fmt.Println(h.ObjectAttributes.Id)
	fmt.Println(h)

	id, err := CheckStatus(h)
	if err != nil {
		panic(err)
	}
	fmt.Prinln("working id:", id)

	/*
		read, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(read))
	*/
}

func InitDb() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	Db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer Db.Close()
	err = Db.Ping()
	if err != nil {
		fmt.Println("Connection to database Failed:", err)
	} else {
		fmt.Println("Connection to database Successful")
	}

}

func CheckStatus(h hook) (int, error) {
	Db.QueryRow("SELECT m.id FROM merge_requests AS m WHERE m.target_project_id = $1 AND m.iid = $2 AND m.state != 'closed' AND m.state != 'merged'", h.ObjectAttributes.TargetProjectId, h.MergeRequest.Iid)
	var id int
	err := row.Scan(&id)
	if err != nil {
		fmt.Println("Error in select:", err)
		return 0, err
	} else {
		return id, nil
	}
}
