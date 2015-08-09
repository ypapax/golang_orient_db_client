# Golang orientDb client
It's a terminal `golang` application that connects to [OrientDB](http://orientdb.com/download/) database using the REST-interface.


You can run it: 

```go run /path/to/go_orient_db_client.go```

First, it tries to find the local address to database (standard `http://localhost:2480/`).
If not found, it asks for the host and port. 

After connection to OrientDb database, it
lists all databases with a number to each. A user can select a database by its number. Then the program asks to input  
 `username/password` for the selected database.

Now you can make queries to OrientDb like `select * from #9:826` or `INSERT INTO V SET name = 'jack', boss = #11:19`. 
If your query is longer than `1000` characters, you have to put it
in a file and input `/absolute/path/to/this/file` to the program to execute the query in OrientDb database.
