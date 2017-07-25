package persistent

import (
    _ "github.com/go-sql-driver/mysql"
    "database/sql"
    "strconv"
);


func GetMysqlDB(userName string, password string,
             host string, port int, dbName string) *sql.DB {

    var connectStr string;

    connectStr = userName + ":" + password + 
                "@tcp(" + host + ":" + strconv.Itoa(port) + 
                ")/" + dbName + "?charset=utf8";

    db, err := sql.Open("mysql", connectStr)

    if err != nil {
        panic(err.Error())
    }

    defer db.Close()

    err = db.Ping()
    if err != nil {
        panic(err.Error()) 
    }

    return db

}


func GetWokerCurrentId(workerId int) int {

    return 1

}
