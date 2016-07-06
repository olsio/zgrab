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

 "github.com/olsio/zgrab/ztools/util"
)

const (
	READ_BUFFER_LENGTH = 8192
	MAX_READ_SIZE = 65536
)

var ftpEndRegex = regexp.MustCompile(`^(?:.*\r?\n)*([0-9]{3})( [^\r\n]*)?\r?\n$`)


func GetFTPBanner(logStruct *FTPLog, connection net.Conn) (bool, error) {
	buffer := make([]byte, 1024)
	respLen, err := util.ReadUntilRegex(connection, buffer, ftpEndRegex)
	logStruct.Banner = string(buffer[0:respLen])
	if err != nil {
		return false, err
	}

	connection.Write([]byte("anonymous\r\n"))
	respLen, err := util.ReadUntilRegex(connection, buffer, ftpEndRegex)
	logStruct.Login = string(buffer[0:respLen])
	if err != nil {
		return false, err
	}

	connection.Write([]byte("me@earth.org\r\n"))
	respLen, err := util.ReadUntilRegex(connection, buffer, ftpEndRegex)
	logStruct.Login = string(buffer[0:respLen])
	if err != nil {
		return false, err
	}

	connection.Write([]byte("dir\r\n"))
	respLen, err := ReadResponse(connection, buffer, ftpEndRegex)
	logStruct.Content
	if err != nil {
		return false, err
	}

	retCode := ftpEndRegex.FindStringSubmatch(logStruct.Banner)[1]

	return strings.HasPrefix(retCode, "2"), nil
}

func ReadResponse(conn net.Conn) (response string, err error) {
	var response []byte
	buffer := make([]byte, READ_BUFFER_LENGTH)

	numBytes := len(buffer)
	rounds := int(math.Ceil(float64(MAX_READ_SIZE) / READ_BUFFER_LENGTH))
	count := 0
	for numBytes != 0 && count < rounds && numBytes == READ_BUFFER_LENGTH {

		numBytes, err = conn.Read(buffer)

		if err, ok := err.(net.Error); ok && err.Timeout() {
			return nil, err
		}

		if err != nil && err != io.EOF {
			return nil, err
		}

		if count == rounds-1 {
			response = append(response, buffer[0:MAX_READ_SIZE%READ_BUFFER_LENGTH]...)
		} else {
			response = append(response, buffer[0:numBytes]...)
		}
		count += 1
	}

	return string(response), nil
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
