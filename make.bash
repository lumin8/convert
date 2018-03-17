#!/bin/bash -

#c/o Ryan Chapman 2017

TRUE=0
FALSE=1

function _run
{
    if [[ $1 == fatal ]]; then
        errors_fatal=$TRUE
    else
        errors_fatal=$FALSE
    fi
    shift
    logit "${BOLD}$*${CLR}"
    eval "$*"
    rc=$?
    if [[ $rc != 0 ]]; then
        msg="${BOLD}${RED}$*${CLR}${RED} returned $rc${CLR}"
        if [[ $errors_fatal == $FALSE ]]; then
            msg+=" (error ignored)"
        fi
    else
        msg="${BOLD}${GREEN}$*${CLR}${GREEN} returned $rc${CLR}"
    fi
    logit "${BOLD}$msg${CLR}"
    # fail hard and fast
    if [[ $rc != 0 && $errors_fatal == $TRUE ]]; then
        pwd
        exit 1
    fi
    return $rc
}

function logit
{
    if [[ "${1}" == "FATAL" ]]; then
        fatal="FATAL"
        shift
    fi
    echo -n "$(date '+%b %d %H:%M:%S.%N %Z') $(basename -- $0)[$$]: "
    if [[ "${fatal}" == "FATAL" ]]; then echo -n "${RED}${fatal} "; fi
    echo "$*"
    if [[ "${fatal}" == "FATAL" ]]; then echo -n "${CLR}"; exit 1; fi
}

function run
{
    _run fatal $*
}

function run_ignerr
{
    _run warn $*
}

function make_version ()
{
    local timestamp=`date +%s`
    local builduser=`id -un`
    local buildhost=`hostname`
        local gitshortsha=`git rev-parse --short HEAD`
cat <<vEOF >version.go
package main

const BUILDTIMESTAMP = $timestamp
const BUILDUSER      = "$builduser"
const BUILDHOST      = "$buildhost"
const BUILDGITSHA    = "$gitshortsha"
vEOF
    logit "Wrote version.go: timestamp=$timestamp; builduser=$builduser; buildhost=$buildhost"
}

# NOTE: cross compiling does not work wuth gokogiri, unless you install a GNU/Linux cross compiler
# https://github.com/golang/go/issues/12888
function build ()
{
    #local os=${1}
    local os=$(uname -s)
    #local arch=${2}
    local arch=$(uname -m)
    local file_ext=""

    #export GOOS=${os}
    #export GOARCH=${arch}

    # workaround for error: "[%!v(PANIC=runtime error: cgo argument has Go pointer to Go pointer)]"
    export GODEBUG=cgocheck=0

    ## our main target is linux. if not building for that OS, append ${os} to build artifact name
    #if [[ "${os}" != "linux" ]]; then
    #    # special case, win needs a .exe extension
    #    if [[ "${os}" == "windows" ]]; then 
    #        file_ext=".exe"
    #    else
    #        file_ext=".${os}"
    #    fi
    #fi

    logit "Building for ${os}:${arch}"
    run_ignerr "go get"
    run "go build -o dataman${file_ext} *.go"
    local rc=$?
    logit "Building for ${os}:${arch}: done"

    return ${rc}
}

function main ()
{
    go get
    make_version
#    build windows amd64
#    build linux amd64
#    build darwin amd64
    build
}

main
