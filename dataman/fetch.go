package main

import (
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os/user"
    _ "github.com/lib/pg"
)

const (
    pgcreds = "/.pgpass"
    dbname = "layers"
    library = "catalog"
)


func fetchHandler(w http.ResponseWriter, r *http.Request) {

    //token, err := paramCheck("token", r)
    //if err != nil {
    //    token = "0"
        //w.Write(err)
        //r.Body.Close()
        //return
    //}

    //_, err := tokencheck(token)
    //if err ! = nil {
    //  w.Write(err)
    //  r.Body.Close()
    //  return
    //}

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

    resp, err := http.Get(url)
    io.Copy(resp,r.Body)
    r.Body.Close()
}

func fetchOrg(org string) string error {
    usr, _ := user.Current()
    pgpass := usr.HomeDir + pgcreds
    var data string

    if _, err := os.Stat(pgpass); err !=  nil {
      err = errors.New("Missing or misconfigured credentials pgpass specified in the host's home directory.")
      return err
    }

    dbinfo := fmt.Sprintf("dbname=%s sslmode=disable", dbname)
    db, err := sql.Open("postgres", dbinfo)
    if err != nil {
      err = errors.New("Could not establish a connection with the host")
      return _, err
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
      err = errors.New("Could not establish a connection with the dataset")
      return _, err
    }

    var subquery string

    if org is nil {
      subquery = fmt.Sprintf("select org,name,did,formats from %s",library)
    } else {
      subquery = fmt.Sprintf("select org,name,did,formats from %s where organization like '%s'",library,org)
    }

    query := fmt.Sprintf("select array_to_json(array_agg(row_to_json(t))) from ( %s ) t",subquery)

    err = db.QueryRow(query,1).Scan(&data)
    if err != nil {
      err = errors.New("No results found")
      return _, err
    }

    return data, _
}


func fetchDid(did int, format string) string error {
    usr, _ := user.Current()
    pgpass := usr.HomeDir + pgcreds
    var url string

    if _, err := os.Stat(pgpass); err !=  nil {
      err = errors.New("Missing or misconfigured credentials pgpass specified in the host's home directory.")
      return err
    }

    dbinfo := fmt.Sprintf("dbname=%s sslmode=disable", dbname)
    db, err := sql.Open("postgres", dbinfo)
    if err != nil {
      err = errors.New("Could not establish a connection with the host")
      return _, err
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
      err = errors.New("Could not establish a connection with the dataset")
      return _, err
    }

    // eg. mapurl, shpurl, csvurl, gsheeturl
    furl := format + "url"

    query := fmt.Sprintf("select %s from %s where did = '%s' limit 1",furl,library,did)
    if err != nil {
      err = errors.New("No results found")
      return _, err
    }

    err = db.QueryRow(query,1).Scan(&url)
    if err != nil {
      err = errors.New("No results found")
      return _, err
    }

    return url, _
}
