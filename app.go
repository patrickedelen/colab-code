package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris"
	"github.com/subosito/twilio"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
)

// handler for mms
// func receiveImage(&msg) {

// }

// func receiveText(&msg) {

// }

// // text everyone who is signed up to authenticate requests for the code
// // if a request is pending, text the requester that their request timed out
// func dispatchRequest() {
// 	rows, err := db.Query("SELECT * FROM test WHERE pending=false")

// }

// func createServer(&client) {

// }

// Member represents someone who has the authority to give out the co-lab code
type Member struct {
	name    string
	number  string
	pending bool
}

func checkMember(num string, db *sql.DB) (bool, string, bool) {
	rows, err := db.Query("SELECT * FROM members WHERE number=$1;", num)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	mems := make([]Member, 0) // create slice of type member
	for rows.Next() {
		var name, number string
		var pending bool
		if err := rows.Scan(&name, &number, &pending); err != nil {
			fmt.Println(err)
		}

		mems = append(mems, Member{name, number, pending})
	}

	if rows.Err() != nil {
		fmt.Println(err)
	}
	if len(mems) != 0 {
		return true, mems[0].name, mems[0].pending
	}

	return false, "not found in members", false

	// fmt.Println(mems)
}

// Request is the db structure of new requests for the co-lab code
type Request struct {
	tonumber   string
	fromnumber string
	atdate     time.Time
	active     bool
}

func checkRequests(num string, db *sql.DB) (string, bool) {
	rows, err := db.Query("SELECT * FROM requests WHERE tonumber=$1;", num)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	return "num", true

}

// send text to num with the code,
func approveRequest(num string, db *sql.DB) {
	fmt.Println("approving request")
}

// send text to num denying them entry, update request table to delete all requests
// then update pending status of members with this request
func denyRequest(num string, db *sql.DB) {
	fmt.Println("denying request")
}

// message all members with the new request for the code and reply to the text
func dispatchRequest(num string, pic string, body string, db *sql.DB) {

	// params := twilio.MessageParams{
	// 	Body:     "Reply yes to approve entry for this person",
	// 	MediaUrl: []string{p},
	// }
	// s, response, err := client.Messages.Send(os.Getenv("T_FROM"), ctx.FormValue("From"), params)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(s, response)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	connStr := os.Getenv("DB_STRING")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error connecting to db")
	}

	fmt.Println(os.Getenv("T_SECRET"))

	app := iris.New()
	app.Logger().SetLevel("debug")

	// Optionally, add two built'n handlers
	// that can recover from any http-relative panics
	// and log the requests to the terminal.
	// app.Use(recover.New())
	// app.Use(logger.New())

	client := twilio.NewClient(os.Getenv("T_SID"), os.Getenv("T_SECRET"), nil)

	// check if the message is from a member
	// then check if the message is from someone who has already made a request
	// then check if the message has a media attachment
	// if none of those pass, reply with an error
	app.Handle("POST", "/msg", func(ctx iris.Context) {
		d := ctx.FormValues()
		numMedia, err := strconv.ParseInt(ctx.FormValue("NumMedia"), 10, 8)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(d, numMedia)

		// check ctx.FormValue("Body")

		// this method checks if the sender is an approved member
		// check ctx.FormValue("From")
		found, name, pending := checkMember(ctx.FormValue("From"), db)
		fmt.Println(found, name)

		if found && pending {
			// search the db for the person who requested
			requestNumber, active := checkRequests(ctx.FormValue("From"), db)
			if active {
				if strings.ToLower(ctx.FormValue("Body")) == "yes" {
					approveRequest(requestNumber, db)
				} else {
					denyRequest(requestNumber, db)
				}
			}
		} else if found {
			// reply with error message
			fmt.Println("member replied to stale request")
			params := twilio.MessageParams{
				Body: "This request either never existed or has already been answered",
			}
			s, response, err := client.Messages.Send(os.Getenv("T_FROM"), ctx.FormValue("From"), params)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(s, response)
		} else if numMedia == 1 {
			m := ctx.FormValue("MediaUrl0")
			b := ctx.FormValue("Body")
			f := ctx.FormValue("From")
			dispatchRequest(f, m, b, db)
		} else {
			// send the catch-all error message
			params := twilio.MessageParams{
				Body: "Unknown request; try contacting someone directly for help",
			}
			s, response, err := client.Messages.Send(os.Getenv("T_FROM"), ctx.FormValue("From"), params)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(s, response)
		}

	})

	app.Run(iris.Addr(":8081"))

	// rows, err := db.Query("INSERT INTO test (name, number, added, pending) VALUES ('patrick', '+18323324935', $1, true);", t)

	// if err != nil {
	// 	fmt.Println(err)
	// 	log.Fatal("Error inserting")
	// } else {
	// 	fmt.Println(rows)
	// }
}
