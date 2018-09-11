package main

import (
    "io/ioutil"
    "log"
    "net/http"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

const (
    user = ""
    password = ""
    dbname = "layers"
    library = "catalog"
)


func listHandler(w http.ResponseWriter, r *http.Request) {

    token, err := paramCheck("token", r)
    if err != nil {
        token = "0"
        //w.Write(err)
        //r.Body.Close()
        //return
    }

    _, err := tokencheck(token)
    if err ! = nil {
      w.Write(err)
      r.Body.Close()
      return
    }

    org, err := paramCheck("org", r)
    if err = nil {
        fetchOrg(org)
        r.Body.Close()
        return
    }

    did, err := paramCheck("did", r)
    if err != nil {
        w.Write([]byte("please specify a dataset id"))
        r.Body.Close()
        return
    }

    format, err := paramCheck("format", r)
    if err != nil {
        w.Write([]byte("please specify a correct file format"))
        r.Body.Close()
        return
    }

    url, err := fetchDid(did,format)
    if err != nil {
        w.Write(err)
        r.Body.Close()
        return
    }

    io.Copy(url,r.Body)
    r.Body.Close()
}

func fetchOrg(org string) string error {

    // data needed:  org name, dataset name, dataset id, formats available

    // username:password@protocol(address)/dbname?param=value

    db, err := sql.Open("mysql", user+":"+password+"/"+dbname)
    if err != nil {
      return _, err
    }

    defer db.Close()

    furl := "format" + "url"

    var stmt string

    if org is nil {
      stmt = ("select org,name,did,formats from %s",library
    } else {
      stmt = ("select org,name,did,formats from %s where organization like '%s'",library,org)
    }

    stmtSelect, err := db.Prepare(stmt)
    if err != nil {
      return _, err
    }

    defer stmtSelect.Close()

    var ymldata string // make this a map

    err = stmtSelect.QueryRow(1).Scan(&url) // scan the results into the map
    if err != nil {
      return _, err
    }

    return ymldata, err
}

func fetchDid(did int, format string) string error {
    db, err := sql.Open("mysql", user+":"+password+"/"+dbname)
    if err != nil {
      return _, err
    }

    defer db.Close()

    furl := "format" + "url"

    stmtSelect, err := db.Prepare("select %s from %s where did = '%s' limit 1",furl,library,did)
    if err != nil {
      return _, err
    }

    defer stmtSelect.Close()

    var url string

    err = stmtSelect.QueryRow(1).Scan(&url)
    if err != nil {
      return _, err
    }

    return url, err
}


