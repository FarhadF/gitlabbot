package cmd

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

/*const (
	dbhost        = "192.168.163.196"
	dbport        = 5432
	dbuser        = "gitlab"
	dbpassword    = "Aa111111"
	dbname        = "gitlabhq_production"
	gitlabBase    = "http://192.168.163.196:10080"
	gitlabToken   = "K8F8SZEHyq4Dm9osdTT3"
	lgtmTreashold = 2
)
*/
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
	ProjectId int `json:"project_id"`
}

func gitlabbot(dbhost string, dbport int, dbuser string, dbpassword string, dbname string, gitlabBase string, gitlabToken string, lgtmTreashold int) {
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
	if h.ObjectKind == "merge_request" {
		fmt.Println("object is merge request")
		id, err := CheckStatus(h)
		if err != nil && err.Error() == "sql: no rows in result set" {
			fmt.Println("No ROws!!")
		} else if err != nil {
			panic(err)
		}
		fmt.Println("working id:", id)

		count, err := CheckInitial(h)
		if err == nil && count != 0 {
			fmt.Println("Number Of Comments:", count)
			lgtms, err := CheckLGTM(h)
			if err != nil {
				panic(err)
			}
			fmt.Println("Number of LGTMs:", lgtms)

		} else if err == nil && count == 0 {
			InitialComment(h)
			//	if err != nil {
			//		panic(err)
			//	}
		} else {
			panic(err)
		}
	} else if h.ObjectKind == "note" {
		fmt.Println("note")
		err := CommentLGTM(h)
		if err != nil {
			panic(err)
		}
	}

	/*
		read, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(read))
	*/
}

func InitDb() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbhost, dbport, dbuser, dbpassword, dbname)
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
		fmt.Println("Successfully connected to the Gitlab database")
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

func CheckInitial(h hook) (int, error) {
	fmt.Println("Check Initial")
	row := Db.QueryRow(`select count(n.id) from merge_requests as m, notes as n where m.iid = n.noteable_id and m.iid = $1 and target_project_id = $2`, h.ObjectAttributes.Iid, h.ObjectAttributes.TargetProjectId)
	var count int
	err := row.Scan(&count)
	if err != nil {
		fmt.Println("Error in select:", err)
		return 0, err
	} else {
		return count, nil
	}
}

func InitialComment(h hook) {
	message := `Total number of unique LGTMs need to be 2. After that this request will be Merged!`
	Post(message, h)
}

func Post(message string, h hook) {

	//	var mes = []byte(message)
	fmt.Println("iid:", h.ObjectAttributes.Iid, "targetprojectid:", h.ObjectAttributes.TargetProjectId)
	//	fmt.Println(string(mes))
	form := url.Values{}
	form.Add("body", message)
	client := &http.Client{}
	var u string
	if h.ObjectKind == "merge_request" {
		u = gitlabBase + "/api/v3/projects/" + strconv.Itoa(h.ObjectAttributes.TargetProjectId) + "/merge_requests/" + strconv.Itoa(h.ObjectAttributes.Iid) + "/notes"
	} else if h.ObjectKind == "note" {
		u = gitlabBase + "/api/v3/projects/" + strconv.Itoa(h.ProjectId) + "/merge_requests/" + strconv.Itoa(h.MergeRequest.Iid) + "/notes"
	} else {
		u = ""
	}
	r, err := http.NewRequest("POST", u, bytes.NewBufferString(form.Encode()))
	r.Header.Set("PRIVATE-TOKEN", gitlabToken)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func CheckLGTM(h hook) (int, error) {
	//Find last push note
	var iiid int
	if h.ObjectKind == "merge_request" {
		iiid = h.ObjectAttributes.Iid
	} else if h.ObjectKind == "note" {
		iiid = h.MergeRequest.Iid
	}

	row := Db.QueryRow(`select id FROM notes where noteable_id = $1 and noteable_type = 'MergeRequest' and system = 't' and note like 'Added % commit%' order by id desc limit 1`, iiid)
	var iid int
	err := row.Scan(&iid)
	if err != nil && err.Error() == "sql: no rows in result set" {
		iid = 0
	} else if err != nil {
		fmt.Println("Error in select:", err)
		return 0, err
	}
	var lgtms int
	//get number of LGTMs
	row1 := Db.QueryRow(`select count(distinct u.username) from notes as n, users as u, merge_requests as m where n.noteable_id = $1 and u.id = n.author_id and n.noteable_type = 'MergeRequest' and u.username != 'GitlabBot' and u.id != m.author_id and m.id = $2 and n.id > $3 and n.system = 'f' and note LIKE '%LGTM%'`, iiid, iiid, iid)
	err = row1.Scan(&lgtms)
	if err != nil && err.Error() == "sql: no rows in result set" {
		fmt.Println("im here")
		return 0, nil

	} else if err != nil {
		fmt.Println("Error in select:", err)
		return 0, err
	} else {
		return lgtms, nil
	}
}

func CommentLGTM(h hook) error {
	fmt.Println("Comment LGTM")
	fmt.Println(h.MergeRequest.Iid)
	row := Db.QueryRow(`SELECT n.note FROM notes AS n, users AS u WHERE n.noteable_id = $1 AND u.id = n.author_id AND n.noteable_type = 'MergeRequest' AND u.username = 'GitlabBot' AND n.system = 'f' ORDER BY n.id DESC LIMIT 1`, h.MergeRequest.Iid)
	var lastComment string
	err := row.Scan(&lastComment)
	if err != nil && err.Error() == "sql: no rows in result set" {
		fmt.Println(err)
		lastComment = "I have no comments here"
	} else if err != nil {
		return err
	}
	var newComment string
	lgtms, err := CheckLGTM(h)
	fmt.Println("lgtms:", lgtms)
	if lgtms < lgtmTreashold {
		newComment = "Current number of LGTMs: " + strconv.Itoa(lgtms) + " Number of LGTMs required: " + strconv.Itoa(lgtmTreashold-lgtms)
	} else {
		mergable, err := CheckMergable(h)
		if err == nil && mergable == "can_be_merged" {
			newComment = "Merged this request!"
			Put(h)
		} else if err == nil && mergable == "cannot_be_merged" {
			newComment = "This merge request requires manual conflict resolution."
		} else {
			fmt.Println("mergable error:", err)
		}
	}
	if newComment != lastComment {
		fmt.Println("Commenting:", newComment)
		fmt.Println("last Comment:", lastComment)
		Post(newComment, h)
		return nil
	} else {
		fmt.Println("Last Comment == New Comment, Standing down.")
		return nil

	}

}

func CheckMergable(h hook) (string, error) {
	fmt.Println("Check Mergable")
	row := Db.QueryRow(`SELECT m.merge_status FROM merge_requests AS m WHERE m.id = $1`, h.MergeRequest.Iid)
	var mergeStatus string
	err := row.Scan(&mergeStatus)
	if err != nil {
		return "", err
	} else {
		return mergeStatus, nil
	}
}

func Put(h hook) {
	fmt.Println("Put")
	client := &http.Client{}
	u := gitlabBase + "/api/v3/projects/" + strconv.Itoa(h.ProjectId) + "/merge_requests/" + strconv.Itoa(h.MergeRequest.Iid) + "/merge"
	r, err := http.NewRequest("PUT", u, nil)
	r.Header.Set("PRIVATE-TOKEN", gitlabToken)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
