/*
 * ZGrab Copyright 2015 Regents of the University of Michigan
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy
 * of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
 * implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package ftp

import (
 "net"
 "regexp"
 "strings"
 "fmt"
 "github.com/olsio/zgrab/ztools/util"
)

var ftpEndRegex = regexp.MustCompile(`^(?:.*\r?\n)*([0-9]{3})( [^\r\n]*)?\r?\n$`)

func GetFTPBanner(logStruct *FTPLog, connection net.Conn) (bool, error) {
	var err error
	var ftp *goftp.FTP

	if ftp, err = goftp.Connect(connection); err != nil {
		panic(err)
	}

	defer ftp.Close()

	if err = ftp.Login("anonymous", "me@earth.org"); err != nil {
		fmt.Println("error: login failure")
		return false, nil
	}

	if err = ftp.Cwd("/"); err != nil {
		fmt.Println("error: cwd")
		return false, nil
	}

	var curpath string
	if curpath, err = ftp.Pwd(); err != nil {
		fmt.Println("error: pwd")
		return false, nil
	}

	fmt.Printf("Current path: %s", curpath)

	var files []string
	if files, err = ftp.List(""); err != nil {
		fmt.Println("error: list")
		return false, nil
	}

	fmt.Println(files)

	return true, nil
}

func SetupFTPS(logStruct *FTPLog, connection net.Conn) (bool, error) {
	buffer := make([]byte, 1024)

	connection.Write([]byte("AUTH TLS\r\n"))
	respLen, err := util.ReadUntilRegex(connection, buffer, ftpEndRegex)
	if err != nil {
		return false, err
	}

	logStruct.AuthTLSResp = string(buffer[0:respLen])
	retCode := ftpEndRegex.FindStringSubmatch(logStruct.AuthTLSResp)[1]
	if strings.HasPrefix(retCode, "2") {
		return true, nil
	} else {
		connection.Write([]byte("AUTH SSL\r\n"))
		respLen, err := util.ReadUntilRegex(connection, buffer, ftpEndRegex)
		if err != nil {
			return false, err
		}

		logStruct.AuthSSLResp = string(buffer[0:respLen])
		retCode := ftpEndRegex.FindStringSubmatch(logStruct.AuthSSLResp)[1]

		if strings.HasPrefix(retCode, "2") {
			return true, nil
		}
	}

	return false, nil
}
