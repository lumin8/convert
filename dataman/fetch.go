package main

import (
        "cloud.google.com/go/storage"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
        "strings"
        "time"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

const (
	cloudsqlcreds = "/home/scott/cloudsqlcreds.yml"
        gskey       = "/home/scott/gskey.pem"
	dbname        = "layers"
	library       = "catalog"
)

type CloudSqlCreds struct {
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
        GSid     string `yaml:"googleaccessid"`
        GSkey    string `yaml:"privatekey"`
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
			log.Printf("%s", err.Error())
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

        //  TBD SCOTT:  proper url parsing, not the hack herein
	format, err := paramCheck("format", r)
	if err != nil || format == "" {
		w.Write([]byte("please specify a correct file format '&format=' "))
		r.Body.Close()
		return
	}

        format = strings.Replace(format, "%20", " ", -1)

	url, layername, cerr := fetchLocation(lid, format)
	if cerr != nil {
		w.Write([]byte(cerr.Error()))
		r.Body.Close()
		return
	}

        switch format {
        case "google sheet":
                log.Printf("%s", *url)
                http.Redirect(w,r,*url,http.StatusSeeOther)
                r.Body.Close()
                return
        case "csv":
                w.Header().Set("Content-Type","application/csv")
                w.Header().Add("Content-Disposition","attachment; filename=" + layername + ".csv")
                w.Header().Add("Pragma","no-cache")
        case "shapefile":
                w.Header().Set("Content-Type","application/zip")
                w.Header().Add("Content-Disposition","attachment; filename=" + layername + ".zip")
                w.Header().Add("Pragma","no-cache")
        case "wms url":
                log.Printf("%s", *url)
                http.Redirect(w,r,*url,http.StatusSeeOther)
                r.Body.Close()
        case "map":
                log.Printf("%s", *url)
                http.Redirect(w,r,*url,http.StatusSeeOther)
                r.Body.Close()
        }

        data, cerr := http.Get(*url)
        if cerr != nil {
                log.Printf("%s", cerr.Error())
                response := fmt.Sprintf("Could not fetch url: %s", url)
                w.Write([]byte(response))
                r.Body.Close()
                return
        }

        defer data.Body.Close()

	io.Copy(w, data.Body)
	r.Body.Close()
        return
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
	case "*":
                subquery = fmt.Sprintf("select organization,name,lid,formats from %s", library)
        default:
		subquery = fmt.Sprintf("select organization,name,lid,formats from %s where organization = '%s' order by name asc", library, org)
	}

	query := fmt.Sprintf("select array_to_json(array_agg(row_to_json(t))) from ( %s ) t", subquery)

	log.Printf(query)

	row := db.QueryRow(query)
	err = row.Scan(&data)
	if err != nil {
		log.Printf("%v", err.Error())
		err = errors.New("No results found")
		return "", err
	}

	return data, nil
}

func fetchLocation(lid string, format string) (*string, string, error) {
	var url string
        var layername string

	creds := getCloudCreds()
	dbinfo := fmt.Sprintf("dbname=%s sslmode=disable user=%s host=%s password=%s", dbname, creds.User, creds.Host, creds.Password)

	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		err = errors.New("Could not establish a connection with the host")
		return &url, "", err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		err = errors.New("Could not establish a connection with the database")
		return &url, "", err
	}

	// eg. mapurl, shpurl, csvurl, gsheeturl

	query := fmt.Sprintf("select name,\"%s\" from %s where lid = %s limit 1", format, library, lid)
	log.Printf("%s", query)

	err = db.QueryRow(query).Scan(&layername,&url)
	if err != nil {
		log.Printf("ola %s", err.Error())
		err = errors.New("No results found")
		return &url, "", err
	}

        if strings.Contains(url, "gs://") {
		err = getSignedURL(&url)
		if err != nil {
			return &url, "", err
		}
        }

	return &url, layername, nil
}

func getSignedURL(url *string) (error) {
        pkey, err := ioutil.ReadFile(gskey)
        if err != nil {
                return err
        }

        creds := getCloudCreds()

        *url = strings.Replace(*url, "gs://data.map.life/","", -1)

        log.Printf("fetching a signed url for %s",*url)

        gsurl, err := storage.SignedURL("data.map.life", *url, &storage.SignedURLOptions{
                GoogleAccessID:	creds.GSid,
                PrivateKey:	pkey,
                Method:		"GET",
                Expires:	time.Now().Add(15 * time.Minute),
        })

        *url = gsurl

        if err != nil {
                log.Printf(err.Error())
                return err
        }
        return nil
}
