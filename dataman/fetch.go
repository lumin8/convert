package main

import (
    "database/sql"
    "errors"
    "fmt"
    "gopkg.in/yaml.v2"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    _ "github.com/lib/pq"
)

const (
    cloudsqlcreds = "/home/scott/cloudsqlcreds.yml"
    dbname = "layers"
    library = "catalog"
)

type CloudSqlCreds struct {
    Password	string  `yaml:"password"`
    Host	string  `yaml:"host"`
    User	string  `yaml:"user"`
}


func getCloudCreds() CloudSqlCreds {
    var creds CloudSqlCreds

    ymlcreds, err := ioutil.ReadFile(cloudsqlcreds)
    if err != nil {
      log.Printf("no cloud sql creds could be found, you'll need these for fetching data")
    }

    yaml.Unmarshal(ymlcreds, &creds)

    return creds
}


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
    if err == nil {
        data, err := fetchOrg(org)
        if err != nil {
          log.Printf("%s",err.Error())
          w.Write([]byte(err.Error()))
          return
        }
        w.Write([]byte(data))
        r.Body.Close()
        return
    }

    lid, err := paramCheck("lid", r)
    if err != nil || lid == "" {
        w.Write([]byte("please specify a dataset layer id '&lid='"))
        r.Body.Close()
        return
    }

    format, err := paramCheck("format", r)
    if err != nil || format == "" {
        w.Write([]byte("please specify a correct file format '&format=' "))
        r.Body.Close()
        return
    }

    url, cerr := fetchUrl(lid, format)
    if cerr != nil {
        w.Write([]byte(cerr.Error()))
        r.Body.Close()
        return
    }

    log.Println("about to fetch the data")

    data, cerr := http.Get(url)
    if cerr != nil {
        log.Printf("%s",cerr.Error())
        response := fmt.Sprintf("Could not fetch url: %s",url)
        w.Write([]byte(response))
        r.Body.Close()
        return
    }

    defer data.Body.Close()

    io.Copy(w,data.Body)
    r.Body.Close()
}


func fetchOrg(org string) (string, error) {
    var data string

    creds := getCloudCreds()
    dbinfo := fmt.Sprintf("dbname=%s sslmode=disable user=%s host=%s password=%s", dbname, creds.User, creds.Host, creds.Password)

    db, err := sql.Open("postgres", dbinfo)
    if err != nil {
      err = errors.New("Could not establish a connection with the host")
      return "", err
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
      err = errors.New("Could not establish a connection with the database")
      return "", err
    }

    var subquery string

    switch org {
      case org:
        subquery = fmt.Sprintf("select organization,name,lid,formats from %s where organization = '%s'",library,org)
      default:
        subquery = fmt.Sprintf("select organization,name,lid,formats from %s",library)
    }

    query := fmt.Sprintf("select array_to_json(array_agg(row_to_json(t))) from ( %s ) t",subquery)

    log.Printf(query)

    row := db.QueryRow(query)
    err = row.Scan(&data)
    if err != nil {
      log.Printf("%v",err.Error())
      err = errors.New("No results found")
      return "", err
    }

    return data, nil
}


func fetchUrl (lid string, format string) (string, error) {
    var url string

    creds := getCloudCreds()
    dbinfo := fmt.Sprintf("dbname=%s sslmode=disable user=%s host=%s password=%s", dbname, creds.User, creds.Host, creds.Password)

    db, err := sql.Open("postgres", dbinfo)
    if err != nil {
      err = errors.New("Could not establish a connection with the host")
      return "", err
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
      err = errors.New("Could not establish a connection with the database")
      return "", err
    }

    // eg. mapurl, shpurl, csvurl, gsheeturl
    furl := format + "url"

    query := fmt.Sprintf("select %s from %s where lid = %s limit 1",furl,library,lid)
    log.Printf("%s",query)

    err = db.QueryRow(query).Scan(&url)
    if err != nil {
      log.Printf("ola %s",err.Error())
      err = errors.New("No results found")
      return "", err
    }

    return url, nil
}
