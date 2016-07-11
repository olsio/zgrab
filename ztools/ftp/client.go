package ftp

import (
  "fmt"
  "net"
  "strconv"
  "strings"
)

type FTP struct {
  host    string
  port    int
  user    string
  passwd  string
  pasv    int
  cmd     string
  Code    string
  Message string
  Debug   bool
  stream  []byte
  conn    net.Conn
  Error   error
}

func (ftp *FTP) debugInfo(s string) {
  if ftp.Debug {
    fmt.Println(s)
  }
}

func (ftp *FTP) Login(user, passwd string) {
  ftp.Request("USER " + user)
  ftp.Request("PASS " + passwd)
  ftp.user = user
  ftp.passwd = passwd
}

func (ftp *FTP) Response() (code string, message string, err error) {
  ret := make([]byte, 1024)
  n, err := ftp.conn.Read(ret)
  fmt.Println("n:")
  fmt.Println(n)
  fmt.Println("err")
  fmt.Println(err)
  fmt.Println("ret:")
  fmt.Println(string(ret))
  msg := string(ret[:n])
  if (len(msg) < 3) {
    code = "999"
  } else {
    code = msg[:3]
  }

  return
}

func (ftp *FTP) Request(cmd string) {
  ftp.conn.Write([]byte(cmd + "\r\n"))
  ftp.cmd = cmd
  ftp.Code, ftp.Message, ftp.Error = ftp.Response()
  if cmd == "PASV" {
    start, end := strings.Index(ftp.Message, "("), strings.Index(ftp.Message, ")")
    s := strings.Split(ftp.Message[start:end], ",")
    l1, _ := strconv.Atoi(s[len(s)-2])
    l2, _ := strconv.Atoi(s[len(s)-1])
    ftp.pasv = l1*256 + l2
  }
  if (cmd != "PASV") && (ftp.pasv > 0) {
    ftp.Message = newRequest(ftp.host, ftp.pasv, ftp.stream)
    ftp.debugInfo("<*response*> " + ftp.Message)
    ftp.pasv = 0
    ftp.stream = nil
    ftp.Code, _, ftp.Error = ftp.Response()
  }
}

func (ftp *FTP) Pasv() {
  ftp.Request("PASV")
}

func (ftp *FTP) Pwd() {
  ftp.Request("PWD")
}

func (ftp *FTP) Cwd(path string) {
  ftp.Request("CWD " + path)
}

func (ftp *FTP) Mkd(path string) {
  ftp.Request("MKD " + path)
}

func (ftp *FTP) Size(path string) (size int) {
  ftp.Request("SIZE " + path)
  size, _ = strconv.Atoi(ftp.Message)
  return
}

func (ftp *FTP) List() {
  ftp.Pasv()
  ftp.Request("LIST")
}

func (ftp *FTP) Stor(file string, data []byte) {
  ftp.Pasv()
  if data != nil {
    ftp.stream = data
  }
  ftp.Request("STOR " + file)
}

func (ftp *FTP) Quit() {
  ftp.Request("QUIT")
  ftp.conn.Close()
}

// new connect to FTP pasv port, return data
func newRequest(host string, port int, b []byte) string {
  conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
  defer conn.Close()
  if b != nil {
    conn.Write(b)
    return "OK"
  }
  ret := make([]byte, 4096)
  n, _ := conn.Read(ret)
  return string(ret[:n])
}