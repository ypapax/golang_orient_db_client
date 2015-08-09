package main

import (
	// "github.com/howeyc/gopass" // commented out to get rid of this dependancy
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

/*
starting point of the program
*/
func main() {
	/*
		set default host and port
	*/
	hostPort := defaultHostPort
	p("Welcome to OrientDb client for executing SQL queries")
	p("Connecting to " + hostPort)
	/*
		check if it's possible to connect to the default host
	*/
	databasesAreListedSuccessfully, databases := listDatabases(hostPort)
	if databasesAreListedSuccessfully {
		afterListingDatabases(hostPort, databases)
	} else {
		/*
			if default orientDb host (localhost:2580) is not available
			let user to input host and port
		*/
		p("Cannot connect to default host: " + hostPort)
		allUserInteractionFromAskingHostPort()
	}

}

/*
all program work, except connectiong to default host:port (localhost:2480)
*/
func allUserInteractionFromAskingHostPort() {
	var (
		databasesAreListedSuccessfully bool
		databases                      []string
		hostPort                       string
	)
	for {
		hostPort = readFromConsole("Please input host:port")
		p("You have inputted " + hostPort)
		databasesAreListedSuccessfully, databases = listDatabases(hostPort)
		/*
			if databases are listed successfully move further to choosing database
		*/
		if databasesAreListedSuccessfully {
			break
		}
	}
	afterListingDatabases(hostPort, databases)

}

/*
asks a user to choose db by a number and then username password, then input query or path to query to execute
*/
func afterListingDatabases(hostPort string, databases []string) {
	/*
		list databases with numbers and let user check one by a number
	*/
	dbChosedSuccessfully, choosedDbName := printDatabasesWithNumbersAndReturnChoosedDatabase(databases)

	/*
		after database is choosed
	*/
	if dbChosedSuccessfully {
		askUsernameAndPasswordUntilSuccess(hostPort, choosedDbName)
	}
}

/*
cyclic action of getting username and password from a user until success
*/
func askUsernameAndPasswordUntilSuccess(hostPort, choosedDbName string) {
	var (
		user, pass              string // credentials variables
		userAndPasswordAreValid bool
	)
	for {
		var userAndPassWordAreInputted bool
		/*
			let a user input username and password
		*/
		userAndPassWordAreInputted, user, pass = askForUsernamePassword(choosedDbName)
		if userAndPassWordAreInputted {
			/*
				check username and password by connecting to orientDb database
			*/
			userAndPasswordAreValid = connectToCheckUserAndPass(hostPort, choosedDbName, user, pass)
			if userAndPasswordAreValid {
				break
			}

		}
	}
	if userAndPasswordAreValid {
		askToInputQueryAndExecuteIt(hostPort, choosedDbName, user, pass)
	}
}

const (
	defaultHostPort = "localhost:2480"
)

/*
GET HTTP query to host:port "hostPort", using command "command",
authintificating with username "user"
and password "pass"

result:
ok is true in case the query do not throw an error
ok is false in case the query throws an error
queryResult is json string with database response
*/
func query(hostPort, command, user, pass string) (ok bool, queryResult string) {
	client := &http.Client{}
	url := fmt.Sprintf("http://%s/%s", hostPort, command)
	fmt.Println("url")
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(user, pass)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("\nError : %s\n", err)
		return
	} else {
		var b bytes.Buffer
		_, err = b.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal("Error : %s", err)

		}
		queryResult = b.String()
		ok = !notValidUserPass(queryResult)
		return
	}
}

/*
Currently not used, but maybe helpful in future.
bash "curl" command for executing POST request to orientDb REST API
*/
func curlRequest(hostPort, dbName, sqlQuery, userName, password string) (ok bool, response string) {
	url := fmt.Sprintf(`http://%s/command/%s/sql`, hostPort, dbName)
	/*
		creating curl command
	*/
	curlCommand := fmt.Sprintf(`curl -X POST -u %s:%s -H "Content-Type: application/json" -d "%s"  %s`, userName, password, sqlQuery, url)
	fmt.Println("curlCommand")
	fmt.Println(curlCommand)
	/*
		executing curl command using sh
	*/
	output, e := exec.Command("sh", "-c", curlCommand).Output()
	// println(string(output))
	if e != nil {
		fmt.Println("error")
		fmt.Println(e)
		ok = true
	}
	response = string(output)
	return
}

/*
native GoLang POST request to orientDb server
*/
func Post(hostPort, dbName, sqlQuery, userName, password string) (ok bool, response string) {
	/*
		url for POST request
	*/
	url := fmt.Sprintf(`http://%s/command/%s/sql`, hostPort, dbName)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(sqlQuery))
	req.SetBasicAuth(userName, password)
	/*
		Accept-Encoding header is necesary for orientDb REST API
	*/
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Content-Type", "Content-Type: text/plain")
	/*
		request executing
	*/
	resp, err := client.Do(req)

	if err != nil {

		fmt.Printf("Error : %s", err)
	} else {
		ok = true
	}

	/*
		reading response
	*/
	if resp != nil {
		var b bytes.Buffer
		_, err = b.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal("Error : %s", err)
			ok = false
		} else {
			ok = true
		}
		/*
			convert response to string
		*/
		response = b.String()
	}
	return
}

/*
get list of databases names of certain orientDb host
*/
func listDatabases(hostPort string) (ok bool, databases []string) {
	var queryResult string

	ok, queryResult = query(hostPort, "listDatabases", "", "")
	var databasesListResult databasesListResultType
	fromJson(queryResult, &databasesListResult)
	databases = databasesListResult.Databases
	return
}

/*
database list type for parsing json and getting list of databases names of a certain server
*/
type databasesListResultType struct {
	Databases []string `json:"databases"`
}

/*
converts json string to GoLang structure
*/
func fromJson(jsonSrc string, objRef interface{}) error {
	return json.Unmarshal([]byte(jsonSrc), &objRef)
}

/*
let user to choose database name from console by a number
*/
func printDatabasesWithNumbersAndReturnChoosedDatabase(databases []string) (ok bool, choosedDb string) {
	if len(databases) == 0 {
		fmt.Println("There are no databases")
	} else {

		for i, db := range databases {
			line := fmt.Sprintf("%d. %s", i, db)
			fmt.Println(line)
		}
		var dbNumber int
		for {
			dbNumberStr := readFromConsole("Please choose a database by number")
			dbNumber = strToInt(dbNumberStr)
			dbNumberIsOk := dbNumber >= 0 && dbNumber < len(databases)
			if dbNumberIsOk {
				choosedDb = databases[dbNumber]
				ok = true
				break
			} else {
				p(fmt.Sprintf("Please input database number between 0 and %d", len(databases)-1))
			}

		}
	}
	return
}

/*
convert string to integer
*/
func strToInt(s string) (i int) {
	var err error
	i, err = strconv.Atoi(s)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	return
}

/*
response in case of not valid username, password pair
*/
const notValidUserPassResponse = "401 Unauthorized."

/*
check if response signals about not valid username and password
*/
func notValidUserPass(response string) bool {
	return response == notValidUserPassResponse
}

/*
let user input to console username and password for connecting to orient db database
*/
func askForUsernamePassword(dbName string) (bool, string, string) {
	var (
		username, pass string
		ok             bool
	)
	username = readFromConsole(fmt.Sprintf(`Please enter username for "%s" database`, dbName))
	pass = readFromConsole(fmt.Sprintf(`Please enter password for "%s" user`, username))
	/*
		Masked password is commented out in order to remove dependancy to gopass
	*/
	// p(fmt.Sprintf(`Please enter password for "%s" user`, username))
	// pass = string(gopass.GetPasswdMasked()) // Masked
	ok = len(username) > 0 && len(pass) > 0
	return ok, username, pass

}

/*
converts query like "Selct * from Employee"
to a part of HTTP request like "query/Employee/sql/Selct * from Employee"
*/
func executeSqlQuery(sqlQuery, hostPort, dbName, user, pass string) (ok bool, result string) {
	ok, result = Post(hostPort, dbName, sqlQuery, user, pass)

	if ok {
		if notValidUserPass(result) {

			ok = false
		}
	}
	return
}

/*
let user input sql query to console, execute it and return result
*/
func askToInputQueryAndExecuteIt(hostPort, dbName, user, pass string) {
	queryValue := readFromConsole(fmt.Sprintf(`Please input not very large query (less than 1000 characters) or /absolute/path/to/file with a query of any size for "%s" database on %s`, dbName, hostPort))
	if len(queryValue) > 0 {
		ok, result := executeRequestByFilePathOrQuery(queryValue, hostPort, dbName, user, pass)
		if !ok {
			/*
				in case of problem with authentification
			*/
			fmt.Println()
			fmt.Println("Problem with db connection occured, trying to reconnect with the same username and password")
			/*
				try to reconnect with the same username and password
			*/
			ok = connectToCheckUserAndPass(hostPort, dbName, user, pass)
			if !ok {
				/*
					or even start from chosing host and port
				*/
				allUserInteractionFromAskingHostPort()
			}
		}
		fmt.Println("result")
		p(result)
	}
	time.Sleep(100 * time.Millisecond)
	/*
		request is done, ask to input new one:
	*/
	askToInputQueryAndExecuteIt(hostPort, dbName, user, pass)
}

/*
remove head and tail spaces of a string
*/
func trim(s string) string {
	return strings.TrimSpace(s)
}

/*
let a user input line to console and read it to string variable; and return it.
*/
func readFromConsole(msg string) (result string) {
	p(msg)
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	result = trim(line)
	return
}

/*
print to console with new line
*/
func p(str string) {
	fmt.Println(str)
}

/*
connects to orientDb database just in order to check that username and password are valid
*/
func connectToCheckUserAndPass(hostPort, dbName, user, pass string) (okConnected bool) {
	command := "connect/" + dbName
	ok, _ := query(hostPort, command, user, pass)

	if !ok {
		p(fmt.Sprintf("Username and password are not valid for db %s on %s", dbName, hostPort))
	}
	return ok
}

/*
get file content by its path
*/
func fileToStr(path string) string {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}
	return string(bytes)
}

/*
for big queries, that are longer than 1000 characters
*/
func executeQueryByFilePath(filePath, hostPort, dbName, user, pass string) (ok bool, result string) {
	/*
		reading query from file
	*/
	p(fmt.Sprintf("Reading query from file %s", filePath))
	query := fileToStr(filePath)
	p(query)
	/*
		execute query
	*/
	ok, result = executeSqlQuery(query, hostPort, dbName, user, pass)
	return
}

/*
returns true if inputtedQueryOrFilepath is absolute path to the file
*/
func detectFileName(inputtedQueryOrFilepath string) (yesItIsFileName bool) {
	yesItIsFileName = strings.HasPrefix(inputtedQueryOrFilepath, "/")
	return
}

/*
executing query by text or filepath with query text
*/
func executeRequestByFilePathOrQuery(inputtedQueryOrFilepath, hostPort, dbName, user, pass string) (ok bool, result string) {
	thisIsFileName := detectFileName(inputtedQueryOrFilepath)
	if thisIsFileName {
		ok, result = executeQueryByFilePath(inputtedQueryOrFilepath, hostPort, dbName, user, pass)
	} else {
		ok, result = executeSqlQuery(inputtedQueryOrFilepath, hostPort, dbName, user, pass)
	}
	return
}
