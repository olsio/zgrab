package ftpc

import (
  "fmt"
  "net"
  "strconv"
  "strings"
)

type FTPC struct {
  host    string
  port    int
  user    string
  passwd  string
  pasv    int
  cmd     string
  Code    int
  Message string
  Debug   bool
  stream  []byte
  conn    net.Conn
  Error   error
}

func (ftp *FTPC) debugInfo(s string) {
  if ftp.Debug {
    fmt.Println(s)
  }
}

func (ftp *FTPC) Connect(host string, port int) {
  addr := fmt.Sprintf("%s:%d", host, port)
  ftp.conn, ftp.Error = net.Dial("tcp", addr)
  ftp.Response()
  ftp.host = host
  ftp.port = port
}

func (ftp *FTPC) Login(user, passwd string) {
  ftp.Request("USER " + user)
  ftp.Request("PASS " + passwd)
  ftp.user = user
  ftp.passwd = passwd
}

func (ftp *FTPC) Response() (code int, message string) {
  ret := make([]byte, 1024)
  n, _ := ftp.conn.Read(ret)
  msg := string(ret[:n])
  code, _ = strconv.Atoi(msg[:3])
  message = msg[4 : len(msg)-2]
  ftp.debugInfo("<*cmd*> " + ftp.cmd)
  ftp.debugInfo(fmt.Sprintf("<*code*> %d", code))
  ftp.debugInfo("<*message*> " + message)
  return
}

func (ftp *FTPC) Request(cmd string) {
  ftp.conn.Write([]byte(cmd + "\r\n"))
  ftp.cmd = cmd
  ftp.Code, ftp.Message = ftp.Response()
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
    ftp.Code, _ = ftp.Response()
  }
}

func (ftp *FTPC) Pasv() {
  ftp.Request("PASV")
}

func (ftp *FTPC) Pwd() {
  ftp.Request("PWD")
}

func (ftp *FTPC) Cwd(path string) {
  ftp.Request("CWD " + path)
}

func (ftp *FTPC) Mkd(path string) {
  ftp.Request("MKD " + path)
}

func (ftp *FTPC) Size(path string) (size int) {
  ftp.Request("SIZE " + path)
  size, _ = strconv.Atoi(ftp.Message)
  return
}

func (ftp *FTPC) List() {
  ftp.Pasv()
  ftp.Request("LIST")
}

func (ftp *FTPC) Stor(file string, data []byte) {
  ftp.Pasv()
  if data != nil {
    ftp.stream = data
  }
  ftp.Request("STOR " + file)
}

func (ftp *FTPC) Quit() {
  ftp.Request("QUIT")
  ftp.conn.Close()
}

// new connect to FTPC pasv port, return data
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