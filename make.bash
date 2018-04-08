#!/bin/bash -
#
# Copyright (C) 2016 Ryan A. Chapman. All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
#
#   1. Redistributions of source code must retain the above copyright notice,
#      this list of conditions and the following disclaimer.
#
#   2. Redistributions in binary form must reproduce the above copyright notice, 
#      this list of conditions and the following disclaimer in the documentation
#      and/or other materials provided with the distribution.
#
# THE SOFTWARE IS PROVIDED "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
# INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
# FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE AUTHORS
# OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, 
# EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, 
# PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
# OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
# WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR 
# OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF 
# ADVISED OF THE POSSIBILITY OF SUCH DAMAGE. 

TRUE=0
FALSE=1

BOLD="$(tput bold)"
CLR="$(tput sgr0)"
RED="$(tput setaf 1 0)"
GREEN="$(tput setaf 10 0)"
CYAN="$(tput setaf 14 0)"

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

function build ()
{
    local os=${1}
    local arch=${2}
    local file_ext=""

    run "export GOOS=${os}"
    run "export GOARCH=${arch}"

    # our main target is linux. if not building for that OS, append ${os} to build artifact name
    if [[ "${os}" != "linux" ]]; then
        # special case, win needs a .exe extension
        if [[ "${os}" == "windows" ]]; then 
            file_ext=".exe"
        else
            file_ext=".${os}"
        fi
    fi

    logit "Building for ${os}:${arch}"
    run "go build -o dataman/*.go"
    local rc=$?
    logit "Building for ${os}:${arch}: done"

    return ${rc}
}

function main ()
{
    # cross compile for linux 64-bit target
    export GOOS=linux 
    export GOARCH=amd64 

    go get
    make_version
    build windows amd64
    build linux amd64
    build darwin amd64
}

main
